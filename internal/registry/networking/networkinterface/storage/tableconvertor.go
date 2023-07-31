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

package storage

import (
	"context"
	"strings"

	"github.com/onmetal/onmetal-api/internal/apis/networking"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/meta/table"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type convertor struct{}

var (
	objectMetaSwaggerDoc = metav1.ObjectMeta{}.SwaggerDoc()

	headers = []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: objectMetaSwaggerDoc["name"]},
		{Name: "Network", Type: "string", Description: "The network this network interface is connected to"},
		{Name: "Machine", Type: "string", Description: "The machine this network interface is used by"},
		{Name: "IPs", Type: "string", Description: "List of effective IPs of the network interface"},
		{Name: "VirtualIP", Type: "string", Description: "The virtual IP assigned to this interface, if any"},
		{Name: "State", Type: "string", Description: "The state of the network interface"},
		{Name: "Age", Type: "string", Format: "date", Description: objectMetaSwaggerDoc["creationTimestamp"]},
	}
)

func newTableConvertor() *convertor {
	return &convertor{}
}

func (c *convertor) ConvertToTable(ctx context.Context, obj runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	tab := &metav1.Table{
		ColumnDefinitions: headers,
	}

	if m, err := meta.ListAccessor(obj); err == nil {
		tab.ResourceVersion = m.GetResourceVersion()
		tab.Continue = m.GetContinue()
	} else {
		if m, err := meta.CommonAccessor(obj); err == nil {
			tab.ResourceVersion = m.GetResourceVersion()
		}
	}

	var err error
	tab.Rows, err = table.MetaToTableRow(obj, func(obj runtime.Object, m metav1.Object, name, age string) (cells []interface{}, err error) {
		networkInterface := obj.(*networking.NetworkInterface)

		cells = append(cells, name)
		cells = append(cells, networkInterface.Spec.NetworkRef.Name)
		if machineRef := networkInterface.Spec.MachineRef; machineRef != nil {
			cells = append(cells, machineRef.Name)
		} else {
			cells = append(cells, "<none>")
		}
		if effectiveIPs := networkInterface.Status.IPs; len(effectiveIPs) > 0 {
			var eIPs []string
			for _, ip := range effectiveIPs {
				eIPs = append(eIPs, ip.String())
			}
			cells = append(cells, strings.Join(eIPs, ","))
		} else {
			cells = append(cells, "<none>")
		}
		if vip := networkInterface.Status.VirtualIP; vip != nil {
			cells = append(cells, vip.String())
		} else {
			cells = append(cells, "<none>")
		}
		cells = append(cells, networkInterface.Status.State)
		cells = append(cells, age)

		return cells, nil
	})
	return tab, err
}
