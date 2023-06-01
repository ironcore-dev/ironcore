// Copyright 2023 OnMetal authors
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

package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	v1alpha11 "github.com/onmetal/onmetal-api/ori/apis/metrics/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8smetrics "k8s.io/component-base/metrics"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

type oriMetricsCollector struct {
	k8smetrics.BaseStableCollector
	log           logr.Logger
	descriptors   map[string]*k8smetrics.Desc
	listMetricsFn func(context.Context) ([]*v1alpha11.Metric, error)
}

func NewORIMetricsCollector(ctx context.Context, log logr.Logger, listMetricsFn func(context.Context) ([]*v1alpha11.Metric, error), listMetricDescriptorsFn func(context.Context) ([]*v1alpha11.MetricDescriptor, error)) k8smetrics.StableCollector {
	descs, err := listMetricDescriptorsFn(ctx)
	if err != nil {
		if status.Code(err) != codes.Unimplemented {
			log.Error(err, "failed to list metric descriptors")
		}

		return &oriMetricsCollector{
			listMetricsFn: listMetricsFn,
		}
	}
	c := &oriMetricsCollector{
		log:           log,
		listMetricsFn: listMetricsFn,
		descriptors:   make(map[string]*k8smetrics.Desc, len(descs)),
	}

	for _, desc := range descs {
		c.descriptors[desc.Name] = oroDescToProm(desc)
	}

	return c
}

func (c *oriMetricsCollector) DescribeWithStability(ch chan<- *k8smetrics.Desc) {
	for _, desc := range c.descriptors {
		ch <- desc
	}
}

func (c *oriMetricsCollector) CollectWithStability(ch chan<- k8smetrics.Metric) {
	oriMetrics, err := c.listMetricsFn(context.Background())
	if err != nil {
		if status.Code(err) != codes.Unimplemented {
			c.log.Error(err, "failed to get pod metrics")
		}
		return
	}

	for _, metric := range oriMetrics {
		promMetric, err := c.oriMetricToProm(metric)
		if err == nil {
			ch <- promMetric
		}
	}
}

func oroDescToProm(m *v1alpha11.MetricDescriptor) *k8smetrics.Desc {
	return k8smetrics.NewDesc(m.Name, m.Help, m.LabelKeys, nil, k8smetrics.INTERNAL, "")
}

func (c *oriMetricsCollector) oriMetricToProm(m *v1alpha11.Metric) (k8smetrics.Metric, error) {
	desc, ok := c.descriptors[m.Name]
	if !ok {
		err := fmt.Errorf("failed to convert ori metric to prometheus format")
		c.log.Error(err, "descriptor not present in pre-populated list of descriptors", "name", m.Name)
		return nil, err
	}

	typ := oriTypeToProm[m.MetricType]

	pm, err := k8smetrics.NewConstMetric(desc, typ, float64(m.GetValue()), m.LabelValues...)
	if err != nil {
		c.log.Error(err, "failed to get ori prometheus metric", "descriptor", desc.String())
		return nil, err
	}

	if m.Timestamp == 0 {
		return pm, nil
	}
	return k8smetrics.NewLazyMetricWithTimestamp(time.Unix(0, m.Timestamp), pm), nil
}

var oriTypeToProm = map[v1alpha11.MetricType]k8smetrics.ValueType{
	v1alpha11.MetricType_COUNTER: k8smetrics.CounterValue,
	v1alpha11.MetricType_GAUGE:   k8smetrics.GaugeValue,
}

func Register(c k8smetrics.StableCollector) error {
	if !c.Create(nil, c) {
		return nil
	}

	if err := metrics.Registry.Register(c); err != nil {
		return fmt.Errorf("failed to register metrics: %w", err)
	}

	return nil
}
