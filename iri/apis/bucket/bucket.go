// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package bucket

import (
	"context"

	api "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
)

type RuntimeService interface {
	Version(context.Context, *api.VersionRequest) (*api.VersionResponse, error)
	ListEvents(context.Context, *api.ListEventsRequest) (*api.ListEventsResponse, error)
	ListBuckets(context.Context, *api.ListBucketsRequest) (*api.ListBucketsResponse, error)
	CreateBucket(context.Context, *api.CreateBucketRequest) (*api.CreateBucketResponse, error)
	ListBucketClasses(ctx context.Context, request *api.ListBucketClassesRequest) (*api.ListBucketClassesResponse, error)
	DeleteBucket(context.Context, *api.DeleteBucketRequest) (*api.DeleteBucketResponse, error)
}
