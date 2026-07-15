// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

func GetAddressWithTimeout(timeout time.Duration, explicitAddress string, addressEnv string, wellKnownEndpoints []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return GetAddress(ctx, explicitAddress, addressEnv, wellKnownEndpoints)
}

func GetAddress(ctx context.Context, explicitAddress string, addressEnv string, wellKnownEndpoints []string) (string, error) {
	if explicitAddress != "" {
		return explicitAddress, nil
	}

	if address := os.Getenv(addressEnv); address != "" {
		return address, nil
	}

	var endpoint string
	if err := wait.PollUntilContextCancel(ctx, 1*time.Second, true, func(ctx context.Context) (done bool, err error) {
		for _, wellKnownEndpoint := range wellKnownEndpoints {
			if stat, err := os.Stat(wellKnownEndpoint); err == nil && stat.Mode().Type()&os.ModeSocket != 0 {
				endpoint = wellKnownEndpoint
				return true, nil
			}
		}
		return false, nil
	}); err != nil {
		return "", fmt.Errorf("could not determine which endpoint to use")
	}

	return fmt.Sprintf("unix://%s", endpoint), nil
}
