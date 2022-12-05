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

package handler

import (
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// EnqueueRequestForName enqueues a reconcile.Request for a name without namespace.
type EnqueueRequestForName string

func (e EnqueueRequestForName) enqueue(queue workqueue.RateLimitingInterface) {
	queue.Add(reconcile.Request{NamespacedName: client.ObjectKey{Name: string(e)}})
}

// Create implements handler.EventHandler.
func (e EnqueueRequestForName) Create(_ event.CreateEvent, queue workqueue.RateLimitingInterface) {
	e.enqueue(queue)
}

// Update implements handler.EventHandler.
func (e EnqueueRequestForName) Update(_ event.UpdateEvent, queue workqueue.RateLimitingInterface) {
	e.enqueue(queue)
}

// Delete implements handler.EventHandler.
func (e EnqueueRequestForName) Delete(_ event.DeleteEvent, queue workqueue.RateLimitingInterface) {
	e.enqueue(queue)
}

// Generic implements handler.EventHandler.
func (e EnqueueRequestForName) Generic(_ event.GenericEvent, queue workqueue.RateLimitingInterface) {
	e.enqueue(queue)
}
