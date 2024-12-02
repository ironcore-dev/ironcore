// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	goflag "flag"
	"fmt"
	"net"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"k8s.io/apiserver/pkg/server/egressselector"

	"github.com/ironcore-dev/ironcore/broker/bucketbroker/server"
	"github.com/ironcore-dev/ironcore/broker/common"
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	"github.com/ironcore-dev/ironcore/utils/client/config"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type Options struct {
	GetConfigOptions config.GetConfigOptions
	Address          string

	QPS   float32
	Burst int

	Namespace          string
	BucketPoolName     string
	BucketPoolSelector map[string]string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.GetConfigOptions.BindFlags(fs)
	fs.StringVar(&o.Address, "address", "/var/run/iri-bucketbroker.sock", "Address to listen on.")

	fs.Float32Var(&o.QPS, "qps", config.QPS, "Kubernetes client qps.")
	fs.IntVar(&o.Burst, "burst", config.Burst, "Kubernetes client burst.")

	fs.StringVar(&o.Namespace, "namespace", o.Namespace, "Target Kubernetes namespace to use.")
	fs.StringVar(&o.BucketPoolName, "bucket-pool-name", o.BucketPoolName, "Name of the target bucket pool to pin buckets to, if any.")
	fs.StringToStringVar(&o.BucketPoolSelector, "bucket-pool-selector", o.BucketPoolSelector, "Selector of the target bucket pools to pin buckets to, if any.")
}

func Command() *cobra.Command {
	var (
		zapOpts = zap.Options{Development: true}
		opts    Options
	)

	cmd := &cobra.Command{
		Use: "bucketbroker",
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

	getter, err := newGetter()
	if err != nil {
		return fmt.Errorf("error creating new getter: %w", err)
	}

	cfg, err := getter.GetConfig(ctx, &opts.GetConfigOptions)
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}

	srv, err := server.New(cfg, server.Options{
		Namespace:          opts.Namespace,
		BucketPoolName:     opts.BucketPoolName,
		BucketPoolSelector: opts.BucketPoolSelector,
	})
	if err != nil {
		return fmt.Errorf("error creating server: %w", err)
	}

	log.V(1).Info("Cleaning up any previous socket")
	if err := common.CleanupSocketIfExists(opts.Address); err != nil {
		return fmt.Errorf("error cleaning up socket: %w", err)
	}

	log.V(1).Info("Start listening on unix socket", "Address", opts.Address)
	l, err := net.Listen("unix", opts.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer func() {
		if err := l.Close(); err != nil {
			log.Error(err, "Error closing socket")
		}
	}()

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
	iri.RegisterBucketRuntimeServer(grpcSrv, srv)

	setupLog.Info("Starting server", "Address", l.Addr().String())
	go func() {
		defer func() {
			setupLog.Info("Shutting down server")
			grpcSrv.Stop()
			setupLog.Info("Shut down server")
		}()
		<-ctx.Done()
	}()
	if err := grpcSrv.Serve(l); err != nil {
		return fmt.Errorf("error serving: %w", err)
	}
	return nil
}

func newGetter() (*config.BrokerGetter, error) {
	return config.NewBrokerGetter(config.GetterOptions{
		Name:           "volumebroker",
		NetworkContext: egressselector.ControlPlane.AsNetworkContext(),
	})
}
