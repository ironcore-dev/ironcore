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

package bucket

import (
	"context"
	"fmt"
	"os"

	ori "github.com/ironcore-dev/ironcore/ori/apis/bucket/v1alpha1"
	"github.com/ironcore-dev/ironcore/orictl-bucket/cmd/orictl-bucket/orictlbucket/common"
	orictlcmd "github.com/ironcore-dev/ironcore/orictl/cmd"
	"github.com/ironcore-dev/ironcore/orictl/decoder"
	"github.com/ironcore-dev/ironcore/orictl/renderer"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Options struct {
	Filename string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Filename, "filename", "f", o.Filename, "Path to a file to read.")
}

func Command(streams orictlcmd.Streams, clientFactory common.ClientFactory) *cobra.Command {
	var (
		outputOpts = common.NewOutputOptions()
		opts       Options
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

			r, err := outputOpts.RendererOrNil()
			if err != nil {
				return err
			}

			return Run(ctx, streams, client, r, opts)
		},
	}

	outputOpts.AddFlags(cmd.Flags())
	opts.AddFlags(cmd.Flags())

	return cmd
}

func Run(ctx context.Context, streams orictlcmd.Streams, client ori.BucketRuntimeClient, r renderer.Renderer, opts Options) error {
	data, err := orictlcmd.ReadFileOrReader(opts.Filename, os.Stdin)
	if err != nil {
		return err
	}

	bucket := &ori.Bucket{}
	if err := decoder.Decode(data, bucket); err != nil {
		return err
	}

	res, err := client.CreateBucket(ctx, &ori.CreateBucketRequest{Bucket: bucket})
	if err != nil {
		return err
	}

	if r != nil {
		return r.Render(res.Bucket, streams.Out)
	}

	_, _ = fmt.Fprintf(streams.Out, "Created bucket %s\n", res.Bucket.Metadata.Id)
	return nil
}
