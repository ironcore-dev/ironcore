// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"

	"github.com/ironcore-dev/ironcore/internal/apis/compute"
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
		{Name: "CPU", Type: "string", Description: "Amount of cpu the machine class offers."},
		{Name: "Memory", Type: "string", Description: "Amount of memory the machine class offers."},
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
		machineClass := obj.(*compute.MachineClass)

		cells = append(cells, name)
		cells = append(cells, machineClass.Capabilities.CPU())
		cells = append(cells, machineClass.Capabilities.Memory())
		cells = append(cells, age)

		return cells, nil
	})
	return tab, err
}
