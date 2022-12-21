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
	"fmt"
	"sort"

	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	LabelsAnnotation = "machinebroker.api.onmetal.de/labels"

	AnnotationsAnnotation = "machinebroker.api.onmetal.de/annotations"

	DependentsAnnotation = "machinebrokerlet.api.onmetal.de/dependents"
)

const (
	PurposeLabel = "machinebroker.api.onmetal.de/purpose"

	ManagerLabel = "machinebroker.api.onmetal.de/manager"

	CreatedLabel = "machinebroker.api.onmetal.de/created"

	NetworkHandleLabel = "machinebrokerlet.api.onmetal.de/network-handle"

	PrefixLabel = "machinebrokerlet.api.onmetal.de/prefix"

	IPLabel = "machinebrokerlet.api.onmetal.de/ip"
)

const (
	MachineBrokerManager = "machinebroker"

	VolumeAccessPurpose = "volume-access"

	IgnitionPurpose = "ignition"

	NetworkInterfacePurpose = "network-interface"
)

type LoadBalancerPort struct {
	Protocol corev1.Protocol
	Port     int32
	EndPort  int32
}

func (p LoadBalancerPort) Key() string {
	return fmt.Sprintf("%s:%d-%d", p.Protocol, p.Port, p.EndPort)
}

func LoadBalancerPortsKey(ports []LoadBalancerPort) string {
	portKeys := make([]string, len(ports))
	for i, port := range ports {
		portKeys[i] = port.Key()
	}
	sort.Strings(portKeys)
	return fmt.Sprintf("%v", portKeys)
}

type LoadBalancerTarget struct {
	IP    commonv1alpha1.IP
	Ports []LoadBalancerPort
}

func (t LoadBalancerTarget) Key() string {
	portKeys := LoadBalancerPortsKey(t.Ports)
	return fmt.Sprintf("%s%s", t.IP, portKeys)
}

type LoadBalancer struct {
	NetworkHandle string
	IP            commonv1alpha1.IP
	Ports         []LoadBalancerPort
	Destinations  []string
}
