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

package apiutils

import (
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func ConvertNetworkingLoadBalancerPortToLoadBalancerPort(port networkingv1alpha1.LoadBalancerPort) machinebrokerv1alpha1.LoadBalancerPort {
	protocol := port.Protocol
	if protocol == nil {
		tcpProtocol := corev1.ProtocolTCP
		protocol = &tcpProtocol
	}

	endPort := port.EndPort
	if endPort == nil {
		endPort = &port.Port
	}

	return machinebrokerv1alpha1.LoadBalancerPort{
		Protocol: *protocol,
		Port:     port.Port,
		EndPort:  *endPort,
	}
}

func ConvertNetworkingLoadBalancerPortsToLoadBalancerPorts(ports []networkingv1alpha1.LoadBalancerPort) []machinebrokerv1alpha1.LoadBalancerPort {
	res := make([]machinebrokerv1alpha1.LoadBalancerPort, len(ports))
	for i, port := range ports {
		res[i] = ConvertNetworkingLoadBalancerPortToLoadBalancerPort(port)
	}
	return res
}

func ConvertLoadBalancerPortToNetworkingLoadBalancerPort(port machinebrokerv1alpha1.LoadBalancerPort) networkingv1alpha1.LoadBalancerPort {
	protocol := port.Protocol
	endPort := port.EndPort
	return networkingv1alpha1.LoadBalancerPort{
		Protocol: &protocol,
		Port:     port.Port,
		EndPort:  &endPort,
	}
}

func ConvertLoadBalancerPortsToNetworkingLoadBalancerPorts(ports []machinebrokerv1alpha1.LoadBalancerPort) []networkingv1alpha1.LoadBalancerPort {
	res := make([]networkingv1alpha1.LoadBalancerPort, len(ports))
	for i, port := range ports {
		res[i] = ConvertLoadBalancerPortToNetworkingLoadBalancerPort(port)
	}
	return res
}
