// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/apis/ipam"
	. "github.com/ironcore-dev/ironcore/internal/apis/ipam/validation"
	. "github.com/ironcore-dev/ironcore/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("Prefix", func() {
	DescribeTable("ValidatePrefix",
		func(prefix *ipam.Prefix, match types.GomegaMatcher) {
			errList := ValidatePrefix(prefix)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&ipam.Prefix{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&ipam.Prefix{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&ipam.Prefix{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
		Entry("missing ip family",
			&ipam.Prefix{},
			ContainElement(RequiredField("spec.ipFamily")),
		),
		Entry("invalid ip family",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					IPFamily: "invalid",
				},
			},
			ContainElement(NotSupportedField("spec.ipFamily")),
		),
		Entry("invalid prefix",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					Prefix: &commonv1alpha1.IPPrefix{},
				},
			},
			ContainElement(InvalidField("spec.prefix")),
		),
		Entry("IPv4 ipFamily prefix mismatch",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					IPFamily: corev1.IPv4Protocol,
					Prefix:   commonv1alpha1.MustParseNewIPPrefix("beef::/64"),
				},
			},
			ContainElement(InvalidField("spec.prefix")),
		),
		Entry("IPv6 ipFamily prefix mismatch",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					IPFamily: corev1.IPv6Protocol,
					Prefix:   commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/8"),
				},
			},
			ContainElement(InvalidField("spec.prefix")),
		),
		Entry("invalid prefixLength",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					PrefixLength: -10,
				},
			},
			ContainElement(InvalidField("spec.prefixLength")),
		),
		Entry("IPv4 ipFamily prefixLength mismatch",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					IPFamily:     corev1.IPv4Protocol,
					PrefixLength: 40,
				},
			},
			ContainElement(InvalidField("spec.prefixLength")),
		),
		Entry("IPv6 ipFamily prefixLength mismatch",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					IPFamily:     corev1.IPv6Protocol,
					PrefixLength: 140,
				},
			},
			ContainElement(InvalidField("spec.prefixLength")),
		),
		Entry("root with prefixLength",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					PrefixLength: 8,
				},
			},
			ContainElement(InvalidField("spec.prefixLength")),
		),
		Entry("root with nil prefix",
			&ipam.Prefix{},
			ContainElement(RequiredField("spec.prefix")),
		),
		Entry("empty parent reference name",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					ParentRef: &corev1.LocalObjectReference{},
				},
			},
			ContainElement(RequiredField("spec.parentRef.name")),
		),
		Entry("invalid parent reference name",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					ParentRef: &corev1.LocalObjectReference{
						Name: "foo*",
					},
				},
			},
			ContainElement(InvalidField("spec.parentRef.name")),
		),
		Entry("prefix to prefixLength mismatch",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					Prefix:       commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/8"),
					PrefixLength: 16,
				},
			},
			ContainElement(InvalidField("spec.prefixLength")),
		),
		Entry("child prefix with no request",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					ParentRef: &corev1.LocalObjectReference{Name: "parent"},
				},
			},
			ContainElement(InvalidField("spec")),
		),
		Entry("valid root prefix",
			&ipam.Prefix{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "foo"},
				Spec: ipam.PrefixSpec{
					IPFamily: corev1.IPv4Protocol,
					Prefix:   commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/8"),
				},
			},
			BeEmpty(),
		),
		Entry("valid child prefix",
			&ipam.Prefix{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "foo"},
				Spec: ipam.PrefixSpec{
					ParentRef: &corev1.LocalObjectReference{
						Name: "parent",
					},
					IPFamily:     corev1.IPv4Protocol,
					PrefixLength: 24,
				},
			},
			BeEmpty(),
		),
	)

	DescribeTable("ValidatePrefixUpdate",
		func(newPrefix, oldPrefix *ipam.Prefix, match types.GomegaMatcher) {
			errList := ValidatePrefixUpdate(newPrefix, oldPrefix)
			Expect(errList).To(match)
		},
		Entry("immutable prefix if set",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					Prefix: commonv1alpha1.MustParseNewIPPrefix("11.0.0.0/8"),
				},
			},
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					Prefix: commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/8"),
				},
			},
			ContainElement(ImmutableField("spec")),
		),
		Entry("immutable prefixLength if set",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					PrefixLength: 8,
				},
			},
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					PrefixLength: 9,
				},
			},
			ContainElement(ImmutableField("spec")),
		),
		Entry("immutable parentRef if root",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					ParentRef: &corev1.LocalObjectReference{Name: "parent"},
				},
			},
			&ipam.Prefix{},
			ContainElement(ImmutableField("spec")),
		),
		Entry("immutable parentRef if set",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					ParentRef: &corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					ParentRef: &corev1.LocalObjectReference{Name: "bar"},
				},
			},
			ContainElement(ImmutableField("spec")),
		),
		Entry("mutable parentRef if child and unset",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					ParentRef: &corev1.LocalObjectReference{Name: "parent"},
					ParentSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"foo": "bar"},
					},
				},
			},
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					ParentSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"foo": "bar"},
					},
				},
			},
			Not(ContainElement(ImmutableField("spec"))),
		),
		Entry("mutable prefix if unset",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					Prefix: commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/8"),
				},
			},
			&ipam.Prefix{},
			Not(ContainElement(ImmutableField("spec"))),
		),
	)

	DescribeTable("ValidatePrefixStatus",
		func(status *ipam.PrefixStatus, match types.GomegaMatcher) {
			errList := ValidatePrefixStatus(status, field.NewPath("status"))
			Expect(errList).To(match)
		},
		Entry("allocated invalid used",
			&ipam.PrefixStatus{
				Phase: ipam.PrefixPhaseAllocated,
				Used:  []commonv1alpha1.IPPrefix{{}},
			},
			ContainElement(InvalidField("status.used[0]")),
		),
		Entry("allocated overlapping used",
			&ipam.PrefixStatus{
				Phase: ipam.PrefixPhaseAllocated,
				Used: []commonv1alpha1.IPPrefix{
					commonv1alpha1.MustParseIPPrefix("10.0.0.0/8"),
					commonv1alpha1.MustParseIPPrefix("10.0.0.0/16"),
				},
			},
			ContainElement(InvalidField("status.used")),
		),
		Entry("allocated different ip families used",
			&ipam.PrefixStatus{
				Phase: ipam.PrefixPhaseAllocated,
				Used: []commonv1alpha1.IPPrefix{
					commonv1alpha1.MustParseIPPrefix("10.0.0.0/8"),
					commonv1alpha1.MustParseIPPrefix("beef::/64"),
				},
			},
			ContainElement(InvalidField("status.used")),
		),
		Entry("not allocated but used",
			&ipam.PrefixStatus{
				Used: []commonv1alpha1.IPPrefix{{}},
			},
			ContainElement(InvalidField("status.used")),
		),
		Entry("allocated valid",
			&ipam.PrefixStatus{
				Phase: ipam.PrefixPhaseAllocated,
				Used: []commonv1alpha1.IPPrefix{
					commonv1alpha1.MustParseIPPrefix("10.0.0.0/8"),
					commonv1alpha1.MustParseIPPrefix("11.0.0.0/8"),
				},
			},
			BeEmpty(),
		),
		Entry("not allocated valid",
			&ipam.PrefixStatus{},
			BeEmpty(),
		),
	)

	DescribeTable("ValidatePrefixStatusUpdate",
		func(newPrefix, oldPrefix *ipam.Prefix, match types.GomegaMatcher) {
			errList := ValidatePrefixStatusUpdate(newPrefix, oldPrefix)
			Expect(errList).To(match)
		},
		Entry("terminal to missing ready condition",
			&ipam.Prefix{},
			&ipam.Prefix{
				Status: ipam.PrefixStatus{
					Phase: ipam.PrefixPhaseAllocated,
				},
			},
			ContainElement(ForbiddenField("status.phase")),
		),
		Entry("terminal to non-terminal ready condition",
			&ipam.Prefix{
				Status: ipam.PrefixStatus{
					Phase: ipam.PrefixPhasePending,
				},
			},
			&ipam.Prefix{
				Status: ipam.PrefixStatus{
					Phase: ipam.PrefixPhaseAllocated,
				},
			},
			ContainElement(ForbiddenField("status.phase")),
		),
		Entry("non-terminal to terminal ready condition",
			&ipam.Prefix{
				Status: ipam.PrefixStatus{
					Phase: ipam.PrefixPhaseAllocated,
				},
			},
			&ipam.Prefix{},
			Not(ContainElement(InvalidField("status.phase"))),
		),
		Entry("empty to allocated but prefix not valid",
			&ipam.Prefix{
				Status: ipam.PrefixStatus{
					Phase: ipam.PrefixPhaseAllocated,
				},
			},
			&ipam.Prefix{},
			ContainElement(RequiredField("spec.prefix")),
		),
		Entry("empty to allocated and prefix valid but child and no parent",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					Prefix: commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/8"),
					ParentSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"foo": "bar"},
					},
				},
				Status: ipam.PrefixStatus{
					Phase: ipam.PrefixPhaseAllocated,
				},
			},
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					Prefix: commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/8"),
					ParentSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"foo": "bar"},
					},
				},
			},
			ContainElement(RequiredField("spec.parentRef")),
		),
		Entry("empty to allocated and prefix valid",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					Prefix: commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/8"),
				},
				Status: ipam.PrefixStatus{
					Phase: ipam.PrefixPhaseAllocated,
				},
			},
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					Prefix: commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/8"),
				},
			},
			Not(ContainElement(ForbiddenField("status.phase"))),
		),
		Entry("used not covered by spec",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					Prefix: commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/8"),
				},
				Status: ipam.PrefixStatus{
					Used: []commonv1alpha1.IPPrefix{
						commonv1alpha1.MustParseIPPrefix("11.0.0.0/8"),
					},
					Phase: ipam.PrefixPhaseAllocated,
				},
			},
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					Prefix: commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/8"),
				},
				Status: ipam.PrefixStatus{
					Phase: ipam.PrefixPhaseAllocated,
				},
			},
			ContainElement(ForbiddenField("status.used[0]")),
		),
		Entry("used covered by spec",
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					Prefix: commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/8"),
				},
				Status: ipam.PrefixStatus{
					Phase: ipam.PrefixPhaseAllocated,
					Used: []commonv1alpha1.IPPrefix{
						commonv1alpha1.MustParseIPPrefix("10.0.0.0/8"),
					},
				},
			},
			&ipam.Prefix{
				Spec: ipam.PrefixSpec{
					Prefix: commonv1alpha1.MustParseNewIPPrefix("10.0.0.0/8"),
				},
				Status: ipam.PrefixStatus{
					Phase: ipam.PrefixPhaseAllocated,
				},
			},
			Not(ContainElement(ForbiddenField("status.used[0]"))),
		),
	)
})
