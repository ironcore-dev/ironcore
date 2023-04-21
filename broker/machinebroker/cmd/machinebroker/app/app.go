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
	"errors"
	goflag "flag"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/onmetal/controller-utils/configutils"
	"github.com/onmetal/onmetal-api/broker/common"
	commongrpc "github.com/onmetal/onmetal-api/broker/common/grpc"
	machinebrokerhttp "github.com/onmetal/onmetal-api/broker/machinebroker/http"
	"github.com/onmetal/onmetal-api/broker/machinebroker/server"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type Options struct {
	Kubeconfig                 string
	Address                    string
	StreamingAddress           string
	BaseURL                    string
	DefaultRootMachineUIDLabel string

	Namespace           string
	MachinePoolName     string
	MachinePoolSelector map[string]string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Kubeconfig, "kubeconfig", o.Kubeconfig, "Path pointing to a kubeconfig file to use.")
	fs.StringVar(&o.Address, "address", "/var/run/ori-machinebroker.sock", "Address to listen on.")
	fs.StringVar(&o.StreamingAddress, "streaming-address", "127.0.0.1:20251", "Address to run the streaming server on")
	fs.StringVar(&o.BaseURL, "base-url", "", "The base url to construct urls for streaming from. If empty it will be "+
		"constructed from the streaming-address")
	fs.StringVar(&o.DefaultRootMachineUIDLabel, "default-root-machine-uid-label", "", "The default label to look up the root UID from.")

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

	if opts.Namespace == "" {
		return fmt.Errorf("must specify namespace")
	}

	baseURL := opts.BaseURL
	if baseURL == "" {
		u := &url.URL{
			Scheme: "http",
			Host:   opts.StreamingAddress,
		}
		baseURL = u.String()
	}

	log.V(1).Info("Creating server",
		"Namespace", opts.Namespace,
		"MachinePoolName", opts.MachinePoolName,
		"MachinePoolSelector", opts.MachinePoolSelector,
		"BaseURL", baseURL,
		"DefaultRootMachineUIDLabel", opts.DefaultRootMachineUIDLabel,
	)

	srv, err := server.New(cfg, opts.Namespace, server.Options{
		BaseURL:                    baseURL,
		DefaultRootMachineUIDLabel: opts.DefaultRootMachineUIDLabel,
		MachinePoolName:            opts.MachinePoolName,
		MachinePoolSelector:        opts.MachinePoolSelector,
	})
	if err != nil {
		return fmt.Errorf("error creating server: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return runGRPCServer(ctx, setupLog, log, srv, opts)
	})
	g.Go(func() error {
		return runStreamingServer(ctx, setupLog, log, srv, opts)
	})
	return g.Wait()
}

func runGRPCServer(ctx context.Context, setupLog logr.Logger, log logr.Logger, srv *server.Server, opts Options) error {
	log.V(1).Info("Cleaning up any previous socket")
	if err := common.CleanupSocketIfExists(opts.Address); err != nil {
		return fmt.Errorf("error cleaning up socket: %w", err)
	}

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			commongrpc.InjectLogger(log),
			commongrpc.LogRequest,
		),
	)
	ori.RegisterMachineRuntimeServer(grpcSrv, srv)

	log.V(1).Info("Start listening on unix socket", "Address", opts.Address)
	l, err := net.Listen("unix", opts.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	setupLog.Info("Starting server", "Address", l.Addr().String())
	go func() {
		<-ctx.Done()
		setupLog.Info("Shutting down server")
		grpcSrv.GracefulStop()
		setupLog.Info("Shut down server")
	}()
	if err := grpcSrv.Serve(l); err != nil {
		return fmt.Errorf("error serving: %w", err)
	}
	return nil
}

func runStreamingServer(ctx context.Context, setupLog, log logr.Logger, srv *server.Server, opts Options) error {
	httpHandler := machinebrokerhttp.NewHandler(srv, machinebrokerhttp.HandlerOptions{
		Log: log.WithName("server"),
	})

	httpSrv := &http.Server{
		Addr:    opts.StreamingAddress,
		Handler: httpHandler,
	}

	go func() {
		<-ctx.Done()
		setupLog.Info("Shutting down http server")
		_ = httpSrv.Close()
		setupLog.Info("Shut down http server")
	}()

	log.V(1).Info("Starting streaming server", "Address", opts.StreamingAddress)
	if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("error listening / serving streaming server: %w", err)
	}
	log.V(1).Info("Server finished")
	return nil
}
