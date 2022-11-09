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

import "fmt"

type machineNotFoundError struct {
	machineID string
}

func (m *machineNotFoundError) Error() string {
	return fmt.Sprintf("Machine %s not found", m.machineID)
}

func newMachineNotFoundError(machineID string) *machineNotFoundError {
	return &machineNotFoundError{
		machineID: machineID,
	}
}

type networkInterfaceNotFoundError struct {
	machineID            string
	networkInterfaceName string
}

func (e *networkInterfaceNotFoundError) Error() string {
	return fmt.Sprintf("Machine %s network interface %s not found", e.machineID, e.networkInterfaceName)
}

func newNetworkInterfaceNotFoundError(machineID, networkInterfaceName string) *networkInterfaceNotFoundError {
	return &networkInterfaceNotFoundError{
		machineID:            machineID,
		networkInterfaceName: networkInterfaceName,
	}
}
