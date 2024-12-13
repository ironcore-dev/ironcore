// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// EnqueueRequestForName enqueues a reconcile.Request for a name without namespace.
type EnqueueRequestForName string

func (e EnqueueRequestForName) enqueue(queue workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	queue.Add(reconcile.Request{NamespacedName: client.ObjectKey{Name: string(e)}})
}

// Create implements handler.EventHandler.
func (e EnqueueRequestForName) Create(_ event.CreateEvent, queue workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	e.enqueue(queue)
}

// Update implements handler.EventHandler.
func (e EnqueueRequestForName) Update(_ event.UpdateEvent, queue workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	e.enqueue(queue)
}

// Delete implements handler.EventHandler.
func (e EnqueueRequestForName) Delete(_ event.DeleteEvent, queue workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	e.enqueue(queue)
}

// Generic implements handler.EventHandler.
func (e EnqueueRequestForName) Generic(_ event.GenericEvent, queue workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	e.enqueue(queue)
}
