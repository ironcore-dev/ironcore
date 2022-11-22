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

package machine

import (
	"fmt"
	"os"
)

const (
	AddressEnv = "ORI_MACHINE_RUNTIME_ENDPOINT"
)

var WellKnownEndpoints = []string{
	"/var/run/ori-machinebroker.sock",
	"/var/run/ori-virtd.sock",
}

func GetAddress(explicitAddress string) (string, error) {
	if explicitAddress != "" {
		return explicitAddress, nil
	}

	if address := os.Getenv(AddressEnv); address != "" {
		return address, nil
	}

	for _, wellKnownEndpoint := range WellKnownEndpoints {
		if stat, err := os.Stat(wellKnownEndpoint); err == nil && stat.Mode().Type()&os.ModeSocket != 0 {
			return wellKnownEndpoint, nil
		}
	}
	return "", fmt.Errorf("could not determine address to use")
}
