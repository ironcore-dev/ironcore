// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"sort"
	"strings"

	"github.com/ironcore-dev/ironcore/internal/apis/storage"
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
		{Name: "VolumeClasses", Type: "string", Description: "Volume classes available on this volume pool."},
		{Name: "Age", Type: "string", Format: "date", Description: objectMetaSwaggerDoc["creationTimestamp"]},
	}
)

func newTableConvertor() *convertor {
	return &convertor{}
}

func volumePoolVolumeClassNames(volumePool *storage.VolumePool) []string {
	names := make([]string, 0, len(volumePool.Status.AvailableVolumeClasses))
	for _, volumeClassRef := range volumePool.Status.AvailableVolumeClasses {
		names = append(names, volumeClassRef.Name)
	}
	sort.Strings(names)
	return names
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
		volumePool := obj.(*storage.VolumePool)

		cells = append(cells, name)
		volumeClassNames := volumePoolVolumeClassNames(volumePool)
		if len(volumeClassNames) == 0 {
			cells = append(cells, "<none>")
		} else {
			cells = append(cells, strings.Join(volumeClassNames, ","))
		}
		cells = append(cells, age)

		return cells, nil
	})
	return tab, err
}
