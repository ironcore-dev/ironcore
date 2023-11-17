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

package orictlbucket

import (
	goflag "flag"

	"github.com/ironcore-dev/ironcore/orictl-bucket/cmd/orictl-bucket/orictlbucket/common"
	"github.com/ironcore-dev/ironcore/orictl-bucket/cmd/orictl-bucket/orictlbucket/create"
	delete2 "github.com/ironcore-dev/ironcore/orictl-bucket/cmd/orictl-bucket/orictlbucket/delete"
	"github.com/ironcore-dev/ironcore/orictl-bucket/cmd/orictl-bucket/orictlbucket/get"
	orictlcmd "github.com/ironcore-dev/ironcore/orictl/cmd"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func Command(streams orictlcmd.Streams) *cobra.Command {
	var (
		zapOpts    zap.Options
		clientOpts common.ClientOptions
	)

	cmd := &cobra.Command{
		Use: "orictl-bucket",
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
