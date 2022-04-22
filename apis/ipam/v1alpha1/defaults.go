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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_PrefixSpec(spec *PrefixSpec) {
	if string(spec.IPFamily) == "" && spec.Prefix.IsValid() && spec.PrefixLength == 0 {
		switch {
		case spec.Prefix.IP().Is4():
			spec.IPFamily = corev1.IPv4Protocol
		case spec.Prefix.IP().Is6():
			spec.IPFamily = corev1.IPv6Protocol
		}
	}
}

func SetDefaults_PrefixAllocationSpec(spec *PrefixAllocationSpec) {
	if string(spec.IPFamily) == "" && spec.Prefix.IsValid() && spec.PrefixLength == 0 {
		switch {
		case spec.Prefix.IP().Is4():
			spec.IPFamily = corev1.IPv4Protocol
		case spec.Prefix.IP().Is6():
			spec.IPFamily = corev1.IPv6Protocol
		}
	}
}
