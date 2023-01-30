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

package bucket

import (
	"context"
	"fmt"

	ori "github.com/onmetal/onmetal-api/ori/apis/bucket/v1alpha1"
	clicommon "github.com/onmetal/onmetal-api/orictl/cli/common"
	"github.com/onmetal/onmetal-api/orictl/cmd/orictl-bucket/orictlbucket/common"
	"github.com/onmetal/onmetal-api/orictl/renderer"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Options struct {
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
}

func Command(streams clicommon.Streams, clientFactory common.ClientFactory) *cobra.Command {
	var (
		opts       Options
		outputOpts = common.NewOutputOptions()
	)

	cmd := &cobra.Command{
		Use:     "bucket",
		Aliases: common.BucketAliases,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := ctrl.LoggerFrom(ctx)

			client, cleanup, err := clientFactory.New()
			if err != nil {
				return err
			}
			defer func() {
				if err := cleanup(); err != nil {
					log.Error(err, "Error cleaning up")
				}
			}()

			render, err := outputOpts.Renderer("table")
			if err != nil {
				return err
			}

			return Run(cmd.Context(), streams, client, render, opts)
		},
	}

	outputOpts.AddFlags(cmd.Flags())
	opts.AddFlags(cmd.Flags())

	return cmd
}

func Run(ctx context.Context, streams clicommon.Streams, client ori.BucketRuntimeClient, render renderer.Renderer, opts Options) error {
	res, err := client.ListBuckets(ctx, &ori.ListBucketsRequest{})
	if err != nil {
		return fmt.Errorf("error listing buckets: %w", err)
	}

	return render.Render(res.Buckets, streams.Out)
}
