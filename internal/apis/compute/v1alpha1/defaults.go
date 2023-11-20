// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
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

func SetDefaults_NetworkInterfaceStatus(status *v1alpha1.NetworkInterfaceStatus) {
	if status.State == "" {
		status.State = v1alpha1.NetworkInterfaceStatePending
	}
}

func SetDefaults_MachineStatus(status *v1alpha1.MachineStatus) {
	if status.State == "" {
		status.State = v1alpha1.MachineStatePending
	}
}

func SetDefaults_MachineSpec(spec *v1alpha1.MachineSpec) {
	if spec.Power == "" {
		spec.Power = v1alpha1.PowerOn
	}
}
