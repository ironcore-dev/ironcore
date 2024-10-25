// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package bcm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/ironcore-dev/ironcore/iri/apis/bucket"
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
)

type capabilities struct {
	tps  int64
	iops int64
}

func getCapabilities(iriCaps *iri.BucketClassCapabilities) capabilities {
	return capabilities{
		tps:  iriCaps.Tps,
		iops: iriCaps.Iops,
	}
}

type Generic struct {
	mu sync.RWMutex

	sync   bool
	synced chan struct{}

	bucketClassByName         map[string]*iri.BucketClass
	bucketClassByCapabilities map[capabilities][]*iri.BucketClass

	bucketRuntime bucket.RuntimeService

	relistPeriod time.Duration
}

func (g *Generic) relist(ctx context.Context, log logr.Logger) error {
	log.V(1).Info("Relisting bucket classes")
	res, err := g.bucketRuntime.ListBucketClasses(ctx, &iri.ListBucketClassesRequest{})
	if err != nil {
		return fmt.Errorf("error listing bucket classes: %w", err)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	maps.Clear(g.bucketClassByName)
	maps.Clear(g.bucketClassByCapabilities)

	for _, bucketClass := range res.BucketClasses {
		caps := getCapabilities(bucketClass.Capabilities)
		g.bucketClassByName[bucketClass.Name] = bucketClass
		g.bucketClassByCapabilities[caps] = append(g.bucketClassByCapabilities[caps], bucketClass)
	}

	if !g.sync {
		g.sync = true
		close(g.synced)
	}

	return nil
}

func (g *Generic) Start(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("vcm")
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		if err := g.relist(ctx, log); err != nil {
			log.Error(err, "Error relisting")
		}
	}, g.relistPeriod)
	return nil
}

func (g *Generic) GetBucketClassFor(ctx context.Context, name string, caps *iri.BucketClassCapabilities) (*iri.BucketClass, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	expected := getCapabilities(caps)
	if byName, ok := g.bucketClassByName[name]; ok && getCapabilities(byName.Capabilities) == expected {
		return byName, nil
	}

	if byCaps, ok := g.bucketClassByCapabilities[expected]; ok {
		switch len(byCaps) {
		case 0:
			return nil, ErrNoMatchingBucketClass
		case 1:
			return byCaps[0], nil
		default:
			return nil, ErrAmbiguousMatchingBucketClass
		}
	}

	return nil, ErrNoMatchingBucketClass
}

func (g *Generic) WaitForSync(ctx context.Context) error {
	select {
	case <-g.synced:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type GenericOptions struct {
	RelistPeriod time.Duration
}

func setGenericOptionsDefaults(o *GenericOptions) {
	if o.RelistPeriod == 0 {
		o.RelistPeriod = 1 * time.Hour
	}
}

func NewGeneric(runtime bucket.RuntimeService, opts GenericOptions) BucketClassMapper {
	setGenericOptionsDefaults(&opts)
	return &Generic{
		synced:                    make(chan struct{}),
		bucketClassByName:         map[string]*iri.BucketClass{},
		bucketClassByCapabilities: map[capabilities][]*iri.BucketClass{},
		bucketRuntime:             runtime,
		relistPeriod:              opts.RelistPeriod,
	}
}
