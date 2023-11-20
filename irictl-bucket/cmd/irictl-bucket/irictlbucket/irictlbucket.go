// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package irictlbucket

import (
	goflag "flag"

	"github.com/ironcore-dev/ironcore/irictl-bucket/cmd/irictl-bucket/irictlbucket/common"
	"github.com/ironcore-dev/ironcore/irictl-bucket/cmd/irictl-bucket/irictlbucket/create"
	delete2 "github.com/ironcore-dev/ironcore/irictl-bucket/cmd/irictl-bucket/irictlbucket/delete"
	"github.com/ironcore-dev/ironcore/irictl-bucket/cmd/irictl-bucket/irictlbucket/get"
	irictlcmd "github.com/ironcore-dev/ironcore/irictl/cmd"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func Command(streams irictlcmd.Streams) *cobra.Command {
	var (
		zapOpts    zap.Options
		clientOpts common.ClientOptions
	)

	cmd := &cobra.Command{
		Use: "irictl-bucket",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logger := zap.New(zap.UseFlagOptions(&zapOpts))
			ctrl.SetLogger(logger)
			cmd.SetContext(ctrl.LoggerInto(cmd.Context(), ctrl.Log))
		},
	}

	goFlags := goflag.NewFlagSet("", 0)
	zapOpts.BindFlags(goFlags)

	cmd.PersistentFlags().AddGoFlagSet(goFlags)
	clientOpts.AddFlags(cmd.PersistentFlags())

	cmd.AddCommand(
		get.Command(streams, &clientOpts),
		delete2.Command(streams, &clientOpts),
		create.Command(streams, &clientOpts),
	)

	return cmd
}
