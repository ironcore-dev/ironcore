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
	"fmt"

	"github.com/onmetal/controller-utils/conditionutils"
	"github.com/onmetal/onmetal-api/apis/ipam"
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

func clusterPrefixAllocationReadyState(clusterPrefixAllocation *ipam.ClusterPrefixAllocation) string {
	readyCond := ipam.ClusterPrefixAllocationCondition{}
	conditionutils.MustFindSlice(clusterPrefixAllocation.Status.Conditions, string(ipam.ClusterPrefixAllocationReady), &readyCond)
	return readyCond.Reason
}

func clusterPrefixAllocationRequest(clusterPrefixAllocation *ipam.ClusterPrefixAllocation) string {
	req := clusterPrefixAllocation.Spec.ClusterPrefixAllocationRequest
	switch {
	case req.Prefix.IsValid():
		return req.Prefix.String()
	case req.PrefixLength > 0:
		return fmt.Sprintf("/%d", req.PrefixLength)
	default:
		return ""
	}
}

func clusterPrefixAllocationResult(clusterPrefixAllocation *ipam.ClusterPrefixAllocation) string {
	req := clusterPrefixAllocation.Status.ClusterPrefixAllocationResult
	switch {
	case req.Prefix.IsValid():
		return req.Prefix.String()
	default:
		return ""
	}
}

func (c *convertor) ConvertToTable(ctx context.Context, obj runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	tab := &metav1.Table{
		ColumnDefinitions: headers,
	}

	if m, err := meta.ListAccessor(obj); err == nil {
		tab.ResourceVersion = m.GetResourceVersion()
		tab.SelfLink = m.GetSelfLink()
		tab.Continue = m.GetContinue()
	} else {
		if m, err := meta.CommonAccessor(obj); err == nil {
			tab.ResourceVersion = m.GetResourceVersion()
			tab.SelfLink = m.GetSelfLink()
		}
	}

	var err error
	tab.Rows, err = table.MetaToTableRow(obj, func(obj runtime.Object, m metav1.Object, name, age string) (cells []interface{}, err error) {
		clusterPrefixAllocation := obj.(*ipam.ClusterPrefixAllocation)

		cells = append(cells, name)
		if prefixRef := clusterPrefixAllocation.Spec.PrefixRef; prefixRef != nil {
			cells = append(cells, prefixRef.Name)
		} else {
			cells = append(cells, "<none>")
		}
		if request := clusterPrefixAllocationRequest(clusterPrefixAllocation); request != "" {
			cells = append(cells, request)
		} else {
			cells = append(cells, "<invalid>")
		}
		if readyState := clusterPrefixAllocationReadyState(clusterPrefixAllocation); readyState != "" {
			cells = append(cells, readyState)
		} else {
			cells = append(cells, "<unknown>")
		}
		if result := clusterPrefixAllocationResult(clusterPrefixAllocation); result != "" {
			cells = append(cells, result)
		} else {
			cells = append(cells, "<unknown>")
		}
		cells = append(cells, age)

		return cells, nil
	})
	return tab, err
}
