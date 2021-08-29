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

package trigger

import (
	"github.com/onmetal/onmetal-api/pkg/utils"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// ReconcilationTrigger is used to register controllers
// to be externally triggerable for a dedicated object.
// given by its object id.
type ReconcilationTrigger interface {
	// RegisterControllerFor registers a controller to receice
	// reconcilation requests for a dedicated group kind, if
	// the Trigger method is called for such an object id.
	//
	// ATTENTION: Because a controller is always responsible for ONLY
	//   one object type, but is does not provide a method to
	//   get this type, this is not validated during the registration.
	//   Therefore this method MUST only be called for a controller
	//   if it is handling this type of object.
	RegisterControllerFor(gk schema.GroupKind, c controller.Controller)

	// Trigger triggers all controllers registered for the GroupKind of the given id
	Trigger(id utils.ObjectId)
	// TriggerAll triggers all controllers registered for the GroupKind of the given ids
	TriggerAll(ids utils.ObjectIds)
}
