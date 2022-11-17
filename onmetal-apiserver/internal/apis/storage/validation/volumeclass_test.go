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
	"github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/storage"
	. "github.com/onmetal/onmetal-api/onmetal-apiserver/internal/apis/storage/validation"
	. "github.com/onmetal/onmetal-api/onmetal-apiserver/internal/testutils/validation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("VolumeClass", func() {
	DescribeTable("ValidateVolumeClass",
		func(volumeClass *storage.VolumeClass, match types.GomegaMatcher) {
			errList := ValidateVolumeClass(volumeClass)
			Expect(errList).To(match)
		},
		Entry("missing name",
			&storage.VolumeClass{},
			ContainElement(RequiredField("metadata.name")),
		),
		Entry("bad name",
			&storage.VolumeClass{ObjectMeta: metav1.ObjectMeta{Name: "foo*"}},
			ContainElement(InvalidField("metadata.name")),
		),
	)

	DescribeTable("ValidateVolumeClassUpdate",
		func(newVolumeClass, oldVolumeClass *storage.VolumeClass, match types.GomegaMatcher) {
			errList := ValidateVolumeClassUpdate(newVolumeClass, oldVolumeClass)
			Expect(errList).To(match)
		},
		Entry("immutable capabilities",
			&storage.VolumeClass{
				Capabilities: map[corev1.ResourceName]resource.Quantity{
					"iops": resource.MustParse("1000"),
				},
			},
			&storage.VolumeClass{
				Capabilities: map[corev1.ResourceName]resource.Quantity{},
			},
			ContainElement(ImmutableField("capabilities")),
		),
	)
})
