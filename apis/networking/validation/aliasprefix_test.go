/*
 * Copyright (c) 2022 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package validation

import (
	"github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	"github.com/onmetal/onmetal-api/apis/networking"
	"github.com/onmetal/onmetal-api/testutils/validation"
	. "github.com/onmetal/onmetal-api/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("AliasPrefix", func() {
	DescribeTable("ValidateAliasPrefix",
		func(aliasPrefix *networking.AliasPrefix, match types.GomegaMatcher) {
			errList := ValidateAliasPrefix(aliasPrefix)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&networking.AliasPrefix{},
			ContainElement(validation.RequiredField("metadata.name")),
		),
		Entry("missing namespace",
			&networking.AliasPrefix{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			ContainElement(validation.RequiredField("metadata.namespace")),
		),
		Entry("bad name",
			&networking.AliasPrefix{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(validation.InvalidField("metadata.name")),
		),
		Entry("no network ref",
			&networking.AliasPrefix{},
			ContainElement(RequiredField("spec.networkRef")),
		),
		Entry("invalid network ref name",
			&networking.AliasPrefix{
				Spec: networking.AliasPrefixSpec{
					NetworkRef: corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(InvalidField("spec.networkRef.name")),
		),
		Entry("missing ephemeral prefix template",
			&networking.AliasPrefix{
				Spec: networking.AliasPrefixSpec{
					Prefix: networking.PrefixSource{
						Ephemeral: &networking.EphemeralPrefixSource{},
					},
				},
			},
			ContainElement(RequiredField("spec.prefix.ephemeral.prefixTemplate")),
		),
		Entry("ephemeral prefix and value are provided",
			&networking.AliasPrefix{
				Spec: networking.AliasPrefixSpec{
					Prefix: networking.PrefixSource{
						Ephemeral: &networking.EphemeralPrefixSource{},
						Value:     v1alpha1.PtrToIPPrefix(v1alpha1.MustParseIPPrefix("10.0.0.0/24")),
					},
				},
			},
			ContainElement(ForbiddenField("spec.prefix")),
		),
	)

	DescribeTable("ValidateAliasPrefixUpdate",
		func(newAliasPrefix, oldAliasPrefix *networking.AliasPrefix, match types.GomegaMatcher) {
			errList := ValidateAliasPrefixUpdate(newAliasPrefix, oldAliasPrefix)
			Expect(errList).To(match)
		},
		Entry("immutable networkRef",
			&networking.AliasPrefix{
				Spec: networking.AliasPrefixSpec{
					NetworkRef: corev1.LocalObjectReference{Name: "foo"},
				},
			},
			&networking.AliasPrefix{
				Spec: networking.AliasPrefixSpec{
					NetworkRef: corev1.LocalObjectReference{Name: "foo*"},
				},
			},
			ContainElement(ForbiddenField("spec.networkRef")),
		),
		Entry("mutable network interface selector",
			&networking.AliasPrefix{
				Spec: networking.AliasPrefixSpec{
					NetworkInterfaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"foo": "bar"},
					},
				},
			},
			&networking.AliasPrefix{
				Spec: networking.AliasPrefixSpec{
					NetworkInterfaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"bar": "baz"},
					},
				},
			},
			Not(ContainElement(ForbiddenField("spec"))),
		),
	)
})
