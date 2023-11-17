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

package storage

import (
	"context"
	"strings"

	"github.com/ironcore-dev/ironcore/internal/apis/core"
	"golang.org/x/exp/slices"
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
		{Name: "Request", Type: "string", Description: "Request represents the usage / request of a resource"},
		{Name: "Age", Type: "string", Format: "date", Description: objectMetaSwaggerDoc["creationTimestamp"]},
	}
)

func newTableConvertor() *convertor {
	return &convertor{}
}

func formatResources(used, hard core.ResourceList) string {
	names := make([]core.ResourceName, 0, len(hard))
	for name := range hard {
		names = append(names, name)
	}
	slices.Sort(names)

	var bldr strings.Builder
	for i := range names {
		name := names[i]
		usedQuantity := used[name]
		hardQuantity := hard[name]

		if i > 0 {
			bldr.WriteString(", ")
		}
		bldr.WriteString(string(name))
		bldr.WriteString(": ")
		bldr.WriteString(usedQuantity.String())
		bldr.WriteString("/")
		bldr.WriteString(hardQuantity.String())
	}
	return bldr.String()
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
		resourceQuota := obj.(*core.ResourceQuota)

		cells = append(cells, name)
		cells = append(cells, formatResources(resourceQuota.Status.Used, resourceQuota.Status.Hard))
		cells = append(cells, age)

		return cells, nil
	})
	return tab, err
}
