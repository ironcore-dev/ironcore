// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"fmt"

	"github.com/ironcore-dev/ironcore/internal/apis/ipam"
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
		{Name: "Prefix", Type: "string", Description: "The targeted prefix"},
		{Name: "Request", Type: "string", Description: "Requested prefix / prefix length"},
		{Name: "State", Type: "string", Description: "The allocation of the prefix"},
		{Name: "Result", Type: "string", Description: "The resulted prefix, if any"},
		{Name: "Age", Type: "string", Format: "date", Description: objectMetaSwaggerDoc["creationTimestamp"]},
	}
)

func newTableConvertor() *convertor {
	return &convertor{}
}

func prefixAllocationRequest(prefixAllocation *ipam.PrefixAllocation) string {
	spec := prefixAllocation.Spec
	switch {
	case spec.Prefix.IsValid():
		return spec.Prefix.String()
	case spec.PrefixLength > 0:
		return fmt.Sprintf("/%d", spec.PrefixLength)
	default:
		return ""
	}
}

func prefixAllocationResult(prefixAllocation *ipam.PrefixAllocation) string {
	if prefix := prefixAllocation.Status.Prefix; prefix.IsValid() {
		return prefix.String()
	}
	return ""
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
		prefixAllocation := obj.(*ipam.PrefixAllocation)

		cells = append(cells, name)
		if prefixRef := prefixAllocation.Spec.PrefixRef; prefixRef != nil {
			cells = append(cells, prefixRef.Name)
		} else {
			cells = append(cells, "<none>")
		}
		if request := prefixAllocationRequest(prefixAllocation); request != "" {
			cells = append(cells, request)
		} else {
			cells = append(cells, "<invalid>")
		}
		if phase := prefixAllocation.Status.Phase; phase != "" {
			cells = append(cells, phase)
		} else {
			cells = append(cells, "<unknown>")
		}
		if result := prefixAllocationResult(prefixAllocation); result != "" {
			cells = append(cells, result)
		} else {
			cells = append(cells, "<unknown>")
		}
		cells = append(cells, age)

		return cells, nil
	})
	return tab, err
}
