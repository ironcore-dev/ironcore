// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"strings"

	"github.com/ironcore-dev/ironcore/internal/apis/networking"
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
