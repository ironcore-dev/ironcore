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

package app

import (
	"context"
	goflag "flag"
	"fmt"
	"net"

	"github.com/onmetal/controller-utils/configutils"
	"github.com/onmetal/onmetal-api/machinebroker/server"
	ori "github.com/onmetal/onmetal-api/ori/apis/runtime/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type Options struct {
	Kubeconfig string
	Address    string

	Namespace           string
	MachinePoolName     string
	MachinePoolSelector map[string]string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Kubeconfig, "kubeconfig", o.Kubeconfig, "Path pointing to a kubeconfig file to use.")
	fs.StringVar(&o.Address, "address", "/var/run/ori-machinebroker.sock", "Address to listen on.")

	fs.StringVar(&o.Namespace, "namespace", o.Namespace, "Target Kubernetes namespace to use.")
	fs.StringVar(&o.MachinePoolName, "machine-pool-name", o.MachinePoolName, "Name of the target machine pool to pin machines to, if any.")
	fs.StringToStringVar(&o.MachinePoolSelector, "machine-pool-selector", o.MachinePoolSelector, "Selector of the target machine pools to pin machines to, if any.")
}

func Command() *cobra.Command {
	var (
		zapOpts = zap.Options{Development: true}
		opts    Options
	)

	cmd := &cobra.Command{
		Use: "machinebroker",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logger := zap.New(zap.UseFlagOptions(&zapOpts))
			ctrl.SetLogger(logger)
			cmd.SetContext(ctrl.LoggerInto(cmd.Context(), ctrl.Log))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(cmd.Context(), opts)
		},
	}

	goFlags := goflag.NewFlagSet("", 0)
	zapOpts.BindFlags(goFlags)
	cmd.PersistentFlags().AddGoFlagSet(goFlags)

	opts.AddFlags(cmd.Flags())

	return cmd
}

func Run(ctx context.Context, opts Options) error {
	log := ctrl.LoggerFrom(ctx)
	setupLog := log.WithName("setup")

	cfg, err := configutils.GetConfig(configutils.Kubeconfig(opts.Kubeconfig))
	if err != nil {
		return err
	}

	srv, err := server.New(cfg, server.Options{
		Namespace:           opts.Namespace,
		MachinePoolName:     opts.MachinePoolName,
		MachinePoolSelector: opts.MachinePoolSelector,
	})
	if err != nil {
		return fmt.Errorf("error creating server: %w", err)
	}

	l, err := net.Listen("unix", opts.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			log := log.WithName(info.FullMethod)
			ctx = ctrl.LoggerInto(ctx, log)
			log.V(1).Info("Request")
			resp, err = handler(ctx, req)
			if err != nil {
				log.Error(err, "Error handling request")
			}
			return resp, err
		}),
	)
	ori.RegisterMachineRuntimeServer(grpcSrv, srv)

	setupLog.Info("Starting server", "Address", l.Addr().String())
	go func() {
		<-ctx.Done()
		setupLog.Info("Shutting down server")
		grpcSrv.Stop()
		setupLog.Info("Shut down server")
	}()
	if err := grpcSrv.Serve(l); err != nil {
		return fmt.Errorf("error serving: %w", err)
	}
	return nil
}
