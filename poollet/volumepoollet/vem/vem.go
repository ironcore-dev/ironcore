// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package vem

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/iri/apis/volume"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/volumepoollet/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type VolumeEventMapper struct {
	manager.Runnable
	record.EventRecorder
	client.Client

	volumeRuntime volume.RuntimeService

	relistPeriod time.Duration
	lastFetched  time.Time
}

func (m *VolumeEventMapper) relist(ctx context.Context, log logr.Logger) error {
	log.V(1).Info("Relisting volume cluster events")
	toEventFilterTime := time.Now()
	res, err := m.volumeRuntime.ListEvents(ctx, &iri.ListEventsRequest{
		Filter: &iri.EventFilter{EventsFromTime: m.lastFetched.Unix(), EventsToTime: toEventFilterTime.Unix()},
	})
	if err != nil {
		return fmt.Errorf("error listing volume cluster events: %w", err)
	}

	m.lastFetched = toEventFilterTime
	for _, volumeEvent := range res.Events {
		if volumeEvent.Spec.InvolvedObjectMeta.Labels != nil {
			involvedVolume := &storagev1alpha1.Volume{
				ObjectMeta: metav1.ObjectMeta{
					UID:       types.UID(volumeEvent.Spec.InvolvedObjectMeta.Labels[v1alpha1.VolumeUIDLabel]),
					Name:      volumeEvent.Spec.InvolvedObjectMeta.Labels[v1alpha1.VolumeNameLabel],
					Namespace: volumeEvent.Spec.InvolvedObjectMeta.Labels[v1alpha1.VolumeNamespaceLabel],
				},
			}
			m.Eventf(involvedVolume, volumeEvent.Spec.Type, volumeEvent.Spec.Reason, volumeEvent.Spec.Message)
		}
	}

	return nil
}

func (m *VolumeEventMapper) Start(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("mem")
	m.lastFetched = time.Now()
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		if err := m.relist(ctx, log); err != nil {
			log.Error(err, "Error relisting")
		}
	}, m.relistPeriod)
	return nil
}

type VolumeEventMapperOptions struct {
	RelistPeriod time.Duration
}

func setVolumeEventMapperOptionsDefaults(o *VolumeEventMapperOptions) {
	if o.RelistPeriod == 0 {
		o.RelistPeriod = 1 * time.Minute
	}
}

func NewVolumeEventMapper(client client.Client, runtime volume.RuntimeService, recorder record.EventRecorder, opts VolumeEventMapperOptions) *VolumeEventMapper {
	setVolumeEventMapperOptionsDefaults(&opts)
	return &VolumeEventMapper{
		Client:        client,
		volumeRuntime: runtime,
		relistPeriod:  opts.RelistPeriod,
		EventRecorder: recorder,
	}
}
