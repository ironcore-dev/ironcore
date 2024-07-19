// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	volumepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/volumepoollet/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ListVolumes", func() {
	_, srv := SetupTest()
	volumeClass := SetupVolumeClass()

	It("should correctly list volumes", func(ctx SpecContext) {
		By("creating multiple volumes")
		const noOfVolumes = 3

		Volumes := make([]any, noOfVolumes)
		for i := 0; i < noOfVolumes; i++ {
			res, err := srv.CreateVolume(ctx, &iri.CreateVolumeRequest{
				Volume: &iri.Volume{
					Metadata: &irimeta.ObjectMetadata{
						Labels: map[string]string{
							volumepoolletv1alpha1.VolumeUIDLabel: "foobar",
						},
					},
					Spec: &iri.VolumeSpec{
						Class: volumeClass.Name,
						Resources: &iri.VolumeResources{
							StorageBytes: 100,
						},
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Volumes[i] = res.Volume
		}

		By("listing the Volumes")
		Expect(srv.ListVolumes(ctx, &iri.ListVolumesRequest{})).To(HaveField("Volumes", ConsistOf(Volumes...)))
	})
})
