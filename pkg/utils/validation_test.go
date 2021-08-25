/*
 * Copyright (c) 2021 by the OnMetal authors.
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

package utils

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("Utils", func() {
	s1 := "a"
	s2 := "b"
	fieldPath := field.NewPath("spec").Child("foo")

	Context("When validating two strings", func() {
		It("Should be successful if the strings are the same", func() {
			Expect(ValidateFieldIsImmutatable(s1, s1, fieldPath)).Should(BeNil())
		})
		It("Should be unsuccessful if the strings are different", func() {
			err := field.Invalid(fieldPath, s1, FieldIsImmutable)
			Expect(ValidateFieldIsImmutatable(s1, s2, fieldPath)).Should(Equal(err))
		})
	})
})
