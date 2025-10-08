// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	volumepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/volumepoollet/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeleteVolumeSnapshot", func() {
	ns, srv := SetupTest()

	It("should correctly delete a volume snapshot", func(ctx SpecContext) {
		By("creating a volume snapshot")
		createRes, err := srv.CreateVolumeSnapshot(ctx, &iri.CreateVolumeSnapshotRequest{
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
		Expect(createRes).NotTo(BeNil())

		By("deleting the volume snapshot")
		deleteRes, err := srv.DeleteVolumeSnapshot(ctx, &iri.DeleteVolumeSnapshotRequest{
			VolumeSnapshotId: createRes.VolumeSnapshot.Metadata.Id,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(deleteRes).NotTo(BeNil())

		By("verifying the volume snapshot is deleted")
		ironcoreVolumeSnapshot := &storagev1alpha1.VolumeSnapshot{}
		ironcoreVolumeSnapshotKey := client.ObjectKey{Namespace: ns.Name, Name: createRes.VolumeSnapshot.Metadata.Id}
		err = k8sClient.Get(ctx, ironcoreVolumeSnapshotKey, ironcoreVolumeSnapshot)
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})

})
