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

package predicates

import (
	"strings"

	"github.com/go-logr/logr"
	"github.com/onmetal/onmetal-api/utils/annotations"
	"github.com/onmetal/onmetal-api/utils/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ResourceHasFilterLabel returns a predicate that returns true only if the provided resource contains
// a label with the WatchLabel key and the configured label value exactly.
func ResourceHasFilterLabel(logger logr.Logger, labelValue string) predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return processIfLabelMatch(logger.WithValues("predicate", "ResourceHasFilterLabel", "eventType", "update"), e.ObjectNew, labelValue)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return processIfLabelMatch(logger.WithValues("predicate", "ResourceHasFilterLabel", "eventType", "create"), e.Object, labelValue)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return processIfLabelMatch(logger.WithValues("predicate", "ResourceHasFilterLabel", "eventType", "delete"), e.Object, labelValue)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return processIfLabelMatch(logger.WithValues("predicate", "ResourceHasFilterLabel", "eventType", "generic"), e.Object, labelValue)
		},
	}
}

func processIfLabelMatch(logger logr.Logger, obj client.Object, labelValue string) bool {
	// Return early if no labelValue was set.
	if labelValue == "" {
		return true
	}

	kind := strings.ToLower(obj.GetObjectKind().GroupVersionKind().Kind)
	log := logger.WithValues("namespace", obj.GetNamespace(), kind, obj.GetName())
	if labels.HasWatchLabel(obj, labelValue) {
		log.V(6).Info("Resource matches label, will attempt to map resource")
		return true
	}
	log.V(4).Info("Resource does not match label, will not attempt to map resource")
	return false
}

// ResourceIsNotExternallyManaged returns a predicate that returns true only if the resource does not contain
// the externally managed annotation.
func ResourceIsNotExternallyManaged(logger logr.Logger) predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return processIfNotExternallyManaged(logger.WithValues("predicate", "ResourceIsNotExternallyManaged", "eventType", "update"), e.ObjectNew)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return processIfNotExternallyManaged(logger.WithValues("predicate", "ResourceIsNotExternallyManaged", "eventType", "create"), e.Object)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return processIfNotExternallyManaged(logger.WithValues("predicate", "ResourceIsNotExternallyManaged", "eventType", "delete"), e.Object)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return processIfNotExternallyManaged(logger.WithValues("predicate", "ResourceIsNotExternallyManaged", "eventType", "generic"), e.Object)
		},
	}
}

func processIfNotExternallyManaged(logger logr.Logger, obj client.Object) bool {
	kind := strings.ToLower(obj.GetObjectKind().GroupVersionKind().Kind)
	log := logger.WithValues("namespace", obj.GetNamespace(), kind, obj.GetName())
	if annotations.IsExternallyManaged(obj) {
		log.V(4).Info("Resource is externally managed, will not attempt to map resource")
		return false
	}
	log.V(6).Info("Resource is managed, will attempt to map resource")
	return true
}
