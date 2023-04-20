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
	MachineUIDLabel       = "machinepoollet.api.onmetal.de/machine-uid"
	MachineNamespaceLabel = "machinepoollet.api.onmetal.de/machine-namespace"
	MachineNameLabel      = "machinepoollet.api.onmetal.de/machine-name"

	VolumeNameLabel = "machinepoollet.api.onmetal.de/volume-name"

	NetworkInterfaceNameLabel = "machinepoollet.api.onmetal.de/networkinterface-name"

	MachineGenerationAnnotation = "machinepoollet.api.onmetal.de/machine-generation"

	FieldOwner       = "machinepoollet.api.onmetal.de/field-owner"
	MachineFinalizer = "machinepoollet.api.onmetal.de/machine"

	// DownwardAPIPrefix is the prefix for any downward label.
	DownwardAPIPrefix = "downward-api.machinepoollet.api.onmetal.de/"
)

// DownwardAPILabel makes a downward api label name from the given name.
func DownwardAPILabel(name string) string {
	return DownwardAPIPrefix + name
}

// DownwardAPIAnnotation makes a downward api annotation name from the given name.
func DownwardAPIAnnotation(name string) string {
	return DownwardAPIPrefix + name
}
