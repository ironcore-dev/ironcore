// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

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
