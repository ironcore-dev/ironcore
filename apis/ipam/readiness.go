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

package ipam

import (
	"github.com/onmetal/controller-utils/conditionutils"
	corev1 "k8s.io/api/core/v1"
)

type Readiness string

const (
	ReadinessFailed    Readiness = "Failed"
	ReadinessUnknown   Readiness = "Unknown"
	ReadinessPending   Readiness = "Pending"
	ReadinessSucceeded Readiness = "Succeeded"
)

const (
	ReasonPending = "Pending"
	ReasonFailed  = "Failed"
)

var readinessToInt = map[Readiness]int{
	ReadinessFailed:    -1,
	ReadinessUnknown:   0,
	ReadinessPending:   1,
	ReadinessSucceeded: 2,
}

func (r1 Readiness) Compare(r2 Readiness) int {
	return readinessToInt[r1] - readinessToInt[r2]
}

// Terminal reports whether the Readiness is a terminal / final Readiness that cannot be changed from.
func (r1 Readiness) Terminal() bool {
	switch r1 {
	case ReadinessSucceeded, ReadinessFailed:
		return true
	default:
		return false
	}
}

func ReadinessFromStatusAndReason(status corev1.ConditionStatus, reason string) Readiness {
	switch {
	case status == corev1.ConditionTrue:
		return ReadinessSucceeded
	case status == corev1.ConditionFalse && reason == ReasonPending:
		return ReadinessPending
	case status == corev1.ConditionFalse && reason == ReasonFailed:
		return ReadinessFailed
	default:
		return ReadinessUnknown
	}
}

func GetPrefixAllocationConditionsReadinessAndIndex(conditions []PrefixAllocationCondition) (Readiness, int) {
	idx := conditionutils.MustFindSliceIndex(conditions, string(PrefixAllocationReady))
	if idx < 0 {
		return ReadinessUnknown, idx
	}

	cond := &conditions[idx]
	readiness := ReadinessFromStatusAndReason(cond.Status, cond.Reason)
	return readiness, idx
}

func GetPrefixAllocationReadiness(prefixAllocation *PrefixAllocation) Readiness {
	readiness, _ := GetPrefixAllocationConditionsReadinessAndIndex(prefixAllocation.Status.Conditions)
	return readiness
}

func GetPrefixConditionsReadinessAndIndex(conditions []PrefixCondition) (Readiness, int) {
	idx := conditionutils.MustFindSliceIndex(conditions, string(PrefixAllocationReady))
	if idx < 0 {
		return ReadinessUnknown, idx
	}

	cond := &conditions[idx]
	readiness := ReadinessFromStatusAndReason(cond.Status, cond.Reason)
	return readiness, idx
}

func GetPrefixReadiness(prefix *Prefix) Readiness {
	readiness, _ := GetPrefixConditionsReadinessAndIndex(prefix.Status.Conditions)
	return readiness
}
