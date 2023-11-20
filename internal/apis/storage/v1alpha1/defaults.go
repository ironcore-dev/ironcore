// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_VolumeStatus(status *v1alpha1.VolumeStatus) {
	if status.State == "" {
		status.State = v1alpha1.VolumeStatePending
	}
}

func SetDefaults_BucketStatus(status *v1alpha1.BucketStatus) {
	if status.State == "" {
		status.State = v1alpha1.BucketStatePending
	}
}

func SetDefaults_VolumeClass(volumeClass *v1alpha1.VolumeClass) {
	if volumeClass.ResizePolicy == "" {
		volumeClass.ResizePolicy = v1alpha1.ResizePolicyStatic
	}
}
