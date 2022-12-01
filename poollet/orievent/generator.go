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

package orievent

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	"github.com/gogo/protobuf/proto"
	"github.com/onmetal/controller-utils/set"
	orimeta "github.com/onmetal/onmetal-api/ori/apis/meta/v1alpha1"
	"k8s.io/apiserver/pkg/server/healthz"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Generator[O orimeta.Object] interface {
	Source[O]
	healthz.HealthChecker
	manager.Runnable
}

type event[O orimeta.Object] struct {
	Create  *CreateEvent[O]
	Update  *UpdateEvent[O]
	Delete  *DeleteEvent[O]
	Generic *GenericEvent[O]
}

type oldNewMapEntry[O orimeta.Object] struct {
	Old     *O
	Current *O
}

type oldNewMap[O orimeta.Object] map[string]*oldNewMapEntry[O]

func (m oldNewMap[O]) id(obj O) string {
	return obj.GetMetadata().GetId()
}

func (m oldNewMap[O]) setCurrent(current []O) {
	for _, v := range m {
		v.Current = nil
	}

	for _, item := range current {
		item := item
		id := m.id(item)
		if r, ok := m[id]; ok {
			r.Current = &item
		} else {
			m[id] = &oldNewMapEntry[O]{
				Current: &item,
			}
		}
	}
}

func (m oldNewMap[O]) getCurrent(id string) (O, bool) {
	r, ok := m[id]
	if ok && r.Current != nil {
		return *r.Current, true
	}
	var zero O
	return zero, false
}

func (m oldNewMap[O]) getOld(id string) (O, bool) {
	r, ok := m[id]
	if ok && r.Old != nil {
		return *r.Old, true
	}
	var zero O
	return zero, false
}

func (m oldNewMap[O]) update(id string) {
	r, ok := m[id]
	if !ok {
		return
	}

	if r.Current == nil {
		delete(m, id)
		return
	}

	r.Old = r.Current
	r.Current = nil
}

type handler[O orimeta.Object] struct {
	Handler[O]
}

type generator[O orimeta.Object] struct {
	mu sync.RWMutex

	eventChannel chan *event[O]

	handlers set.Set[*handler[O]]

	// relistPeriod is the period for relisting.
	relistPeriod time.Duration
	// relistThreshold is the maximum threshold between two relists to become unhealthy.
	relistThreshold time.Duration
	// relistTime is the last time a relist happened.
	relistTime atomic.Pointer[time.Time]
	// firstListTime is the first time a relist happened.
	firstListTime time.Time

	items oldNewMap[O]

	list func(ctx context.Context) ([]O, error)
}

type GeneratorOptions struct {
	ChannelCapacity int
	RelistPeriod    time.Duration
	RelistThreshold time.Duration
}

func setGeneratorOptionsDefaults(o *GeneratorOptions) {
	if o.ChannelCapacity == 0 {
		o.ChannelCapacity = 1024
	}
	if o.RelistPeriod <= 0 {
		o.RelistPeriod = 1 * time.Second
	}
	if o.RelistThreshold <= 0 {
		o.RelistThreshold = 3 * time.Minute
	}
}

func NewGenerator[O orimeta.Object](list func(ctx context.Context) ([]O, error), opts GeneratorOptions) Generator[O] {
	setGeneratorOptionsDefaults(&opts)

	return &generator[O]{
		eventChannel:    make(chan *event[O], opts.ChannelCapacity),
		relistPeriod:    opts.RelistPeriod,
		relistThreshold: opts.RelistThreshold,
		relistTime:      atomic.Pointer[time.Time]{},
		firstListTime:   time.Time{},
		items:           make(oldNewMap[O]),
		list:            list,
		handlers:        set.New[*handler[O]](),
	}
}

func (g *generator[O]) Name() string {
	var zero O
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return fmt.Sprintf("%s-event-generator", strings.ToLower(t.Name()))
}

func (g *generator[O]) Check(_ *http.Request) error {
	relistTime := g.relistTime.Load()
	if relistTime == nil {
		return fmt.Errorf("mleg did not relist yet")
	}

	elapsed := time.Since(*relistTime)
	if elapsed > g.relistThreshold {
		return fmt.Errorf("mleg was last seen active %v ago, threshold is %v", elapsed, g.relistThreshold)
	}
	return nil
}

func (g *generator[O]) readHandlers() []*handler[O] {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.handlers.Slice()
}

func (g *generator[O]) Start(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("event-generator")

	go func() {
		for evt := range g.eventChannel {
			handlers := g.readHandlers()

			for _, handler := range handlers {
				switch {
				case evt.Create != nil:
					handler.Create(*evt.Create)
				case evt.Update != nil:
					handler.Update(*evt.Update)
				case evt.Delete != nil:
					handler.Delete(*evt.Delete)
				case evt.Generic != nil:
					handler.Generic(*evt.Generic)
				}
			}
		}
	}()

	go func() {
		defer close(g.eventChannel)

		t := time.NewTicker(g.relistPeriod)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if err := g.relist(ctx, log); err != nil {
					log.Error(err, "Error relisting")
				}
			}
		}
	}()

	return nil
}

func (g *generator[O]) relist(ctx context.Context, log logr.Logger) error {
	timestamp := time.Now()
	objects, err := g.list(ctx)
	if err != nil {
		return fmt.Errorf("error listing: %w", err)
	}
	g.relistTime.Store(&timestamp)
	g.firstListTime = timestamp

	g.items.setCurrent(objects)

	eventsByKey := make(map[string][]*event[O])
	for key := range g.items {
		itemOld, oldOK := g.items.getOld(key)
		itemNew, newOK := g.items.getCurrent(key)
		switch {
		case !oldOK && newOK:
			createdAt := time.Unix(0, itemNew.GetMetadata().CreatedAt)
			if createdAt.Before(g.firstListTime) {
				eventsByKey[key] = []*event[O]{{Create: &CreateEvent[O]{Object: itemNew}}}
			} else {
				eventsByKey[key] = []*event[O]{{Generic: &GenericEvent[O]{Object: itemNew}}}
			}
		case oldOK && !newOK:
			eventsByKey[key] = []*event[O]{{Delete: &DeleteEvent[O]{Object: itemOld}}}
		case oldOK && newOK:
			if !proto.Equal(itemOld, itemNew) {
				eventsByKey[key] = []*event[O]{{Update: &UpdateEvent[O]{ObjectOld: itemOld, ObjectNew: itemNew}}}
			}
		}
	}

	for machineID, events := range eventsByKey {
		g.items.update(machineID)
		for i := range events {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case g.eventChannel <- events[i]:
			default:
				log.Info("Event channel is full, discarding event", "MachineID", machineID)
			}
		}
	}

	return nil
}

type handlerRegistration[O orimeta.Object] struct {
	generator *generator[O]
	handler   *handler[O]
}

func (r *handlerRegistration[O]) Remove() error {
	r.generator.mu.Lock()
	defer r.generator.mu.Unlock()

	r.generator.handlers.Delete(r.handler)
	return nil
}

func (g *generator[O]) AddHandler(hdl Handler[O]) (HandlerRegistration, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	h := &handler[O]{Handler: hdl}

	g.handlers.Insert(h)
	return &handlerRegistration[O]{
		generator: g,
		handler:   h,
	}, nil
}
