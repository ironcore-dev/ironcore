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

package v1alpha1

const (
	// WatchLabel is a label that can be applied to any onmetal resource.
	//
	// Provider controllers that allow for selective reconciliation may check this label and proceed
	// with reconciliation of the object only if this label and a configured value are present.
	WatchLabel = "common.api.onmetal.de/watch-filter"

	// ReconcileRequestAnnotation is an annotation that requested a reconciliation at a specific time.
	ReconcileRequestAnnotation = "reconcile.common.api.onmetal.de/requested-at"

	// ManagedByAnnotation is an annotation that can be applied to resources to signify that
	// some external system is managing the resource.
	ManagedByAnnotation = "common.api.onmetal.de/managed-by"
)
