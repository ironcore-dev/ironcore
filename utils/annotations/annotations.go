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

package annotations

import (
	"time"

	"github.com/onmetal/controller-utils/metautils"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func HasReconcileAnnotation(o metav1.Object) bool {
	return metautils.HasAnnotation(o, commonv1alpha1.ReconcileRequestAnnotation)
}

func SetReconcileAnnotation(o metav1.Object) {
	metautils.SetAnnotation(o, commonv1alpha1.ReconcileRequestAnnotation, time.Now().Format(time.RFC3339Nano))
}

func RemoveReconcileAnnotation(o metav1.Object) {
	metautils.DeleteAnnotation(o, commonv1alpha1.ReconcileRequestAnnotation)
}

func IsExternallyManaged(o metav1.Object) bool {
	return metautils.HasAnnotation(o, commonv1alpha1.ManagedByAnnotation)
}

func SetExternallyMangedBy(o metav1.Object, manager string) {
	metautils.SetAnnotation(o, commonv1alpha1.ManagedByAnnotation, manager)
}

func RemoveExternallyManaged(o metav1.Object) {
	metautils.DeleteAnnotation(o, commonv1alpha1.ManagedByAnnotation)
}
