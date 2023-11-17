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

package bcm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	ori "github.com/ironcore-dev/ironcore/ori/apis/bucket/v1alpha1"
	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
)

type capabilities struct {
	tps  int64
	iops int64
}

func getCapabilities(oriCaps *ori.BucketClassCapabilities) capabilities {
	return capabilities{
		tps:  oriCaps.Tps,
		iops: oriCaps.Iops,
	}
}

type Generic struct {
	mu sync.RWMutex

	sync   bool
	synced chan struct{}

	bucketClassByName         map[string]*ori.BucketClass
	bucketClassByCapabilities map[capabilities][]*ori.BucketClass

	bucketRuntime ori.BucketRuntimeClient

	relistPeriod time.Duration
}

func (g *Generic) relist(ctx context.Context, log logr.Logger) error {
	log.V(1).Info("Relisting bucket classes")
	res, err := g.bucketRuntime.ListBucketClasses(ctx, &ori.ListBucketClassesRequest{})
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

func (g *Generic) GetBucketClassFor(ctx context.Context, name string, caps *ori.BucketClassCapabilities) (*ori.BucketClass, error) {
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
			class := *byCaps[0]
			return &class, nil
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

func NewGeneric(runtime ori.BucketRuntimeClient, opts GenericOptions) BucketClassMapper {
	setGenericOptionsDefaults(&opts)
	return &Generic{
		synced:                    make(chan struct{}),
		bucketClassByName:         map[string]*ori.BucketClass{},
		bucketClassByCapabilities: map[capabilities][]*ori.BucketClass{},
		bucketRuntime:             runtime,
		relistPeriod:              opts.RelistPeriod,
	}
}
