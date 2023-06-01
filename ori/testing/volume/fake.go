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

package volume

import (
	"context"
	"sync"
	"time"

	"github.com/onmetal/onmetal-api/broker/common/idgen"
	orimetrics "github.com/onmetal/onmetal-api/ori/apis/metrics/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/volume/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/labels"
)

func filterInLabels(labelSelector, lbls map[string]string) bool {
	return labels.SelectorFromSet(labelSelector).Matches(labels.Set(lbls))
}

type FakeVolume struct {
	ori.Volume
}

type FakeVolumeClass struct {
	ori.VolumeClass
}

type FakeMetric struct {
	orimetrics.Metric
}

type FakeMetricDescriptor struct {
	orimetrics.MetricDescriptor
}

type FakeRuntimeService struct {
	sync.Mutex

	idGen idgen.IDGen

	Volumes       map[string]*FakeVolume
	VolumeClasses map[string]*FakeVolumeClass

	MetricDesc []*FakeMetricDescriptor
	Metrics    []*FakeMetric
}

func NewFakeRuntimeService() *FakeRuntimeService {
	return &FakeRuntimeService{
		idGen: idgen.Default,

		Volumes:       make(map[string]*FakeVolume),
		VolumeClasses: make(map[string]*FakeVolumeClass),
	}
}

func (r *FakeRuntimeService) SetVolumes(volumes []*FakeVolume) {
	r.Lock()
	defer r.Unlock()

	r.Volumes = make(map[string]*FakeVolume)
	for _, volume := range volumes {
		r.Volumes[volume.Metadata.Id] = volume
	}
}

func (r *FakeRuntimeService) SetVolumeClasses(volumeClasses []*FakeVolumeClass) {
	r.Lock()
	defer r.Unlock()

	r.VolumeClasses = make(map[string]*FakeVolumeClass)
	for _, volumeClass := range volumeClasses {
		r.VolumeClasses[volumeClass.Name] = volumeClass
	}
}

func (r *FakeRuntimeService) ListVolumes(ctx context.Context, req *ori.ListVolumesRequest, opts ...grpc.CallOption) (*ori.ListVolumesResponse, error) {
	r.Lock()
	defer r.Unlock()

	filter := req.Filter

	var res []*ori.Volume
	for _, v := range r.Volumes {
		if filter != nil {
			if filter.Id != "" && filter.Id != v.Metadata.Id {
				continue
			}
			if filter.LabelSelector != nil && !filterInLabels(filter.LabelSelector, v.Metadata.Labels) {
				continue
			}
		}

		volume := v.Volume
		res = append(res, &volume)
	}
	return &ori.ListVolumesResponse{Volumes: res}, nil
}

func (r *FakeRuntimeService) CreateVolume(ctx context.Context, req *ori.CreateVolumeRequest, opts ...grpc.CallOption) (*ori.CreateVolumeResponse, error) {
	r.Lock()
	defer r.Unlock()

	volume := *req.Volume
	volume.Metadata.Id = r.idGen.Generate()
	volume.Metadata.CreatedAt = time.Now().UnixNano()
	volume.Status = &ori.VolumeStatus{}

	r.Volumes[volume.Metadata.Id] = &FakeVolume{
		Volume: volume,
	}

	return &ori.CreateVolumeResponse{
		Volume: &volume,
	}, nil
}

func (r *FakeRuntimeService) ExpandVolume(ctx context.Context, req *ori.ExpandVolumeRequest, opts ...grpc.CallOption) (*ori.ExpandVolumeResponse, error) {
	r.Lock()
	defer r.Unlock()

	volume, ok := r.Volumes[req.VolumeId]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "volume %q not found", req.VolumeId)
	}

	volume.Spec.Resources.StorageBytes = req.Resources.StorageBytes

	return &ori.ExpandVolumeResponse{}, nil
}

func (r *FakeRuntimeService) DeleteVolume(ctx context.Context, req *ori.DeleteVolumeRequest, opts ...grpc.CallOption) (*ori.DeleteVolumeResponse, error) {
	r.Lock()
	defer r.Unlock()

	volumeID := req.VolumeId
	if _, ok := r.Volumes[volumeID]; !ok {
		return nil, status.Errorf(codes.NotFound, "volume %q not found", volumeID)
	}

	delete(r.Volumes, volumeID)
	return &ori.DeleteVolumeResponse{}, nil
}

func (r *FakeRuntimeService) ListVolumeClasses(ctx context.Context, req *ori.ListVolumeClassesRequest, opts ...grpc.CallOption) (*ori.ListVolumeClassesResponse, error) {
	r.Lock()
	defer r.Unlock()

	var res []*ori.VolumeClass
	for _, m := range r.VolumeClasses {
		volumeClass := m.VolumeClass
		res = append(res, &volumeClass)
	}
	return &ori.ListVolumeClassesResponse{VolumeClasses: res}, nil
}

func (r *FakeRuntimeService) ListMetricDescriptors(ctx context.Context, req *ori.ListMetricDescriptorsRequest, opts ...grpc.CallOption) (*ori.ListMetricDescriptorsResponse, error) {
	r.Lock()
	defer r.Unlock()

	var res []*orimetrics.MetricDescriptor
	for _, m := range r.MetricDesc {
		metricDescriptor := m.MetricDescriptor
		res = append(res, &metricDescriptor)
	}
	return &ori.ListMetricDescriptorsResponse{Descriptors: res}, nil
}

func (r *FakeRuntimeService) ListMetrics(ctx context.Context, req *ori.ListMetricsRequest, opts ...grpc.CallOption) (*ori.ListMetricsResponse, error) {
	r.Lock()
	defer r.Unlock()

	var res []*orimetrics.Metric
	for _, m := range r.Metrics {
		metric := m.Metric
		res = append(res, &metric)
	}
	return &ori.ListMetricsResponse{Metrics: res}, nil
}
