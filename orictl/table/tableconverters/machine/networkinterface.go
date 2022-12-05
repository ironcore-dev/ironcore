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
	"strings"
	"time"

	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"github.com/onmetal/onmetal-api/orictl/table"
	"github.com/onmetal/onmetal-api/orictl/table/tableconverter"
	"k8s.io/apimachinery/pkg/util/duration"
)

var (
	networkInterfaceHeaders = []table.Header{
		{Name: "ID"},
		{Name: "Network Handle"},
		{Name: "IPs"},
		{Name: "Virtual IP"},
		{Name: "Age"},
	}
)

var NetworkInterface, NetworkInterfaceSlice = tableconverter.ForType[*ori.NetworkInterface]( //nolint:revive
	func() ([]table.Header, error) {
		return networkInterfaceHeaders, nil
	},
	func(networkInterface *ori.NetworkInterface) ([]table.Row, error) {
		return []table.Row{
			{
				networkInterface.Metadata.Id,
				networkInterface.Spec.Network.Handle,
				strings.Join(networkInterface.Spec.Ips, ","),
				networkInterface.Spec.GetVirtualIp().GetIp(),
				duration.HumanDuration(time.Since(time.Unix(0, networkInterface.Metadata.CreatedAt))),
			},
		}, nil
	},
)

func init() {
	RegistryBuilder.Register(
		tableconverter.ToTaggedAny(NetworkInterface),
		tableconverter.ToTaggedAny(NetworkInterfaceSlice),
	)
}
