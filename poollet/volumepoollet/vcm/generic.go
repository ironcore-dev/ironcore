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

package vcm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	ori "github.com/onmetal/onmetal-api/ori/apis/volume/v1alpha1"
	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
)

type capabilities struct {
	tps  int64
	iops int64
}

func getCapabilities(oriCaps *ori.VolumeClassCapabilities) capabilities {
	return capabilities{
		tps:  oriCaps.Tps,
		iops: oriCaps.Iops,
	}
}

type Generic struct {
	mu sync.RWMutex

	sync   bool
	synced chan struct{}

	volumeClassByName         map[string]*ori.VolumeClass
	volumeClassByCapabilities map[capabilities][]*ori.VolumeClass

	volumeRuntime ori.VolumeRuntimeClient

	relistPeriod time.Duration
}

func (g *Generic) relist(ctx context.Context, log logr.Logger) error {
	log.V(1).Info("Relisting volume classes")
	res, err := g.volumeRuntime.ListVolumeClasses(ctx, &ori.ListVolumeClassesRequest{})
	if err != nil {
		return fmt.Errorf("error listing volume classes: %w", err)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	maps.Clear(g.volumeClassByName)
	maps.Clear(g.volumeClassByCapabilities)

	for _, volumeClass := range res.VolumeClasses {
		caps := getCapabilities(volumeClass.Capabilities)
		g.volumeClassByName[volumeClass.Name] = volumeClass
		g.volumeClassByCapabilities[caps] = append(g.volumeClassByCapabilities[caps], volumeClass)
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

func (g *Generic) GetVolumeClassFor(ctx context.Context, name string, caps *ori.VolumeClassCapabilities) (*ori.VolumeClass, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	expected := getCapabilities(caps)
	if byName, ok := g.volumeClassByName[name]; ok && getCapabilities(byName.Capabilities) == expected {
		return byName, nil
	}

	if byCaps, ok := g.volumeClassByCapabilities[expected]; ok {
		switch len(byCaps) {
		case 0:
			return nil, ErrNoMatchingVolumeClass
		case 1:
			class := *byCaps[0]
			return &class, nil
		default:
			return nil, ErrAmbiguousMatchingVolumeClass
		}
	}

	return nil, ErrNoMatchingVolumeClass
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

func NewGeneric(runtime ori.VolumeRuntimeClient, opts GenericOptions) VolumeClassMapper {
	setGenericOptionsDefaults(&opts)
	return &Generic{
		synced:                    make(chan struct{}),
		volumeClassByName:         map[string]*ori.VolumeClass{},
		volumeClassByCapabilities: map[capabilities][]*ori.VolumeClass{},
		volumeRuntime:             runtime,
		relistPeriod:              opts.RelistPeriod,
	}
}
