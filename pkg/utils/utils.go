/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// AssureFinalizer ensures that a finalizer is on a given runtime object
func AssureFinalizer(ctx context.Context, client client.Client, finalizerName string, object client.Object) error {
	if !ContainsString(object.GetFinalizers(), finalizerName) {
		controllerutil.AddFinalizer(object, finalizerName)
		return client.Update(ctx, object)
	}
	return nil
}

// AssureFinalizerRemoved ensures that a finalizer does not exist anymore for a given runtime object
func AssureFinalizerRemoved(ctx context.Context, client client.Client, finalizerName string, object client.Object) error {
	if ContainsString(object.GetFinalizers(), finalizerName) {
		controllerutil.RemoveFinalizer(object, finalizerName)
		return client.Update(ctx, object)
	}
	return nil
}

// AssureDeleting ensures that the object is in deletion mode
func AssureDeleting(ctx context.Context, clt client.Client, object client.Object) error {
	if !object.GetDeletionTimestamp().IsZero() {
		return nil
	}
	return client.IgnoreNotFound(clt.Delete(ctx, object, client.PropagationPolicy(metav1.DeletePropagationBackground)))
}

// ContainsString is a helper functions to check and remove string from a slice of strings.
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
