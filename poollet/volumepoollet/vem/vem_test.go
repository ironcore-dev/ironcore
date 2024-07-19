// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package vem_test

import (
	"context"
	"fmt"
	"time"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	irievent "github.com/ironcore-dev/ironcore/iri/apis/event/v1alpha1"
	"github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"

	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	fakevolume "github.com/ironcore-dev/ironcore/iri/testing/volume"
	"github.com/ironcore-dev/ironcore/poollet/volumepoollet/controllers"
	"github.com/ironcore-dev/ironcore/poollet/volumepoollet/vcm"
	"github.com/ironcore-dev/ironcore/poollet/volumepoollet/vem"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var _ = Describe("VolumeEventMapper", func() {
	var srv = &fakevolume.FakeRuntimeService{}
	ns, vp, vc := SetupTest()

	BeforeEach(func(ctx SpecContext) {
		*srv = *fakevolume.NewFakeRuntimeService()
		srv.SetVolumeClasses([]*fakevolume.FakeVolumeClassStatus{
			{
				VolumeClassStatus: iri.VolumeClassStatus{
					VolumeClass: &iri.VolumeClass{
						Name: vc.Name,
						Capabilities: &iri.VolumeClassCapabilities{
							Tps:  vc.Capabilities.TPS().Value(),
							Iops: vc.Capabilities.IOPS().Value(),
						},
					},
				},
			},
		})

		k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
			Metrics: metricserver.Options{
				BindAddress: "0",
			},
		})
		Expect(err).ToNot(HaveOccurred())

		volumeClassMapper := vcm.NewGeneric(srv, vcm.GenericOptions{
			RelistPeriod: 2 * time.Second,
		})
		Expect(k8sManager.Add(volumeClassMapper)).To(Succeed())

		volumeEventMapper := vem.NewVolumeEventMapper(k8sManager.GetClient(), srv, k8sManager.GetEventRecorderFor("test"), vem.VolumeEventMapperOptions{
			RelistPeriod: 2 * time.Second,
		})
		Expect(k8sManager.Add(volumeEventMapper)).To(Succeed())

		Expect((&controllers.VolumeReconciler{
			EventRecorder:     &record.FakeRecorder{},
			Client:            k8sManager.GetClient(),
			VolumeRuntime:     srv,
			VolumeClassMapper: volumeClassMapper,
			VolumePoolName:    vp.Name,
		}).SetupWithManager(k8sManager)).To(Succeed())

		mgrCtx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)

		go func() {
			defer GinkgoRecover()
			Expect(k8sManager.Start(mgrCtx)).To(Succeed(), "failed to start manager")
		}()
	})

	It("should get event list for volume", func(ctx SpecContext) {
		By("creating a volume")
		const fooAnnotationValue = "bar"
		volume := &storagev1alpha1.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "volume-",
				Annotations: map[string]string{
					fooAnnotation: fooAnnotationValue,
				},
			},
			Spec: storagev1alpha1.VolumeSpec{
				Resources: corev1alpha1.ResourceList{
					corev1alpha1.ResourceStorage: resource.MustParse("250"),
				},
				VolumeClassRef: &corev1.LocalObjectReference{Name: vc.Name},
				VolumePoolRef:  &corev1.LocalObjectReference{Name: vp.Name},
			},
		}
		Expect(k8sClient.Create(ctx, volume)).To(Succeed(), "failed to create volume")

		By("waiting for the runtime to report the volume")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Volumes", HaveLen(1)),
		))
		_, iriVolume := GetSingleMapEntry(srv.Volumes)

		By("setting an event for iri volume")
		eventList := []*fakevolume.FakeEvent{{
			Event: irievent.Event{
				Spec: &irievent.EventSpec{
					InvolvedObjectMeta: &v1alpha1.ObjectMetadata{
						Labels: iriVolume.Metadata.Labels,
					},
					Reason:    "testing",
					Message:   "this is test event",
					Type:      "Normal",
					EventTime: time.Now().Unix(),
				}},
		},
		}
		srv.SetEvents(eventList)

		By("validating event has been emitted for correct volume")
		volumeEventList := &corev1.EventList{}
		selectorField := fields.Set{}
		selectorField["involvedObject.name"] = volume.GetName()
		Eventually(func(g Gomega) []corev1.Event {
			err := k8sClient.List(ctx, volumeEventList,
				client.InNamespace(ns.Name), client.MatchingFieldsSelector{Selector: selectorField.AsSelector()},
			)
			g.Expect(err).NotTo(HaveOccurred())
			return volumeEventList.Items
		}).Should(ContainElement(SatisfyAll(
			HaveField("Reason", Equal("testing")),
			HaveField("Message", Equal("this is test event")),
			HaveField("Type", Equal(corev1.EventTypeNormal)),
		)))
	})
})

func GetSingleMapEntry[K comparable, V any](m map[K]V) (K, V) {
	if n := len(m); n != 1 {
		Fail(fmt.Sprintf("Expected for map to have a single entry but got %d", n), 1)
	}
	for k, v := range m {
		return k, v
	}
	panic("unreachable")
}
