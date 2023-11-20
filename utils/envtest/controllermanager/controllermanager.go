// Copyright 2022 IronCore authors
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

package controllermanager

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ironcore-dev/controller-utils/buildutils"
	"github.com/ironcore-dev/ironcore/utils/envtest/internal/testing/controlplane"
	"github.com/ironcore-dev/ironcore/utils/envtest/process"
	"k8s.io/client-go/rest"
)

type ControllerManager struct {
	process *process.Process
}

type Options struct {
	MainPath     string
	BuildOptions []buildutils.BuildOption
	Command      []string
	Args         process.Args
	MergeArgs    func(customArgs, defaultArgs process.Args) process.Args

	Host string
	Port int

	AttachOutput bool
	Stdout       io.Writer
	Stderr       io.Writer

	HealthTimeout time.Duration
	WaitTimeout   time.Duration
}

func setOptionsDefaults(opts *Options) {
	if opts.MergeArgs == nil {
		opts.MergeArgs = process.MergeArgs
	}
	if opts.Host == "" {
		opts.Host = "127.0.0.1"
	}
	if opts.Port == 0 {
		opts.Port = 8081
	}
	if opts.Args == nil {
		opts.Args = process.EmptyArgs()
	}
}

func New(cfg *rest.Config, opts Options) (*ControllerManager, error) {
	setOptionsDefaults(&opts)

	healthProbeBindAddress := fmt.Sprintf("http://%s:%d/healthz", opts.Host, opts.Port)

	argsFunc := func(ctx process.InitContext) ([]string, error) {
		cfgData, err := controlplane.KubeConfigFromREST(cfg)
		if err != nil {
			return nil, fmt.Errorf("error getting kube config from rest: %w", err)
		}

		kubeconfigPath := filepath.Join(ctx.TempDir, "kubeconfig")
		if err := os.WriteFile(kubeconfigPath, cfgData, 0666); err != nil {
			return nil, fmt.Errorf("error writing kubeconfig: %w", err)
		}

		defaultArgs := process.EmptyArgs().
			Set("kubeconfig", kubeconfigPath).
			Set("health-probe-bind-address", fmt.Sprintf("%s:%d", opts.Host, opts.Port)).
			Set("metrics-bind-address", "0")

		return process.BuildArgs(opts.Args, defaultArgs, opts.MergeArgs), nil
	}

	proc, err := process.New(process.Options{
		MainPath:      opts.MainPath,
		BuildOptions:  opts.BuildOptions,
		Command:       opts.Command,
		ArgsFunc:      argsFunc,
		AttachOutput:  opts.AttachOutput,
		Stdout:        opts.Stdout,
		Stderr:        opts.Stderr,
		HealthTimeout: opts.HealthTimeout,
		WaitTimeout:   opts.WaitTimeout,
		HealthCheck:   process.InsecurePollHealthCheck(healthProbeBindAddress),
	})
	if err != nil {
		return nil, fmt.Errorf("error creating process: %w", err)
	}

	return &ControllerManager{
		process: proc,
	}, nil
}

func (a *ControllerManager) Start() error {
	return a.process.Start()
}

func (a *ControllerManager) Stop() error {
	return a.process.Stop()
}
