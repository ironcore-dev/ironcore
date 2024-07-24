// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package bem_test

import (
	"context"
	"fmt"
	"time"

	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	irievent "github.com/ironcore-dev/ironcore/iri/apis/event/v1alpha1"
	"github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	fakebucket "github.com/ironcore-dev/ironcore/iri/testing/bucket"
	"github.com/ironcore-dev/ironcore/poollet/bucketpoollet/bcm"
	"github.com/ironcore-dev/ironcore/poollet/bucketpoollet/bem"
	"github.com/ironcore-dev/ironcore/poollet/bucketpoollet/controllers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var _ = Describe("BucketEventMapper", func() {
	var srv = &fakebucket.FakeRuntimeService{}
	ns, bp, bc := SetupTest()

	BeforeEach(func(ctx SpecContext) {
		*srv = *fakebucket.NewFakeRuntimeService()
		srv.SetBucketClasses([]*fakebucket.FakeBucketClass{
			{
				BucketClass: iri.BucketClass{
					Name: bc.Name,
					Capabilities: &iri.BucketClassCapabilities{
						Tps:  262144000,
						Iops: 15000,
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

		bucketClassMapper := bcm.NewGeneric(srv, bcm.GenericOptions{
			RelistPeriod: 2 * time.Second,
		})
		Expect(k8sManager.Add(bucketClassMapper)).To(Succeed())

		bucketEventMapper := bem.NewBucketEventMapper(k8sManager.GetClient(), srv, k8sManager.GetEventRecorderFor("test"), bem.BucketEventMapperOptions{
			RelistPeriod: 2 * time.Second,
		})
		Expect(k8sManager.Add(bucketEventMapper)).To(Succeed())

		Expect((&controllers.BucketReconciler{
			EventRecorder:     &record.FakeRecorder{},
			Client:            k8sManager.GetClient(),
			BucketRuntime:     srv,
			BucketClassMapper: bucketClassMapper,
			BucketPoolName:    bp.Name,
		}).SetupWithManager(k8sManager)).To(Succeed())

		mgrCtx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)

		go func() {
			defer GinkgoRecover()
			Expect(k8sManager.Start(mgrCtx)).To(Succeed(), "failed to start manager")
		}()
	})

	It("should get event list for bucket", func(ctx SpecContext) {
		By("creating a bucket")
		bucket := &storagev1alpha1.Bucket{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "bucket-",
			},
			Spec: storagev1alpha1.BucketSpec{
				BucketClassRef: &corev1.LocalObjectReference{Name: bc.Name},
				BucketPoolRef:  &corev1.LocalObjectReference{Name: bp.Name},
			},
		}
		Expect(k8sClient.Create(ctx, bucket)).To(Succeed())

		By("waiting for the runtime to report the bucket")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Buckets", HaveLen(1)),
		))
		_, iriBucket := GetSingleMapEntry(srv.Buckets)
		By("setting an event for iri bucket")
		eventList := []*fakebucket.FakeEvent{{
			Event: irievent.Event{
				Spec: &irievent.EventSpec{
					InvolvedObjectMeta: &v1alpha1.ObjectMetadata{
						Labels: iriBucket.Metadata.Labels,
					},
					Reason:    "testing",
					Message:   "this is test bucket event",
					Type:      "Normal",
					EventTime: time.Now().Unix(),
				}},
		},
		}
		srv.SetEvents(eventList)

		By("validating event has been emitted for correct bucket")
		bucketEventList := &corev1.EventList{}
		selectorField := fields.Set{}
		selectorField["involvedObject.name"] = bucket.GetName()
		Eventually(func(g Gomega) []corev1.Event {
			err := k8sClient.List(ctx, bucketEventList,
				client.InNamespace(ns.Name), client.MatchingFieldsSelector{Selector: selectorField.AsSelector()},
			)
			g.Expect(err).NotTo(HaveOccurred())
			return bucketEventList.Items
		}).Should(ContainElement(SatisfyAll(
			HaveField("Reason", Equal("testing")),
			HaveField("Message", Equal("this is test bucket event")),
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
