// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1_test

import (
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	. "github.com/ironcore-dev/ironcore/internal/apis/storage/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Defaults", func() {
	It("Should default the VolumeClass expansion policy if not set", func() {
		class := &storagev1alpha1.VolumeClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			},
			ResizePolicy: "",
		}
		SetDefaults_VolumeClass(class)
		Expect(class.ResizePolicy).To(Equal(storagev1alpha1.ResizePolicyStatic))
	})
})
