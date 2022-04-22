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

package apiserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"syscall"
	"time"

	"github.com/onmetal/onmetal-api/envtestutils/internal/controlplane"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

type ProcessArgs map[string][]string

type APIServer struct {
	mainPath string
	command  []string

	config      *rest.Config
	etcdServers []string
	args        ProcessArgs
	mergeArgs   func(customArgs, defaultArgs ProcessArgs) ProcessArgs

	host    string
	port    int
	certDir string
	stdout  io.Writer
	stderr  io.Writer

	healthTimeout time.Duration
	waitTimeout   time.Duration
}

type Options struct {
	MainPath  string
	Command   []string
	MergeArgs func(customArgs, defaultArgs ProcessArgs) ProcessArgs

	ETCDServers []string
	Host        string
	Port        int
	CertDir     string
	Stdout      io.Writer
	Stderr      io.Writer

	HealthTimeout time.Duration
	WaitTimeout   time.Duration
}

func MergeArgs(customArgs, defaultArgs ProcessArgs) ProcessArgs {
	res := make(ProcessArgs)
	for key, value := range defaultArgs {
		res[key] = value
	}
	for key, value := range customArgs {
		res[key] = value
	}
	return res
}

func setAPIServerOptionsDefaults(opts *Options) {
	if opts.MergeArgs == nil {
		opts.MergeArgs = MergeArgs
	}
	if opts.Host == "" {
		opts.Host = "127.0.0.1"
	}
	if opts.Port == 0 {
		opts.Port = 8443
	}
	if opts.HealthTimeout == 0 {
		opts.HealthTimeout = 20 * time.Second
	}
	if opts.WaitTimeout == 0 {
		opts.WaitTimeout = 20 * time.Second
	}
}

func New(cfg *rest.Config, opts Options) (*APIServer, error) {
	if opts.MainPath == "" && len(opts.Command) == 0 {
		return nil, fmt.Errorf("must specify opts.MainPath or opts.Command")
	}
	if len(opts.ETCDServers) == 0 {
		return nil, fmt.Errorf("must specify opts.ETCDServers")
	}
	setAPIServerOptionsDefaults(&opts)

	return &APIServer{
		mainPath:      opts.MainPath,
		command:       opts.Command,
		config:        cfg,
		etcdServers:   opts.ETCDServers,
		args:          make(ProcessArgs),
		mergeArgs:     opts.MergeArgs,
		host:          opts.Host,
		port:          opts.Port,
		certDir:       opts.CertDir,
		stdout:        opts.Stdout,
		stderr:        opts.Stderr,
		healthTimeout: opts.HealthTimeout,
		waitTimeout:   opts.WaitTimeout,
	}, nil
}

func (a *APIServer) Start(ctx context.Context) error {
	tmp, err := a.setupTempDir()
	if err != nil {
		return fmt.Errorf("error setting up temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmp.Name()) }()

	cmd := a.createCmd(tmp)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting api server: %w", err)
	}

	var (
		waitDone = make(chan struct{})
		waitErr  error
	)
	go func() {
		defer close(waitDone)
		waitErr = cmd.Wait()
	}()

	var (
		healthDone              = make(chan struct{})
		healthErr               error
		healthCtx, healthCancel = context.WithTimeout(ctx, a.healthTimeout)
	)
	defer healthCancel()
	go func() {
		defer close(healthDone)
		healthErr = a.pollHealthCheck(healthCtx)
	}()

	select {
	case <-waitDone:
		healthCancel()
		if waitErr != nil {
			return fmt.Errorf("wait returned with error before healthy: %w", waitErr)
		}
		return fmt.Errorf("wait returned before ready")
	case <-healthDone:
		if healthErr != nil {
			return fmt.Errorf("healthiness check returned an error: %w", healthErr)
		}
	}

	select {
	case <-ctx.Done():
		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			return fmt.Errorf("error signaling process to stop: %w", err)
		}
	case <-waitDone:
		if waitErr != nil {
			return fmt.Errorf("process exited early with error: %w", waitErr)
		}
		return fmt.Errorf("process exited early with no error")
	}

	t := time.NewTimer(a.waitTimeout)
	defer t.Stop()

	select {
	case <-waitDone:
		return waitErr
	case <-t.C:
		return fmt.Errorf("timeout waiting for process to stop")
	}
}

func (a *APIServer) pollHealthCheck(ctx context.Context) error {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // skip verify for doing local health checks is ok.
			},
		},
	}

	return wait.PollImmediateInfiniteWithContext(ctx, 1*time.Second, func(ctx context.Context) (done bool, err error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s:%d/readyz", a.host, a.port), nil)
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

func (a *APIServer) setupTempDir() (*os.File, error) {
	tmpDir, err := os.MkdirTemp("", "apiserver")
	if err != nil {
		return nil, fmt.Errorf("error creating temp directory")
	}

	if a.mainPath != "" {
		apiSrvBinary, err := os.CreateTemp(tmpDir, "apiserver")
		if err != nil {
			_ = os.RemoveAll(tmpDir)
			return nil, fmt.Errorf("error creating api server binary file")
		}

		cmd := exec.Command("go", "build", "-o", apiSrvBinary.Name(), a.mainPath)
		cmd.Stdout = a.stdout
		cmd.Stderr = a.stderr
		if err := cmd.Run(); err != nil {
			_ = os.RemoveAll(tmpDir)
			return nil, fmt.Errorf("error building api server binary: %w", err)
		}

		a.command = []string{apiSrvBinary.Name()}
	}

	cfgData, err := controlplane.KubeConfigFromREST(a.config)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		return nil, err
	}

	tmp, err := os.CreateTemp(tmpDir, "kubeconfig")
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		return nil, err
	}

	if err := os.WriteFile(tmp.Name(), cfgData, 0666); err != nil {
		_ = os.RemoveAll(tmpDir)
		return nil, err
	}
	return tmp, nil
}

func (a *APIServer) createCmd(tmp *os.File) *exec.Cmd {
	defaultArgs := ProcessArgs{
		"etcd-servers":              a.etcdServers,
		"kubeconfig":                []string{tmp.Name()},
		"authentication-kubeconfig": []string{tmp.Name()},
		"authorization-kubeconfig":  []string{tmp.Name()},
		"bind-address":              []string{a.host},
		"secure-port":               []string{strconv.Itoa(a.port)},
		"feature-gates":             []string{"APIPriorityAndFairness=false"},
		"audit-log-path":            []string{"-"},
		"audit-log-maxage":          []string{"0"},
		"audit-log-maxbackup":       []string{"0"},
		"tls-cert-file":             []string{path.Join(a.certDir, "tls.crt")},
		"tls-private-key-file":      []string{path.Join(a.certDir, "tls.key")},
	}
	args := a.mergeArgs(a.args, defaultArgs)

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
	for key := range args {
		if !keySet.Has(key) {
			additionalKeys = append(additionalKeys, key)
		}
	}
	sort.Strings(additionalKeys)

	keys := append(defaultKeys, additionalKeys...)
	var execArgs []string
	if len(a.command) > 1 {
		execArgs = append(execArgs, a.command[1:]...)
	}
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
	cmd := exec.Command(a.command[0], execArgs...)
	cmd.Stdout = a.stdout
	cmd.Stderr = a.stderr
	return cmd
}
