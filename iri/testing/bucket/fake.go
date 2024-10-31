// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package bucket

import (
	"context"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	irievent "github.com/ironcore-dev/ironcore/iri/apis/event/v1alpha1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/ironcore-dev/ironcore/broker/common/idgen"
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FakeBucket struct {
	iri.Bucket
}

type FakeBucketClass struct {
	iri.BucketClass
}

type FakeEvent struct {
	irievent.Event
}

type FakeRuntimeService struct {
	sync.Mutex

	idGen idgen.IDGen

	Buckets       map[string]*FakeBucket
	BucketClasses map[string]*FakeBucketClass
	Events        []*FakeEvent
}

func NewFakeRuntimeService() *FakeRuntimeService {
	return &FakeRuntimeService{
		idGen: idgen.Default,

		Buckets:       make(map[string]*FakeBucket),
		BucketClasses: make(map[string]*FakeBucketClass),
		Events:        []*FakeEvent{},
	}
}

func (r *FakeRuntimeService) SetBuckets(buckets []*FakeBucket) {
	r.Lock()
	defer r.Unlock()

	r.Buckets = make(map[string]*FakeBucket)
	for _, bucket := range buckets {
		r.Buckets[bucket.Metadata.Id] = bucket
	}
}

func (r *FakeRuntimeService) SetBucketClasses(bucketClasses []*FakeBucketClass) {
	r.Lock()
	defer r.Unlock()

	r.BucketClasses = make(map[string]*FakeBucketClass)
	for _, class := range bucketClasses {
		r.BucketClasses[class.Name] = class
	}
}

func (r *FakeRuntimeService) SetEvents(events []*FakeEvent) {
	r.Lock()
	defer r.Unlock()

	r.Events = events
}

func (r *FakeRuntimeService) ListEvents(ctx context.Context, req *iri.ListEventsRequest) (*iri.ListEventsResponse, error) {
	r.Lock()
	defer r.Unlock()

	var res []*irievent.Event
	for _, e := range r.Events {
		event := e.Event
		res = append(res, &event)
	}

	return &iri.ListEventsResponse{Events: res}, nil
}

func (r *FakeRuntimeService) ListBuckets(ctx context.Context, req *iri.ListBucketsRequest) (*iri.ListBucketsResponse, error) {
	r.Lock()
	defer r.Unlock()

	filter := req.Filter

	var res []*iri.Bucket
	for _, v := range r.Buckets {
		if filter != nil {
			if filter.Id != "" && filter.Id != v.Metadata.Id {
				continue
			}
			if filter.LabelSelector != nil && !filterInLabels(filter.LabelSelector, v.Metadata.Labels) {
				continue
			}
		}

		bucket := v.Bucket
		res = append(res, &bucket)
	}
	return &iri.ListBucketsResponse{Buckets: res}, nil
}

func (r *FakeRuntimeService) CreateBucket(ctx context.Context, req *iri.CreateBucketRequest) (*iri.CreateBucketResponse, error) {
	r.Lock()
	defer r.Unlock()

	bucket := *req.Bucket
	bucket.Metadata.Id = r.idGen.Generate()
	bucket.Metadata.CreatedAt = time.Now().UnixNano()
	bucket.Status = &iri.BucketStatus{}

	r.Buckets[bucket.Metadata.Id] = &FakeBucket{
		Bucket: bucket,
	}

	return &iri.CreateBucketResponse{
		Bucket: &bucket,
	}, nil
}

func (r *FakeRuntimeService) DeleteBucket(ctx context.Context, req *iri.DeleteBucketRequest) (*iri.DeleteBucketResponse, error) {
	r.Lock()
	defer r.Unlock()

	bucketID := req.BucketId
	if _, ok := r.Buckets[bucketID]; !ok {
		return nil, status.Errorf(codes.NotFound, "bucket %q not found", bucketID)
	}

	delete(r.Buckets, bucketID)
	return &iri.DeleteBucketResponse{}, nil
}

func (r *FakeRuntimeService) ListBucketClasses(ctx context.Context, req *iri.ListBucketClassesRequest) (*iri.ListBucketClassesResponse, error) {
	r.Lock()
	defer r.Unlock()

	var res []*iri.BucketClass
	for _, b := range r.BucketClasses {
		bucketClass := proto.Clone(&b.BucketClass).(*iri.BucketClass)
		res = append(res, bucketClass)
	}
	return &iri.ListBucketClassesResponse{BucketClasses: res}, nil
}

func filterInLabels(labelSelector, lbls map[string]string) bool {
	return labels.SelectorFromSet(labelSelector).Matches(labels.Set(lbls))
}
