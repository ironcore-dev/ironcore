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

package account

import (
	corev1alpha1 "github.com/onmetal/onmetal-api/apis/core/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type AccountWebhook struct {
	corev1alpha1.Account
}

func (r *AccountWebhook) SetupWebhookWithManager(mgr manager.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&corev1alpha1.Account{}).
		Complete()
}
