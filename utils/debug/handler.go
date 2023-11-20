// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package debug

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

type loggingQueue struct {
	mu sync.RWMutex

	done bool
	log  logr.Logger
	workqueue.RateLimitingInterface
}

func newLoggingQueue(log logr.Logger, queue workqueue.RateLimitingInterface) *loggingQueue {
	return &loggingQueue{log: log, RateLimitingInterface: queue}
}

func (q *loggingQueue) Add(item interface{}) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	q.log.Info("Add", "Item", item, "Done", q.done)
	q.RateLimitingInterface.Add(item)
}

func (q *loggingQueue) AddRateLimited(item interface{}) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	q.log.Info("AddRateLimited", "Item", item, "Done", q.done)
	q.RateLimitingInterface.AddRateLimited(item)
}

func (q *loggingQueue) AddAfter(item interface{}, duration time.Duration) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	q.log.Info("AddAfter", "Item", item, "Duration", duration, "Done", q.done)
	q.RateLimitingInterface.AddAfter(item, duration)
}

func (q *loggingQueue) Finish() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.done = true
}

type debugHandler struct {
	log         logr.Logger
	handler     handler.EventHandler
	objectValue func(client.Object) any
}

func (d *debugHandler) Create(ctx context.Context, evt event.CreateEvent, queue workqueue.RateLimitingInterface) {
	log := d.log.WithValues("Event", "Create", "Object", d.objectValue(evt.Object))
	log.Info("Handling Event")

	lQueue := newLoggingQueue(log.WithName("Queue"), queue)
	defer lQueue.Finish()

	d.handler.Create(ctx, evt, lQueue)
}

func (d *debugHandler) Update(ctx context.Context, evt event.UpdateEvent, queue workqueue.RateLimitingInterface) {
	log := d.log.WithValues("Event", "Update", "ObjectOld", d.objectValue(evt.ObjectOld), "ObjectNew", d.objectValue(evt.ObjectNew))
	log.Info("Handling Event")

	lQueue := newLoggingQueue(log.WithName("Queue"), queue)
	defer lQueue.Finish()

	d.handler.Update(ctx, evt, lQueue)
}

func (d *debugHandler) Delete(ctx context.Context, evt event.DeleteEvent, queue workqueue.RateLimitingInterface) {
	log := d.log.WithValues("Event", "Delete", "Object", d.objectValue(evt.Object))
	log.Info("Handling Event")

	lQueue := newLoggingQueue(log.WithName("Queue"), queue)
	defer lQueue.Finish()

	d.handler.Delete(ctx, evt, lQueue)
}

func (d *debugHandler) Generic(ctx context.Context, evt event.GenericEvent, queue workqueue.RateLimitingInterface) {
	log := d.log.WithValues("Event", "Generic", "Object", d.objectValue(evt.Object))
	log.Info("Handling Event")

	lQueue := newLoggingQueue(log.WithName("Queue"), queue)
	defer lQueue.Finish()

	d.handler.Generic(ctx, evt, lQueue)
}

// Handler allows debugging a handler.EventHandler by wrapping it and logging each action it does.
//
// Caution: This has a heavy toll on runtime performance and should *not* be used in production code.
// Use only for debugging handlers and remove once done.
func Handler(name string, handler handler.EventHandler, opts ...HandlerOption) handler.EventHandler {
	o := (&HandlerOptions{}).ApplyOptions(opts)
	setHandlerOptionsDefaults(o)

	return &debugHandler{
		log:         o.Log.WithName(name),
		handler:     handler,
		objectValue: o.ObjectValue,
	}
}
