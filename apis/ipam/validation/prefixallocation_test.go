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

package validation_test

import (
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	"github.com/onmetal/onmetal-api/apis/ipam"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"

	. "github.com/onmetal/onmetal-api/apis/ipam/validation"
)

var _ = Describe("PrefixAllocation", func() {
	DescribeTable("ValidatePrefixAllocation",
		func(prefixAllocation *ipam.PrefixAllocation, match types.GomegaMatcher) {
			errList := ValidatePrefixAllocation(prefixAllocation)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&ipam.PrefixAllocation{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&ipam.PrefixAllocation{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&ipam.PrefixAllocation{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("missing ip family",
			&ipam.PrefixAllocation{},
			ContainElement(RequiredField("spec.ipFamily")),
		),
		Entry("invalid ip family",
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					IPFamily: "invalid",
				},
			},
			ContainElement(UnsupportedField("spec.ipFamily")),
		),
		Entry("negative prefixLength",
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					PrefixLength: -8,
				},
			},
			ContainElement(InvalidField("spec.prefixLength")),
		),
		Entry("empty prefix",
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					Prefix: &commonv1alpha1.IPPrefix{},
				},
			},
			ContainElement(InvalidField("spec.prefix")),
		),
		Entry("IPv4 ipFamily prefixLength mismatch",
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					IPFamily:     corev1.IPv4Protocol,
					PrefixLength: 40,
				},
			},
			ContainElement(InvalidField("spec.prefixLength")),
		),
		Entry("IPv6 ipFamily prefixLength mismatch",
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					IPFamily:     corev1.IPv6Protocol,
					PrefixLength: 140,
				},
			},
			ContainElement(InvalidField("spec.prefixLength")),
		),
		Entry("no request",
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					PrefixRef: &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			ContainElement(InvalidField("spec")),
		),
		Entry("multiple requests",
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					Prefix:       commonv1alpha1.PtrToIPPrefix(commonv1alpha1.MustParseIPPrefix("10.0.0.0/8")),
					PrefixLength: 8,
				},
			},
			ContainElement(ForbiddenField("spec.prefixLength")),
		),
		Entry("no source",
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					Prefix:       commonv1alpha1.PtrToIPPrefix(commonv1alpha1.MustParseIPPrefix("10.0.0.0/8")),
					PrefixLength: 8,
				},
			},
			ContainElement(InvalidField("spec")),
		),
		Entry("empty prefix ref",
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					PrefixRef: &corev1.LocalObjectReference{},
				},
			},
			ContainElement(RequiredField("spec.prefixRef.name")),
		),
		Entry("invalid prefix ref name",
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					PrefixRef: &corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.prefixRef.name")),
		),
		Entry("valid prefix allocation",
			&ipam.PrefixAllocation{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "foo"},
				Spec: ipam.PrefixAllocationSpec{
					IPFamily:     corev1.IPv4Protocol,
					PrefixLength: 24,
					PrefixRef: &corev1.LocalObjectReference{
						Name: "foo",
					},
				},
			},
			BeEmpty(),
		),
	)

	DescribeTable("ValidatePrefixAllocationUpdate",
		func(newPrefixAllocation, oldPrefixAllocation *ipam.PrefixAllocation, match types.GomegaMatcher) {
			errList := ValidatePrefixAllocationUpdate(newPrefixAllocation, oldPrefixAllocation)
			Expect(errList).To(match)
		},
		Entry("immutable prefix ref if set",
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					PrefixRef: &corev1.LocalObjectReference{
						Name: "foo",
					},
				},
			},
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					PrefixRef: &corev1.LocalObjectReference{
						Name: "bar",
					},
				},
			},
			ContainElement(ImmutableField("spec")),
		),
		Entry("mutable prefix ref if unset",
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					PrefixRef: &corev1.LocalObjectReference{
						Name: "foo",
					},
				},
			},
			&ipam.PrefixAllocation{},
			Not(ContainElement(ImmutableField("spec"))),
		),
	)

	DescribeTable("ValidatePrefixAllocationStatus",
		func(status *ipam.PrefixAllocationStatus, match types.GomegaMatcher) {
			errList := ValidatePrefixAllocationStatus(status, field.NewPath("status"))
			Expect(errList).To(match)
		},
		Entry("succeeded but no result",
			&ipam.PrefixAllocationStatus{
				Conditions: []ipam.PrefixAllocationCondition{
					{
						Type:   ipam.PrefixAllocationReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
			ContainElement(RequiredField("status.prefix")),
		),
		Entry("not succeeded but result",
			&ipam.PrefixAllocationStatus{
				Prefix: commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/8"),
			},
			ContainElement(ForbiddenField("status.prefix")),
		),
		Entry("succeeded with result",
			&ipam.PrefixAllocationStatus{
				Prefix: commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/8"),
				Conditions: []ipam.PrefixAllocationCondition{
					{
						Type:   ipam.PrefixAllocationReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
			Not(ContainElement(ForbiddenField("status.prefix"))),
		),
	)

	DescribeTable("ValidatePrefixAllocationStatusUpdate",
		func(newPrefixAllocation, oldPrefixAllocation *ipam.PrefixAllocation, match types.GomegaMatcher) {
			errList := ValidatePrefixAllocationStatusUpdate(newPrefixAllocation, oldPrefixAllocation)
			Expect(errList).To(match)
		},
		Entry("update from terminal readiness to missing readiness",
			&ipam.PrefixAllocation{},
			&ipam.PrefixAllocation{
				Status: ipam.PrefixAllocationStatus{
					Conditions: []ipam.PrefixAllocationCondition{
						{
							Type:   ipam.PrefixAllocationReady,
							Status: corev1.ConditionFalse,
							Reason: ipam.ReasonFailed,
						},
					},
				},
			},
			ContainElement(RequiredField("status.conditions[0]")),
		),
		Entry("update from terminal readiness to other readiness",
			&ipam.PrefixAllocation{
				Status: ipam.PrefixAllocationStatus{
					Conditions: []ipam.PrefixAllocationCondition{
						{
							Type:   ipam.PrefixAllocationReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			&ipam.PrefixAllocation{
				Status: ipam.PrefixAllocationStatus{
					Conditions: []ipam.PrefixAllocationCondition{
						{
							Type:   ipam.PrefixAllocationReady,
							Status: corev1.ConditionFalse,
							Reason: ipam.ReasonFailed,
						},
					},
				},
			},
			ContainElement(ForbiddenField("status.conditions[0]")),
		),
		Entry("prefix mismatch with spec.prefix",
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					Prefix:    commonv1alpha1.MustParseNewIPPrefix("10.0.0.8/32"),
					PrefixRef: &corev1.LocalObjectReference{Name: "foo"},
				},
				Status: ipam.PrefixAllocationStatus{
					Prefix: commonv1alpha1.MustParseNewIPPrefix("10.0.0.9/32"),
					Conditions: []ipam.PrefixAllocationCondition{
						{
							Type:   ipam.PrefixAllocationReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					Prefix:    commonv1alpha1.MustParseNewIPPrefix("10.0.0.8/32"),
					PrefixRef: &corev1.LocalObjectReference{Name: "foo"},
				},
				Status: ipam.PrefixAllocationStatus{},
			},
			ContainElement(ForbiddenField("status.prefix")),
		),
		Entry("prefix mismatch with spec.prefixLength",
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					PrefixLength: 8,
					PrefixRef:    &corev1.LocalObjectReference{Name: "foo"},
				},
				Status: ipam.PrefixAllocationStatus{
					Prefix: commonv1alpha1.MustParseNewIPPrefix("10.0.0.9/32"),
					Conditions: []ipam.PrefixAllocationCondition{
						{
							Type:   ipam.PrefixAllocationReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					PrefixLength: 8,
					PrefixRef:    &corev1.LocalObjectReference{Name: "foo"},
				},
				Status: ipam.PrefixAllocationStatus{},
			},
			ContainElement(ForbiddenField("status.prefix")),
		),
		Entry("prefix match with spec.prefixLength but not spec.ipFamily",
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					IPFamily:     corev1.IPv4Protocol,
					PrefixLength: 8,
					PrefixRef:    &corev1.LocalObjectReference{Name: "foo"},
				},
				Status: ipam.PrefixAllocationStatus{
					Prefix: commonv1alpha1.MustParseNewIPPrefix("beef::/8"),
					Conditions: []ipam.PrefixAllocationCondition{
						{
							Type:   ipam.PrefixAllocationReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					PrefixLength: 8,
					PrefixRef:    &corev1.LocalObjectReference{Name: "foo"},
				},
				Status: ipam.PrefixAllocationStatus{},
			},
			ContainElement(InvalidField("status.prefix")),
		),
		Entry("prefix but no spec.prefixRef",
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					PrefixLength: 32,
				},
				Status: ipam.PrefixAllocationStatus{
					Prefix: commonv1alpha1.MustParseNewIPPrefix("10.0.0.9/32"),
					Conditions: []ipam.PrefixAllocationCondition{
						{
							Type:   ipam.PrefixAllocationReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			&ipam.PrefixAllocation{
				Spec: ipam.PrefixAllocationSpec{
					PrefixLength: 32,
					PrefixRef:    &corev1.LocalObjectReference{Name: "foo"},
				},
				Status: ipam.PrefixAllocationStatus{},
			},
			ContainElement(ForbiddenField("status.prefix")),
		),
	)
})
