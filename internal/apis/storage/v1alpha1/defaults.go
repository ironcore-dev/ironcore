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

package v1alpha1

import (
	"github.com/onmetal/onmetal-api/api/storage/v1alpha1"
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
