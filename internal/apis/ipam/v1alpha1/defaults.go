// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_PrefixSpec(spec *v1alpha1.PrefixSpec) {
	if string(spec.IPFamily) == "" && spec.Prefix.IsValid() && spec.PrefixLength == 0 {
		switch {
		case spec.Prefix.IP().Is4():
			spec.IPFamily = corev1.IPv4Protocol
		case spec.Prefix.IP().Is6():
			spec.IPFamily = corev1.IPv6Protocol
		}
	}
}

func SetDefaults_PrefixAllocationSpec(spec *v1alpha1.PrefixAllocationSpec) {
	if string(spec.IPFamily) == "" && spec.Prefix.IsValid() && spec.PrefixLength == 0 {
		switch {
		case spec.Prefix.IP().Is4():
			spec.IPFamily = corev1.IPv4Protocol
		case spec.Prefix.IP().Is6():
			spec.IPFamily = corev1.IPv6Protocol
		}
	}
}

func SetDefaults_PrefixAllocationStatus(status *v1alpha1.PrefixAllocationStatus) {
	if status.Phase == "" {
		status.Phase = v1alpha1.PrefixAllocationPhasePending
	}
}

func SetDefaults_PrefixStatus(status *v1alpha1.PrefixStatus) {
	if status.Phase == "" {
		status.Phase = v1alpha1.PrefixPhasePending
	}
}
