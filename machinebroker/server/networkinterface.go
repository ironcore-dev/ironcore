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

package server

import (
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	ori "github.com/onmetal/onmetal-api/ori/apis/compute/v1alpha1"
)

func (s *Server) convertOnmetalNetworkInterface(
	machineID string,
	machineMetadata *ori.MachineMetadata,
	onmetalNetworkInterface *computev1alpha1.NetworkInterface,
	onmetalNetworkingNetworkInterface *networkingv1alpha1.NetworkInterface,
) (*ori.NetworkInterface, error) {
	ips := make([]string, len(onmetalNetworkingNetworkInterface.Status.IPs))
	for i, ip := range onmetalNetworkingNetworkInterface.Status.IPs {
		ips[i] = ip.String()
	}

	var virtualIPConfig *ori.VirtualIPConfig
	if onmetalVirtualIP := onmetalNetworkingNetworkInterface.Status.VirtualIP; onmetalVirtualIP != nil {
		virtualIPConfig = &ori.VirtualIPConfig{
			Ip: onmetalVirtualIP.String(),
		}
	}

	return &ori.NetworkInterface{
		MachineId:       machineID,
		MachineMetadata: machineMetadata,
		Name:            onmetalNetworkInterface.Name,
		Network:         &ori.NetworkConfig{Handle: onmetalNetworkingNetworkInterface.Status.NetworkHandle},
		Ips:             ips,
		VirtualIp:       virtualIPConfig,
	}, nil
}
