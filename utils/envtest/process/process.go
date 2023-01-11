// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package process

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/onmetal/controller-utils/buildutils"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
)

type Args map[string][]string

func EmptyArgs() Args {
	return make(Args)
}

func (a Args) Set(key string, values ...string) Args {
	a[key] = values
	return a
}

func MergeArgs(customArgs, defaultArgs Args) Args {
	res := make(Args)
	for key, value := range defaultArgs {
		res[key] = value
	}
	for key, value := range customArgs {
		res[key] = value
	}
	return res
}

func BuildArgs(customArgs, defaultArgs Args, mergeArgs func(customArgs, defaultArgs Args) Args) []string {
	if mergeArgs == nil {
		mergeArgs = MergeArgs
	}
	args := mergeArgs(customArgs, defaultArgs)

	keySet := sets.NewString()
	var defaultKeys []string
	for key := range defaultArgs {
		if _, ok := args[key]; ok {
			keySet.Insert(key)
			defaultKeys = append(defaultKeys, key)
		}
	}
	sort.Strings(defaultKeys)

	var additionalKeys []string
	for key := range customArgs {
		if !keySet.Has(key) {
			additionalKeys = append(additionalKeys, key)
		}
	}
	sort.Strings(additionalKeys)

	keys := append(defaultKeys, additionalKeys...)
	var execArgs []string
	for _, key := range keys {
		values := args[key]
		switch len(values) {
		case 0:
			execArgs = append(execArgs, fmt.Sprintf("--%s", key))
		default:
			for _, val := range values {
				execArgs = append(execArgs, fmt.Sprintf("--%s=%s", key, val))
			}
		}
	}
	return execArgs
}

type Process struct {
	mainPackage  string
	buildOptions []buildutils.BuildOption
	command      []string
	cmd          *exec.Cmd

	args     []string
	argsFunc func(ctx InitContext) ([]string, error)

	stdout io.Writer
	stderr io.Writer

	healthTimeout time.Duration
	waitTimeout   time.Duration

	healthCheck HealthCheck

	waitDone chan struct{}
	errMu    sync.Mutex
	exitErr  error
	exited   bool
	dir      string
}

type InitContext struct {
	TempDir string
}

type Options struct {
	MainPath     string
	BuildOptions []buildutils.BuildOption
	Command      []string

	Args     []string
	ArgsFunc func(ctx InitContext) ([]string, error)

	AttachOutput bool
	Stdout       io.Writer
	Stderr       io.Writer

	HealthTimeout time.Duration
	WaitTimeout   time.Duration

	HealthCheck HealthCheck
}

func setOptionsDefaults(opts *Options) {
	if opts.HealthTimeout == 0 {
		opts.HealthTimeout = 20 * time.Second
	}
	if opts.WaitTimeout == 0 {
		opts.WaitTimeout = 20 * time.Second
	}
	if opts.AttachOutput {
		opts.Stdout = os.Stdout
		opts.Stderr = os.Stderr
	}
}

func New(opts Options) (*Process, error) {
	if opts.MainPath == "" && len(opts.Command) == 0 {
		return nil, fmt.Errorf("must specify opts.MainPath or opts.Command")
	}
	if opts.AttachOutput && (opts.Stdout != nil || opts.Stderr != nil) {
		return nil, fmt.Errorf("must not specify AttachOutput and Stdout / Stderr simultaneously")
	}
	if opts.Args != nil && opts.ArgsFunc != nil {
		return nil, fmt.Errorf("must not specify Args and ArgsFunc simultaneously")
	}
	setOptionsDefaults(&opts)

	return &Process{
		mainPackage:   opts.MainPath,
		buildOptions:  opts.BuildOptions,
		command:       opts.Command,
		args:          opts.Args,
		argsFunc:      opts.ArgsFunc,
		stdout:        opts.Stdout,
		stderr:        opts.Stderr,
		healthTimeout: opts.HealthTimeout,
		waitTimeout:   opts.WaitTimeout,
		healthCheck:   opts.HealthCheck,
	}, nil
}

func (a *Process) Exited() (bool, error) {
	a.errMu.Lock()
	defer a.errMu.Unlock()
	return a.exited, a.exitErr
}

func (a *Process) Start() error {
	var err error
	a.dir, err = a.setupTempDir()
	if err != nil {
		return fmt.Errorf("error setting up temp dir: %w", err)
	}

	a.cmd, err = a.createCmd()
	if err != nil {
		a.errMu.Lock()
		defer a.errMu.Unlock()
		a.exited = true
		return fmt.Errorf("error creating command: %w", err)
	}

	if err := a.cmd.Start(); err != nil {
		a.errMu.Lock()
		defer a.errMu.Unlock()
		a.exited = true
		return fmt.Errorf("error starting process: %w", err)
	}

	a.waitDone = make(chan struct{})
	go func() {
		defer close(a.waitDone)
		err := a.cmd.Wait()

		a.errMu.Lock()
		defer a.errMu.Unlock()
		a.exitErr = err
		a.exited = true
	}()

	if a.healthCheck == nil {
		return nil
	}

	var (
		healthDone              = make(chan struct{})
		healthErr               error
		healthCtx, healthCancel = context.WithTimeout(context.Background(), a.healthTimeout)
	)
	defer healthCancel()
	go func() {
		defer close(healthDone)
		healthErr = a.healthCheck.Check(healthCtx)
	}()

	select {
	case <-a.waitDone:
		healthCancel()
		_, exitErr := a.Exited()
		if exitErr != nil {
			return fmt.Errorf("wait returned with error before healthy: %w", exitErr)
		}
		return fmt.Errorf("wait returned before ready")
	case <-healthDone:
		if healthErr != nil {
			if a.cmd != nil {
				// intentionally ignore this -- we might've crashed, failed to start, etc
				_ = a.cmd.Process.Signal(syscall.SIGTERM)
			}
			return fmt.Errorf("healthiness check returned an error: %w", healthErr)
		}
		// This means we started successfully, health is done and returned no error.
		return nil
	}
}

func (a *Process) Stop() error {
	defer func() {
		if a.dir != "" {
			_ = os.RemoveAll(a.dir)
		}
	}()
	if a.cmd == nil {
		return nil
	}
	if done, _ := a.Exited(); done {
		return nil
	}
	if err := a.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("unable to signal for process to stop: %w", err)
	}

	t := time.NewTimer(a.waitTimeout)
	defer t.Stop()

	select {
	case <-a.waitDone:
		return nil
	case <-t.C:
		return fmt.Errorf("timeout waiting for process to stop")
	}
}

type HealthCheck interface {
	Check(ctx context.Context) error
}

type HealthCheckFunc func(ctx context.Context) error

func (f HealthCheckFunc) Check(ctx context.Context) error {
	return f(ctx)
}

func InsecurePollHealthCheck(address string) HealthCheckFunc {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // skip verify for doing local health checks is ok.
			},
		},
	}

	return func(ctx context.Context) error {
		return wait.PollImmediateInfiniteWithContext(ctx, 1*time.Second, func(ctx context.Context) (done bool, err error) {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
			if err != nil {
				return false, fmt.Errorf("error creating health request: %w", err)
			}

			res, err := httpClient.Do(req)
			if err != nil {
				return false, nil
			}

			_ = res.Body.Close()
			return res.StatusCode == http.StatusOK, nil
		})
	}
}

func (a *Process) setupTempDir() (string, error) {
	tmpDir, err := os.MkdirTemp("", "process")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory")
	}

	if a.mainPackage != "" {
		binary, err := os.CreateTemp(tmpDir, "process")
		if err != nil {
			_ = os.RemoveAll(tmpDir)
			return "", fmt.Errorf("error creating process binary file")
		}

		if err := buildutils.Build(a.mainPackage, binary.Name(), a.buildOptions...); err != nil {
			_ = os.RemoveAll(tmpDir)
			return "", fmt.Errorf("error building binary: %w", err)
		}

		a.command = []string{binary.Name()}
	}
	return tmpDir, nil
}

func (a *Process) createCmd() (*exec.Cmd, error) {
	var execArgs []string
	if len(a.command) > 1 {
		execArgs = append(execArgs, a.command[1:]...)
	}

	var args []string
	switch {
	case a.args != nil:
		args = a.args
	case a.argsFunc != nil:
		a, err := a.argsFunc(InitContext{TempDir: a.dir})
		if err != nil {
			return nil, fmt.Errorf("error getting args: %w", err)
		}
		args = a
	}

	execArgs = append(execArgs, args...)

	cmd := exec.Command(a.command[0], execArgs...)
	cmd.Stdout = a.stdout
	cmd.Stderr = a.stderr
	return cmd, nil
}
