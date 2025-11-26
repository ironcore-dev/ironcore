// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SetTopologyLabels(log logr.Logger, om *v1.ObjectMeta, labels map[commonv1alpha1.TopologyLabel]string) {
	if len(labels) == 0 {
		return
	}

	if om.Labels == nil {
		om.Labels = make(map[string]string)
	}

	for key, val := range labels {
		log.V(1).Info("Setting topology label", "Label", key, "Value", val)
		om.Labels[string(key)] = val
	}
}
