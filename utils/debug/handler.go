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
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type loggingQueue struct {
	mu sync.RWMutex

	done  bool
	log   logr.Logger
	queue workqueue.TypedRateLimitingInterface[reconcile.Request]
}

func newLoggingQueue(log logr.Logger, queue workqueue.TypedRateLimitingInterface[reconcile.Request]) *loggingQueue {
	return &loggingQueue{log: log, queue: queue}
}

func (q *loggingQueue) Add(item reconcile.Request) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	q.log.Info("Add", "Item", item, "Done", q.done)
	q.queue.Add(item)
}

func (q *loggingQueue) AddRateLimited(item reconcile.Request) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	q.log.Info("AddRateLimited", "Item", item, "Done", q.done)
	q.queue.AddRateLimited(item)
}

func (q *loggingQueue) AddAfter(item reconcile.Request, duration time.Duration) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	q.log.Info("AddAfter", "Item", item, "Duration", duration, "Done", q.done)
	q.queue.AddAfter(item, duration)
}

func (q *loggingQueue) Done(item reconcile.Request) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.done = true
	q.queue.Done(item)
}

func (q *loggingQueue) Forget(item reconcile.Request) {
	q.queue.Forget(item)
}

func (q *loggingQueue) NumRequeues(item reconcile.Request) int {
	return q.queue.NumRequeues(item)
}

func (q *loggingQueue) Get() (reconcile.Request, bool) {
	item, shutdown := q.queue.Get()
	if shutdown {
		return reconcile.Request{}, false
	}
	return item, true
}

func (q *loggingQueue) Len() int {
	return q.queue.Len()
}

func (q *loggingQueue) ShutDown() {
	q.queue.ShutDown()
}

func (q *loggingQueue) ShutDownWithDrain() {
	q.queue.ShutDownWithDrain()
}

func (q *loggingQueue) ShuttingDown() bool {
	return q.queue.ShuttingDown()
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

func (d *debugHandler) Create(ctx context.Context, evt event.CreateEvent, queue workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	log := d.log.WithValues("Event", "Create", "Object", d.objectValue(evt.Object))
	log.Info("Handling Event")

	lQueue := newLoggingQueue(log.WithName("Queue"), queue)
	defer lQueue.Finish()

	d.handler.Create(ctx, evt, lQueue)
}

func (d *debugHandler) Update(ctx context.Context, evt event.UpdateEvent, queue workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	log := d.log.WithValues("Event", "Update", "ObjectOld", d.objectValue(evt.ObjectOld), "ObjectNew", d.objectValue(evt.ObjectNew))
	log.Info("Handling Event")

	lQueue := newLoggingQueue(log.WithName("Queue"), queue)
	defer lQueue.Finish()

	d.handler.Update(ctx, evt, lQueue)
}

func (d *debugHandler) Delete(ctx context.Context, evt event.DeleteEvent, queue workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	log := d.log.WithValues("Event", "Delete", "Object", d.objectValue(evt.Object))
	log.Info("Handling Event")

	lQueue := newLoggingQueue(log.WithName("Queue"), queue)
	defer lQueue.Finish()

	d.handler.Delete(ctx, evt, lQueue)
}

func (d *debugHandler) Generic(ctx context.Context, evt event.GenericEvent, queue workqueue.TypedRateLimitingInterface[reconcile.Request]) {
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
