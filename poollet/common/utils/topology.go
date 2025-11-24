package utils

import (
	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func EnforceTopolgyLabels(log logr.Logger, om *v1.ObjectMeta, labels map[commonv1alpha1.TopologyLabel]string) bool {
	changed := false

	for key, val := range labels {
		if om.Labels[string(key)] != val {
			log.V(1).Info("Restoring topology label", "Key", key, "Value", val)
			if om.Labels == nil {
				om.Labels = make(map[string]string)
			}
			om.Labels[string(key)] = val
			changed = true
		}
	}

	return changed
}

func SetTopologyLabels(log logr.Logger, om *v1.ObjectMeta, labels map[commonv1alpha1.TopologyLabel]string) {
	log.V(1).Info("Initially setting topology labels")
	for key, val := range labels {
		if om.Labels == nil {
			om.Labels = make(map[string]string)
		}
		log.V(1).Info("Setting topology label", "Label", key, "Value", val)
		om.Labels[string(key)] = val
	}
}
