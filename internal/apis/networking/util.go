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

import "fmt"

// NetworkInterfaceVirtualIPName returns the name of a VirtualIP for a NetworkInterface VirtualIPSource.
func NetworkInterfaceVirtualIPName(nicName string, vipSource VirtualIPSource) string {
	switch {
	case vipSource.VirtualIPRef != nil:
		return vipSource.VirtualIPRef.Name
	case vipSource.Ephemeral != nil:
		return nicName
	default:
		return ""
	}
}

// NetworkInterfaceIPSourceEphemeralPrefixName returns the name of a Prefix for a network interface ephemeral prefix.
func NetworkInterfaceIPSourceEphemeralPrefixName(nicName string, idx int) string {
	return fmt.Sprintf("%s-%d", nicName, idx)
}
