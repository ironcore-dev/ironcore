// Copyright 2022 IronCore authors
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

package volume

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	AddressEnv = "IRI_VOLUME_RUNTIME_ENDPOINT"
)

var WellKnownEndpoints = []string{
	"/var/run/iri-volumebroker.sock",
	"/var/run/iri-cephd.sock",
}

func GetAddressWithTimeout(timeout time.Duration, explicitAddress string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return GetAddress(ctx, explicitAddress)
}

func GetAddress(ctx context.Context, explicitAddress string) (string, error) {
	if explicitAddress != "" {
		return explicitAddress, nil
	}

	if address := os.Getenv(AddressEnv); address != "" {
		return address, nil
	}

	var endpoint string
	if err := wait.PollUntilContextCancel(ctx, 1*time.Second, true, func(ctx context.Context) (done bool, err error) {
		for _, wellKnownEndpoint := range WellKnownEndpoints {
			if stat, err := os.Stat(wellKnownEndpoint); err == nil && stat.Mode().Type()&os.ModeSocket != 0 {
				endpoint = wellKnownEndpoint
				return true, nil
			}
		}
		return false, nil
	}); err != nil {
		return "", fmt.Errorf("could not determine which enpdoint to use")
	}

	return fmt.Sprintf("unix://%s", endpoint), nil
}
