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

func prefixAllocationReadyState(prefixAllocation *ipam.PrefixAllocation) string {
	readyCond := ipam.PrefixAllocationCondition{}
	conditionutils.MustFindSlice(prefixAllocation.Status.Conditions, string(ipam.PrefixAllocationReady), &readyCond)
	return readyCond.Reason
}

func prefixAllocationRequest(prefixAllocation *ipam.PrefixAllocation) string {
	req := prefixAllocation.Spec.PrefixAllocationRequest
	switch {
	case req.Prefix.IsValid():
		return req.Prefix.String()
	case req.PrefixLength > 0:
		return fmt.Sprintf("/%d", req.PrefixLength)
	case req.Range.IsValid():
		return req.Range.String()
	case req.RangeLength > 0:
		return fmt.Sprintf("rlen %d", req.RangeLength)
	default:
		return ""
	}
}

func prefixAllocationResult(prefixAllocation *ipam.PrefixAllocation) string {
	res := prefixAllocation.Status.PrefixAllocationResult
	switch {
	case res.Prefix.IsValid():
		return res.Prefix.String()
	case res.Range.IsValid():
		return res.Range.String()
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
		prefixAllocation := obj.(*ipam.PrefixAllocation)

		cells = append(cells, name)
		if prefixRef := prefixAllocation.Spec.PrefixRef; prefixRef != nil {
			cells = append(cells, prefixRef.String())
		} else {
			cells = append(cells, "<none>")
		}
		if request := prefixAllocationRequest(prefixAllocation); request != "" {
			cells = append(cells, request)
		} else {
			cells = append(cells, "<invalid>")
		}
		if readyState := prefixAllocationReadyState(prefixAllocation); readyState != "" {
			cells = append(cells, readyState)
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
