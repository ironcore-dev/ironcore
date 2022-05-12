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

package networking

import (
	"context"
	"fmt"

	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NetworkInterfaceVirtualIPName(nicName string, vipSource networkingv1alpha1.VirtualIPSource) string {
	switch {
	case vipSource.VirtualIPRef != nil:
		return vipSource.VirtualIPRef.Name
	case vipSource.Ephemeral != nil:
		return nicName
	default:
		return ""
	}
}

func NetworkInterfaceEphemeralIPName(nicName string, idx int) string {
	return fmt.Sprintf("%s-%d", nicName, idx)
}

const networkInterfaceVirtualIPNames = "networkinterface-virtual-ip-names"

func SetupNetworkInterfaceVirtualIPNameFieldIndexer(mgr ctrl.Manager) error {
	ctx := context.Background()
	return mgr.GetFieldIndexer().IndexField(ctx, &networkingv1alpha1.NetworkInterface{}, networkInterfaceVirtualIPNames, func(obj client.Object) []string {
		nic := obj.(*networkingv1alpha1.NetworkInterface)

		virtualIP := nic.Spec.VirtualIP
		if virtualIP == nil {
			return nil
		}

		virtualIPName := NetworkInterfaceVirtualIPName(nic.Name, *nic.Spec.VirtualIP)
		if virtualIPName == "" {
			return nil
		}

		return []string{virtualIPName}
	})
}
