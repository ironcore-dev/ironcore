// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	volumepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/volumepoollet/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ListVolumeSnapshots", func() {
	_, srv := SetupTest()

	It("should correctly list volume snapshots", func(ctx SpecContext) {
		By("creating multiple volume snapshots")
		const noOfVolumeSnapshots = 3

		volumeSnapshots := make([]any, noOfVolumeSnapshots)
		for i := 0; i < noOfVolumeSnapshots; i++ {
			res, err := srv.CreateVolumeSnapshot(ctx, &iri.CreateVolumeSnapshotRequest{
				VolumeSnapshot: &iri.VolumeSnapshot{
					Metadata: &irimeta.ObjectMetadata{
						Labels: map[string]string{
							volumepoolletv1alpha1.VolumeSnapshotUIDLabel: "foobar",
						},
					},
					Spec: &iri.VolumeSnapshotSpec{
						VolumeId: "test-volume",
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			volumeSnapshots[i] = res.VolumeSnapshot
		}

		By("listing the volume snapshots")
		Expect(srv.ListVolumeSnapshots(ctx, &iri.ListVolumeSnapshotsRequest{})).To(HaveField("VolumeSnapshots", ConsistOf(volumeSnapshots...)))
	})
})
