/*
 * Copyright (c) 2021 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"context"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func HasFinalizer(object client.Object, finalizerName string) bool {
	return ContainsString(object.GetFinalizers(), finalizerName)
}

// CheckAndAssureFinalizer ensures that a finalizer is on a given runtime object
// Returns false if the finalizer has been added.
func CheckAndAssureFinalizer(ctx context.Context, client client.Client, finalizerName string, object client.Object) (bool, error) {
	log := logr.FromContextOrDiscard(ctx)
	if !ContainsString(object.GetFinalizers(), finalizerName) {
		log.Info("setting finalizer", "finalizer", finalizerName)
		controllerutil.AddFinalizer(object, finalizerName)
		return false, client.Update(ctx, object)
	}
	return true, nil
}

// AssureFinalizer ensures that a finalizer is on a given runtime object
func AssureFinalizer(ctx context.Context, client client.Client, finalizerName string, object client.Object) error {
	_, err := CheckAndAssureFinalizer(ctx, client, finalizerName, object)
	return err
}

// AssureFinalizerRemoved ensures that a finalizer does not exist anymore for a given runtime object
func AssureFinalizerRemoved(ctx context.Context, client client.Client, finalizerName string, object client.Object) error {
	log := logr.FromContextOrDiscard(ctx)
	if ContainsString(object.GetFinalizers(), finalizerName) {
		log.Info("removing finalizer", "finalizer", finalizerName)
		controllerutil.RemoveFinalizer(object, finalizerName)
		return client.Update(ctx, object)
	}
	return nil
}
