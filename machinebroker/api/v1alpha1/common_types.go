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

const (
	MetadataAnnotation = "machinebroker.api.onmetal.de/metadata"

	LabelsAnnotation = "machinebroker.api.onmetal.de/labels"

	AnnotationsAnnotation = "machinebroker.api.onmetal.de/annotations"

	CreatedLabel = "machinebroker.api.onmetal.de/created"
)

const (
	MachineIDLabel = "machinebroker.api.onmetal.de/machine-id"

	PurposeLabel = "machinebroker.api.onmetal.de/purpose"

	VolumeNameLabel = "machinebroker.api.onmetal.de/volume-name"

	NetworkInterfaceNameLabel = "machinebroker.api.onmetal.de/network-interface"

	IPFamilyLabel = "machinebroker.api.onmetal.de/ip-family"

	ManagerLabel = "machinebroker.api.onmetal.de/manager"

	DeviceLabel = "machinebroker.api.onmetal.de/device"
)

const (
	MachineBrokerManager = "machinebroker"

	VolumeAccessPurpose = "volume-access"

	IgnitionPurpose = "ignition"

	NetworkInterfacePurpose = "network-interface"
)
