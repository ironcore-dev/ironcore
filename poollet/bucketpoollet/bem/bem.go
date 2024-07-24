// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package bem

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/iri/apis/bucket"
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/bucketpoollet/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type BucketEventMapper struct {
	manager.Runnable
	record.EventRecorder
	client.Client

	bucketRuntime bucket.RuntimeService

	relistPeriod time.Duration
	lastFetched  time.Time
}

func (m *BucketEventMapper) relist(ctx context.Context, log logr.Logger) error {
	log.V(1).Info("Relisting bucket cluster events")
	toEventFilterTime := time.Now()
	res, err := m.bucketRuntime.ListEvents(ctx, &iri.ListEventsRequest{
		Filter: &iri.EventFilter{EventsFromTime: m.lastFetched.Unix(), EventsToTime: toEventFilterTime.Unix()},
	})
	if err != nil {
		return fmt.Errorf("error listing bucket cluster events: %w", err)
	}

	m.lastFetched = toEventFilterTime
	for _, bucketEvent := range res.Events {
		if bucketEvent.Spec.InvolvedObjectMeta.Labels != nil {
			involvedBucket := &storagev1alpha1.Bucket{
				ObjectMeta: metav1.ObjectMeta{
					UID:       types.UID(bucketEvent.Spec.InvolvedObjectMeta.Labels[v1alpha1.BucketUIDLabel]),
					Name:      bucketEvent.Spec.InvolvedObjectMeta.Labels[v1alpha1.BucketNameLabel],
					Namespace: bucketEvent.Spec.InvolvedObjectMeta.Labels[v1alpha1.BucketNamespaceLabel],
				},
			}
			m.Eventf(involvedBucket, bucketEvent.Spec.Type, bucketEvent.Spec.Reason, bucketEvent.Spec.Message)
		}
	}

	return nil
}

func (m *BucketEventMapper) Start(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("bem")
	m.lastFetched = time.Now()
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		if err := m.relist(ctx, log); err != nil {
			log.Error(err, "Error relisting")
		}
	}, m.relistPeriod)
	return nil
}

type BucketEventMapperOptions struct {
	RelistPeriod time.Duration
}

func setBucketEventMapperOptionsDefaults(o *BucketEventMapperOptions) {
	if o.RelistPeriod == 0 {
		o.RelistPeriod = 1 * time.Minute
	}
}

func NewBucketEventMapper(client client.Client, runtime bucket.RuntimeService, recorder record.EventRecorder, opts BucketEventMapperOptions) *BucketEventMapper {
	setBucketEventMapperOptionsDefaults(&opts)
	return &BucketEventMapper{
		Client:        client,
		bucketRuntime: runtime,
		relistPeriod:  opts.RelistPeriod,
		EventRecorder: recorder,
	}
}
