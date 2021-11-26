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

package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var machinelog = logf.Log.WithName("machine-resource")

// "verbs=update" since only update validation is enabled
//+kubebuilder:webhook:path=/validate-compute-onmetal-de-v1alpha1-machine,mutating=false,failurePolicy=fail,sideEffects=None,groups=compute.onmetal.de,resources=machines,verbs=update,versions=v1alpha1,name=machine.v1alpha1.compute.onmetal.de,admissionReviewVersions={v1,v1beta1}

// SetupWebhookWithManager creates a new webhook which will be started by mgr
func (m *Machine) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(m).
		Complete()
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
// Satisfy the interface, thus empty
func (m *Machine) ValidateCreate() error {
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
// Satisfy the interface, thus empty
func (r *Machine) ValidateDelete() error {
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (m *Machine) ValidateUpdate(old runtime.Object) error {
	machinelog.Info("validate update", "name", m.Name)

	newName := m.Spec.MachinePool.Name
	oldName := old.(*Machine).Spec.MachineClass.Name
	if newName != oldName {
		path := field.NewPath("spec", "machinePool", "name")
		fieldInvalid := field.Invalid(path, newName, "machinepool should be immutable")
		var machineGK = schema.GroupKind{
			Group: GroupVersion.Group,
			Kind:  "Machine",
		}

		return apierrors.NewInvalid(machineGK, m.Name, field.ErrorList{fieldInvalid})
	}
	return nil
}
