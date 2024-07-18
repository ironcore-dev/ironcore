// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package bucket

import (
	"context"
	"fmt"

	"github.com/ironcore-dev/ironcore/iri/apis/bucket"
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type remoteRuntime struct {
	client iri.BucketRuntimeClient
}

func NewRemoteRuntime(endpoint string) (bucket.RuntimeService, error) {
	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("error dialing: %w", err)
	}

	return &remoteRuntime{
		client: iri.NewBucketRuntimeClient(conn),
	}, nil
}

func (r *remoteRuntime) ListEvents(ctx context.Context, req *iri.ListEventsRequest) (*iri.ListEventsResponse, error) {
	return r.client.ListEvents(ctx, req)
}

func (r *remoteRuntime) ListBuckets(ctx context.Context, request *iri.ListBucketsRequest) (*iri.ListBucketsResponse, error) {
	return r.client.ListBuckets(ctx, request)
}

func (r *remoteRuntime) CreateBucket(ctx context.Context, request *iri.CreateBucketRequest) (*iri.CreateBucketResponse, error) {
	return r.client.CreateBucket(ctx, request)
}

func (r *remoteRuntime) DeleteBucket(ctx context.Context, request *iri.DeleteBucketRequest) (*iri.DeleteBucketResponse, error) {
	return r.client.DeleteBucket(ctx, request)
}

func (r *remoteRuntime) ListBucketClasses(ctx context.Context, request *iri.ListBucketClassesRequest) (*iri.ListBucketClassesResponse, error) {
	return r.client.ListBucketClasses(ctx, request)
}
