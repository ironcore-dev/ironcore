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
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Command(streams clicommon.Streams, clientFactory common.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bucket id [ids...]",
		Aliases: common.BucketAliases,
		Args:    cobra.MinimumNArgs(1),
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

			ids := args

			return Run(cmd.Context(), streams, client, ids)
		},
	}

	return cmd
}

func Run(ctx context.Context, streams clicommon.Streams, client ori.BucketRuntimeClient, ids []string) error {
	for _, id := range ids {
		if _, err := client.DeleteBucket(ctx, &ori.DeleteBucketRequest{
			BucketId: id,
		}); err != nil {
			if status.Code(err) != codes.NotFound {
				return fmt.Errorf("error deleting bucket %s: %w", id, err)
			}

			_, _ = fmt.Fprintf(streams.Out, "Bucket %s not found\n", id)
		} else {
			_, _ = fmt.Fprintf(streams.Out, "Bucket %s deleted\n", id)
		}
	}
	return nil
}
