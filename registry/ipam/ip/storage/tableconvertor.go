package storage

import (
	"context"

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
		{Name: "IP", Type: "string", Description: "The allocated ip, if any"},
		{Name: "Prefix", Type: "string", Description: "The ip to allocate the prefix from, if any"},
		{Name: "State", Type: "string", Description: "The allocation of the prefix"},
		{Name: "Age", Type: "string", Format: "date", Description: objectMetaSwaggerDoc["creationTimestamp"]},
	}
)

func newTableConvertor() *convertor {
	return &convertor{}
}

func ipReadyState(ip *ipam.IP) string {
	readyCond := ipam.IPCondition{}
	conditionutils.MustFindSlice(ip.Status.Conditions, string(ipam.IPReady), &readyCond)
	return readyCond.Reason
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
		ip := obj.(*ipam.IP)

		cells = append(cells, name)
		if ip := ip.Spec.IP; ip.IsValid() {
			cells = append(cells, ip.String())
		} else {
			cells = append(cells, "<none>")
		}
		if prefixRef := ip.Spec.PrefixRef; prefixRef != nil {
			cells = append(cells, prefixRef.String())
		} else {
			cells = append(cells, "<none>")
		}
		if readyState := ipReadyState(ip); readyState != "" {
			cells = append(cells, readyState)
		} else {
			cells = append(cells, "unknown")
		}
		cells = append(cells, age)

		return cells, nil
	})
	return tab, err
}
