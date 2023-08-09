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
	"github.com/gogo/protobuf/proto"
	ori "github.com/onmetal/onmetal-api/ori/apis/volume/v1alpha1"
	"github.com/onmetal/onmetal-api/poollet/orievent"
	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
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

	listener sets.Set[*listener]

	volumeClassByName         map[string]*ori.VolumeClassStatus
	volumeClassByCapabilities map[capabilities][]*ori.VolumeClassStatus

	volumeRuntime ori.VolumeRuntimeClient

	relistPeriod time.Duration
}

func (g *Generic) AddListener(handler orievent.Listener) (orievent.ListenerRegistration, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	h := &listener{Listener: handler}

	g.listener.Insert(h)
	return &listenerRegistration{
		vcm:      g,
		listener: h,
	}, nil
}

func (g *Generic) RemoveListener(listener orievent.ListenerRegistration) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	reg, ok := listener.(*listenerRegistration)
	if !ok {
		return fmt.Errorf("invalid listener registration")
	}

	g.listener.Delete(reg.listener)

	return nil
}

func shouldNotify(oldVolumeClassByName map[string]*ori.VolumeClassStatus, class *ori.VolumeClassStatus) bool {
	oldVolumeClass, ok := oldVolumeClassByName[class.VolumeClass.Name]
	if !ok {
		return true
	}

	return proto.Equal(class, oldVolumeClass)
}

func (g *Generic) relist(ctx context.Context, log logr.Logger) error {
	log.V(1).Info("Relisting volume classes")
	res, err := g.volumeRuntime.Status(ctx, &ori.StatusRequest{})
	if err != nil {
		return fmt.Errorf("error listing volume classes: %w", err)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	oldVolumeClassByName := maps.Clone(g.volumeClassByName)

	maps.Clear(g.volumeClassByName)
	maps.Clear(g.volumeClassByCapabilities)

	var notify bool
	for _, volumeClassStatus := range res.VolumeClassStatus {
		volumeClass := volumeClassStatus.GetVolumeClass()
		notify = notify || shouldNotify(oldVolumeClassByName, volumeClassStatus)

		caps := getCapabilities(volumeClass.Capabilities)
		g.volumeClassByName[volumeClass.Name] = volumeClassStatus
		g.volumeClassByCapabilities[caps] = append(g.volumeClassByCapabilities[caps], volumeClassStatus)
	}

	if notify {
		log.V(1).Info("Notify")
		for _, n := range g.listener.UnsortedList() {
			n.Enqueue()
		}
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

func (g *Generic) GetVolumeClassFor(ctx context.Context, name string, caps *ori.VolumeClassCapabilities) (*ori.VolumeClass, *resource.Quantity, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	expected := getCapabilities(caps)
	if byName, ok := g.volumeClassByName[name]; ok && getCapabilities(byName.VolumeClass.Capabilities) == expected {
		return byName.VolumeClass, resource.NewQuantity(byName.Quantity, resource.BinarySI), nil
	}

	if byCaps, ok := g.volumeClassByCapabilities[expected]; ok {
		switch len(byCaps) {
		case 0:
			return nil, nil, ErrNoMatchingVolumeClass
		case 1:
			classStatus := *byCaps[0]
			return classStatus.VolumeClass, resource.NewQuantity(classStatus.Quantity, resource.BinarySI), nil
		default:
			return nil, nil, ErrAmbiguousMatchingVolumeClass
		}
	}

	return nil, nil, ErrNoMatchingVolumeClass
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
		volumeClassByName:         map[string]*ori.VolumeClassStatus{},
		volumeClassByCapabilities: map[capabilities][]*ori.VolumeClassStatus{},
		listener:                  sets.New[*listener](),
		volumeRuntime:             runtime,
		relistPeriod:              opts.RelistPeriod,
	}
}

type listener struct {
	orievent.Listener
}

type listenerRegistration struct {
	vcm      *Generic
	listener *listener
}
