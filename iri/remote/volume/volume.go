// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package volume

import (
	"context"
	"time"

	"github.com/ironcore-dev/ironcore/iri/remote/common"
)

const (
	AddressEnv = "IRI_VOLUME_RUNTIME_ENDPOINT"
)

var WellKnownEndpoints = []string{
	"/var/run/iri-volumebroker.sock",
	"/var/run/iri-volumeprovider.sock",
}

func GetAddressWithTimeout(timeout time.Duration, explicitAddress string) (string, error) {
	return common.GetAddressWithTimeout(timeout, explicitAddress, AddressEnv, WellKnownEndpoints)
}

func GetAddress(ctx context.Context, explicitAddress string) (string, error) {
	return common.GetAddress(ctx, explicitAddress, AddressEnv, WellKnownEndpoints)
}
