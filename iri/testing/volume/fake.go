// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package volume

import (
	"context"
	"sync"
	"time"

	"github.com/ironcore-dev/ironcore/broker/common/idgen"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/labels"
)

func filterInLabels(labelSelector, lbls map[string]string) bool {
	return labels.SelectorFromSet(labelSelector).Matches(labels.Set(lbls))
}

type FakeVolume struct {
	iri.Volume
}

type FakeVolumeClassStatus struct {
	iri.VolumeClassStatus
}

type FakeRuntimeService struct {
	sync.Mutex

	idGen idgen.IDGen

	Volumes             map[string]*FakeVolume
	VolumeClassesStatus map[string]*FakeVolumeClassStatus
}

func NewFakeRuntimeService() *FakeRuntimeService {
	return &FakeRuntimeService{
		idGen: idgen.Default,

		Volumes:             make(map[string]*FakeVolume),
		VolumeClassesStatus: make(map[string]*FakeVolumeClassStatus),
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

func (r *FakeRuntimeService) SetVolumeClasses(volumeClassStatus []*FakeVolumeClassStatus) {
	r.Lock()
	defer r.Unlock()

	r.VolumeClassesStatus = make(map[string]*FakeVolumeClassStatus)
	for _, status := range volumeClassStatus {
		r.VolumeClassesStatus[status.VolumeClass.Name] = status
	}
}

func (r *FakeRuntimeService) ListVolumes(ctx context.Context, req *iri.ListVolumesRequest, opts ...grpc.CallOption) (*iri.ListVolumesResponse, error) {
	r.Lock()
	defer r.Unlock()

	filter := req.Filter

	var res []*iri.Volume
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
	return &iri.ListVolumesResponse{Volumes: res}, nil
}

func (r *FakeRuntimeService) CreateVolume(ctx context.Context, req *iri.CreateVolumeRequest, opts ...grpc.CallOption) (*iri.CreateVolumeResponse, error) {
	r.Lock()
	defer r.Unlock()

	volume := *req.Volume
	volume.Metadata.Id = r.idGen.Generate()
	volume.Metadata.CreatedAt = time.Now().UnixNano()
	volume.Status = &iri.VolumeStatus{}

	r.Volumes[volume.Metadata.Id] = &FakeVolume{
		Volume: volume,
	}

	return &iri.CreateVolumeResponse{
		Volume: &volume,
	}, nil
}

func (r *FakeRuntimeService) ExpandVolume(ctx context.Context, req *iri.ExpandVolumeRequest, opts ...grpc.CallOption) (*iri.ExpandVolumeResponse, error) {
	r.Lock()
	defer r.Unlock()

	volume, ok := r.Volumes[req.VolumeId]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "volume %q not found", req.VolumeId)
	}

	volume.Spec.Resources.StorageBytes = req.Resources.StorageBytes

	return &iri.ExpandVolumeResponse{}, nil
}

func (r *FakeRuntimeService) DeleteVolume(ctx context.Context, req *iri.DeleteVolumeRequest, opts ...grpc.CallOption) (*iri.DeleteVolumeResponse, error) {
	r.Lock()
	defer r.Unlock()

	volumeID := req.VolumeId
	if _, ok := r.Volumes[volumeID]; !ok {
		return nil, status.Errorf(codes.NotFound, "volume %q not found", volumeID)
	}

	delete(r.Volumes, volumeID)
	return &iri.DeleteVolumeResponse{}, nil
}

func (r *FakeRuntimeService) Status(ctx context.Context, req *iri.StatusRequest, opts ...grpc.CallOption) (*iri.StatusResponse, error) {
	r.Lock()
	defer r.Unlock()

	var res []*iri.VolumeClassStatus
	for _, m := range r.VolumeClassesStatus {
		volumeClassStatus := m.VolumeClassStatus
		res = append(res, &volumeClassStatus)
	}
	return &iri.StatusResponse{VolumeClassStatus: res}, nil
}
