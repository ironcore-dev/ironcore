// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	MachineUIDLabel       = "machinepoollet.ironcore.dev/machine-uid"
	MachineNamespaceLabel = "machinepoollet.ironcore.dev/machine-namespace"
	MachineNameLabel      = "machinepoollet.ironcore.dev/machine-name"

	MachineGenerationAnnotation    = "machinepoollet.ironcore.dev/machine-generation"
	IRIMachineGenerationAnnotation = "machinepoollet.ironcore.dev/irimachine-generation"

	NetworkInterfaceMappingAnnotation = "machinepoollet.ironcore.dev/networkinterfacemapping"

	FieldOwner       = "machinepoollet.ironcore.dev/field-owner"
	MachineFinalizer = "machinepoollet.ironcore.dev/machine"

	ReservationUIDLabel       = "machinepoollet.ironcore.dev/reservation-uid"
	ReservationNamespaceLabel = "machinepoollet.ironcore.dev/reservation-namespace"
	ReservationNameLabel      = "machinepoollet.ironcore.dev/reservation-name"

	ReservationFinalizerBase = "machinepoollet.ironcore.dev/reservation-"

	ReservationGenerationAnnotation    = "machinepoollet.ironcore.dev/reservation-generation"
	IRIReservationGenerationAnnotation = "machinepoollet.ironcore.dev/irireservation-generation"

	// DownwardAPIPrefix is the prefix for any downward label.
	DownwardAPIPrefix = "downward-api.machinepoollet.ironcore.dev/"
)

func ReservationFinalizer(poolName string) string {
	return ReservationFinalizerBase + poolName
}

// DownwardAPILabel makes a downward api label name from the given name.
func DownwardAPILabel(name string) string {
	return DownwardAPIPrefix + name
}

// DownwardAPIAnnotation makes a downward api annotation name from the given name.
func DownwardAPIAnnotation(name string) string {
	return DownwardAPIPrefix + name
}

// EncodeNetworkInterfaceMapping encodes the given network interface mapping to be used as an annotation.
func EncodeNetworkInterfaceMapping(nicMapping map[string]ObjectUIDRef) (string, error) {
	data, err := json.Marshal(nicMapping)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func DecodeNetworkInterfaceMapping(nicMappingString string) (map[string]ObjectUIDRef, error) {
	var nicMapping map[string]ObjectUIDRef
	if err := json.Unmarshal([]byte(nicMappingString), &nicMapping); err != nil {
		return nil, err
	}

	return nicMapping, nil
}

// ObjectUIDRef is a name-uid-reference to an object.
type ObjectUIDRef struct {
	Name string    `json:"name"`
	UID  types.UID `json:"uid"`
}

func ObjUID(obj client.Object) ObjectUIDRef {
	return ObjectUIDRef{
		Name: obj.GetName(),
		UID:  obj.GetUID(),
	}
}
