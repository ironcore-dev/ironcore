// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

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
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/ironcore-dev/controller-utils/buildutils"
	"github.com/ironcore-dev/ironcore/utils/envtest/internal/testing/controlplane"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

type ProcessArgs map[string][]string

func EmptyProcessArgs() ProcessArgs {
	return make(ProcessArgs)
}

func (p ProcessArgs) Set(key string, values ...string) ProcessArgs {
	p[key] = values
	return p
}

type APIServer struct {
	mainPackage  string
	buildOptions []buildutils.BuildOption
	command      []string
	cmd          *exec.Cmd

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

	waitDone chan struct{}
	errMu    sync.Mutex
	exitErr  error
	exited   bool
	dir      string
}

type Options struct {
	MainPath     string
	BuildOptions []buildutils.BuildOption
	Command      []string
	Args         ProcessArgs
	MergeArgs    func(customArgs, defaultArgs ProcessArgs) ProcessArgs

	ETCDServers []string
	Host        string
	Port        int
	CertDir     string

	AttachOutput bool
	Stdout       io.Writer
	Stderr       io.Writer

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
	if opts.AttachOutput {
		opts.Stdout = os.Stdout
		opts.Stderr = os.Stderr
	}
	if opts.Args == nil {
		opts.Args = make(ProcessArgs)
	}
}

func New(cfg *rest.Config, opts Options) (*APIServer, error) {
	if opts.MainPath == "" && len(opts.Command) == 0 {
		return nil, fmt.Errorf("must specify opts.MainPath or opts.Command")
	}
	if opts.AttachOutput && (opts.Stdout != nil || opts.Stderr != nil) {
		return nil, fmt.Errorf("must not specify AttachOutput and Stdout / Stderr simultaneously")
	}
	if len(opts.ETCDServers) == 0 {
		return nil, fmt.Errorf("must specify opts.ETCDServers")
	}
	setAPIServerOptionsDefaults(&opts)

	return &APIServer{
		mainPackage:   opts.MainPath,
		buildOptions:  opts.BuildOptions,
		command:       opts.Command,
		config:        cfg,
		etcdServers:   opts.ETCDServers,
		args:          opts.Args,
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

func (a *APIServer) Exited() (bool, error) {
	a.errMu.Lock()
	defer a.errMu.Unlock()
	return a.exited, a.exitErr
}

func (a *APIServer) Start() error {
	var err error
	a.dir, err = a.setupTempDir()
	if err != nil {
		return fmt.Errorf("error setting up temp dir: %w", err)
	}

	a.cmd = a.createCmd()
	if err := a.cmd.Start(); err != nil {
		a.errMu.Lock()
		defer a.errMu.Unlock()
		a.exited = true
		return fmt.Errorf("error starting api server: %w", err)
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

	var (
		healthDone              = make(chan struct{})
		healthErr               error
		healthCtx, healthCancel = context.WithTimeout(context.Background(), a.healthTimeout)
	)
	defer healthCancel()
	go func() {
		defer close(healthDone)
		healthErr = a.pollHealthCheck(healthCtx)
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

func (a *APIServer) Stop() error {
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

func (a *APIServer) pollHealthCheck(ctx context.Context) error {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // skip verify for doing local health checks is ok.
			},
		},
	}

	return wait.PollUntilContextCancel(ctx, 1*time.Second, true, func(ctx context.Context) (done bool, err error) {
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

func (a *APIServer) setupTempDir() (string, error) {
	tmpDir, err := os.MkdirTemp("", "apiserver")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory")
	}

	if a.mainPackage != "" {
		apiSrvBinary, err := os.CreateTemp(tmpDir, "apiserver")
		if err != nil {
			_ = os.RemoveAll(tmpDir)
			return "", fmt.Errorf("error creating api server binary file")
		}

		if err := buildutils.Build(a.mainPackage, apiSrvBinary.Name(), a.buildOptions...); err != nil {
			_ = os.RemoveAll(tmpDir)
			return "", fmt.Errorf("error building api server binary: %w", err)
		}

		a.command = []string{apiSrvBinary.Name()}
	}

	cfgData, err := controlplane.KubeConfigFromREST(a.config)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", err
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "kubeconfig"), cfgData, 0666); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", err
	}
	return tmpDir, nil
}

func (a *APIServer) createCmd() *exec.Cmd {
	kubeconfig := filepath.Join(a.dir, "kubeconfig")
	defaultArgs := ProcessArgs{
		"etcd-servers":                 a.etcdServers,
		"kubeconfig":                   []string{kubeconfig},
		"authentication-kubeconfig":    []string{kubeconfig},
		"authorization-kubeconfig":     []string{kubeconfig},
		"bind-address":                 []string{a.host},
		"secure-port":                  []string{strconv.Itoa(a.port)},
		"enable-priority-and-fairness": []string{"false"},
		"audit-log-path":               []string{"-"},
		"audit-log-maxage":             []string{"0"},
		"audit-log-maxbackup":          []string{"0"},
		"tls-cert-file":                []string{path.Join(a.certDir, "tls.crt")},
		"tls-private-key-file":         []string{path.Join(a.certDir, "tls.key")},
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
