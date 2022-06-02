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
	"github.com/onmetal/onmetal-api/apis/storage"
	. "github.com/onmetal/onmetal-api/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"

	. "github.com/onmetal/onmetal-api/apis/storage/validation"
)

var _ = Describe("VolumeTemplate", func() {
	DescribeTable("ValidateVolumeTemplateSpec",
		func(spec *storage.VolumeTemplateSpec, match types.GomegaMatcher) {
			errList := ValidateVolumeTemplateSpec(spec, field.NewPath("spec"))
			Expect(errList).To(match)
		},
		Entry("forbidden metadata name",
			&storage.VolumeTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
			ContainElement(ForbiddenField("spec.metadata.name")),
		),
		Entry("forbidden metadata namespace",
			&storage.VolumeTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
				},
			},
			ContainElement(ForbiddenField("spec.metadata.namespace")),
		),
		Entry("forbidden metadata generate name",
			&storage.VolumeTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "foo",
				},
			},
			ContainElement(ForbiddenField("spec.metadata.generateName")),
		),
	)
})
