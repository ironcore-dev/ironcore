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

	volumebrokerconfig "github.com/ironcore-dev/ironcore/broker/bucketbroker/client/config"
	"github.com/ironcore-dev/ironcore/broker/common"
	"github.com/ironcore-dev/ironcore/broker/volumebroker/server"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	"github.com/ironcore-dev/ironcore/utils/client/config"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type Options struct {
	GetConfigOptions        config.GetConfigOptions
	Address                 string
	BrokerDownwardAPILabels map[string]string

	QPS   float32
	Burst int

	Namespace          string
	VolumePoolName     string
	VolumePoolSelector map[string]string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.GetConfigOptions.BindFlags(fs)
	fs.StringVar(&o.Address, "address", "/var/run/iri-volumebroker.sock", "Address to listen on.")
	fs.StringToStringVar(&o.BrokerDownwardAPILabels, "broker-downward-api-label", nil, "The labels to broker via downward API. "+
		"Example is for instance to broker \"root-volume-uid\" initially obtained via \"volumepoollet.ironcore.dev/volume-uid\".")

	fs.StringVar(&o.Namespace, "namespace", o.Namespace, "Target Kubernetes namespace to use.")
	fs.StringVar(&o.VolumePoolName, "volume-pool-name", o.VolumePoolName, "Name of the target volume pool to pin volumes to, if any.")
	fs.StringToStringVar(&o.VolumePoolSelector, "volume-pool-selector", o.VolumePoolSelector, "Selector of the target volume pools to pin volumes to, if any.")
}

func Command() *cobra.Command {
	var (
		zapOpts = zap.Options{Development: true}
		opts    Options
	)

	cmd := &cobra.Command{
		Use: "volumebroker",
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

	getter, err := volumebrokerconfig.NewGetter()
	if err != nil {
		return fmt.Errorf("error creating new getter: %w", err)
	}

	cfg, err := getter.GetConfig(ctx, &opts.GetConfigOptions)
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}

	srv, err := server.New(cfg, server.Options{
		BrokerDownwardAPILabels: opts.BrokerDownwardAPILabels,
		Namespace:               opts.Namespace,
		VolumePoolName:          opts.VolumePoolName,
		VolumePoolSelector:      opts.VolumePoolSelector,
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
	iri.RegisterVolumeRuntimeServer(grpcSrv, srv)

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

	if err := srv.SetVolumeUIDLabelToAllVolumes(ctx); err != nil {
		return fmt.Errorf("failed to set volume uid label to all brokered volumes: %w", err)
	}
	return nil
}
