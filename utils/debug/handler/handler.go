// Copyright 2023 OnMetal authors
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

package handler

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

var log = ctrl.Log.WithName("debug").WithName("eventhandler")

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

// HandlerOptions are options for construction a debug handler.
type HandlerOptions struct {
	// Log is the logger to use. If unspecified, the debug package logger will be used.
	Log logr.Logger

	// ObjectValue controls how an object will be represented as in the log values.
	ObjectValue func(client.Object) any
}

func (o *HandlerOptions) ApplyToHandler(o2 *HandlerOptions) {
	if o.Log.GetSink() != nil {
		o2.Log = o.Log
	}
	if o.ObjectValue != nil {
		o2.ObjectValue = DefaultObjectValue
	}
}

func (o *HandlerOptions) ApplyOptions(opts []Option) *HandlerOptions {
	for _, opt := range opts {
		opt.ApplyToHandler(o)
	}
	return o
}

func setHandlerOptionsDefaults(o *HandlerOptions) {
	if o.Log.GetSink() == nil {
		o.Log = log
	}
	if o.ObjectValue == nil {
		o.ObjectValue = DefaultObjectValue
	}
}

// DefaultObjectValue provides object logging values by using klog.KObj.
func DefaultObjectValue(obj client.Object) any {
	return klog.KObj(obj)
}

type Option interface {
	ApplyToHandler(o *HandlerOptions)
}

// WithLog specifies the logger to use.
type WithLog struct {
	Log logr.Logger
}

func (w WithLog) ApplyToHandler(o *HandlerOptions) {
	o.Log = w.Log
}

// WithObjectValue specifies the function to log an client.Object's value with.
type WithObjectValue func(obj client.Object) any

func (w WithObjectValue) ApplyToHandler(o *HandlerOptions) {
	o.ObjectValue = w
}

// Handler allows debugging a handler.EventHandler by wrapping it and logging each action it does.
//
// Caution: This has a heavy toll on runtime performance and should *not* be used in production code.
// Use only for debugging handlers and remove once done.
func Handler(name string, handler handler.EventHandler, opts ...Option) handler.EventHandler {
	o := (&HandlerOptions{}).ApplyOptions(opts)
	setHandlerOptionsDefaults(o)

	return &debugHandler{
		log:         o.Log.WithName(name),
		handler:     handler,
		objectValue: o.ObjectValue,
	}
}
