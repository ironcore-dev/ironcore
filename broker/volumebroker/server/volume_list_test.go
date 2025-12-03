// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	volumebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/volumebroker/api/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	poolletutils "github.com/ironcore-dev/ironcore/poollet/common/utils"
	volumepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/volumepoollet/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ListVolumes", func() {
	ns, srv := SetupTest()
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

	It("should set volume uid label to all existing volumes", func(ctx SpecContext) {
		By("creating multiple volumes")
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

		ironcoreVolume := &storagev1alpha1.Volume{}
		ironcoreVolumeKey := client.ObjectKey{Namespace: ns.Name, Name: res.Volume.Metadata.Id}
		Expect(k8sClient.Get(ctx, ironcoreVolumeKey, ironcoreVolume)).To(Succeed())
		Expect(ironcoreVolume.Labels).To(Equal(map[string]string{
			poolletutils.DownwardAPILabel(volumepoolletv1alpha1.VolumeDownwardAPIPrefix, "root-volume-uid"): "foobar",
			volumebrokerv1alpha1.CreatedLabel:    "true",
			volumebrokerv1alpha1.ManagerLabel:    volumebrokerv1alpha1.VolumeBrokerManager,
			volumepoolletv1alpha1.VolumeUIDLabel: "foobar",
		}))

		base := ironcoreVolume.DeepCopy()
		delete(ironcoreVolume.Labels, volumepoolletv1alpha1.VolumeUIDLabel)
		Expect(k8sClient.Patch(ctx, ironcoreVolume, client.MergeFrom(base))).To(Succeed())
		Expect(k8sClient.Get(ctx, ironcoreVolumeKey, ironcoreVolume)).To(Succeed())
		Expect(ironcoreVolume.Labels).To(Equal(map[string]string{
			poolletutils.DownwardAPILabel(volumepoolletv1alpha1.VolumeDownwardAPIPrefix, "root-volume-uid"): "foobar",
			volumebrokerv1alpha1.CreatedLabel: "true",
			volumebrokerv1alpha1.ManagerLabel: volumebrokerv1alpha1.VolumeBrokerManager,
		}))

		Expect(srv.SetVolumeUIDLabelToAllVolumes(ctx)).NotTo(HaveOccurred())

		By("getting the ironcore volume")
		Expect(k8sClient.Get(ctx, ironcoreVolumeKey, ironcoreVolume)).To(Succeed())

		By("inspecting the ironcore volume")
		Expect(ironcoreVolume.Labels).To(Equal(map[string]string{
			poolletutils.DownwardAPILabel(volumepoolletv1alpha1.VolumeDownwardAPIPrefix, "root-volume-uid"): "foobar",
			volumebrokerv1alpha1.CreatedLabel:    "true",
			volumebrokerv1alpha1.ManagerLabel:    volumebrokerv1alpha1.VolumeBrokerManager,
			volumepoolletv1alpha1.VolumeUIDLabel: "foobar",
		}))
	})
})
