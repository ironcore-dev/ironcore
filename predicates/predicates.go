package predicates

import (
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"github.com/onmetal/onmetal-api/equality"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type IPAMRangeAllocationsChangedPredicate struct {
	predicate.Funcs
}

func (IPAMRangeAllocationsChangedPredicate) Update(event event.UpdateEvent) bool {
	oldIpamRange, newIpamRange := event.ObjectOld.(*networkv1alpha1.IPAMRange), event.ObjectNew.(*networkv1alpha1.IPAMRange)

	return !equality.Semantic.DeepEqual(oldIpamRange.Status.Allocations, newIpamRange.Status.Allocations)
}
