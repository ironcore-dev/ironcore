// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"time"

	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	"github.com/ironcore-dev/ironcore/iri/apis/event/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	bucketpoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/bucketpoollet/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ListEvents", func() {
	ns, _, srv := SetupTest()
	bucketClass := SetupBucketClass("250Mi", "1500")

	FIt("should correctly list events", func(ctx SpecContext) {
		Expect(storagev1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
		k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
		})
		Expect(err).ToNot(HaveOccurred())

		By("creating bucket")
		res, err := srv.CreateBucket(ctx, &iri.CreateBucketRequest{
			Bucket: &iri.Bucket{
				Metadata: &irimeta.ObjectMetadata{
					Labels: map[string]string{
						bucketpoolletv1alpha1.BucketUIDLabel: "foobar",
					},
				},
				Spec: &iri.BucketSpec{
					Class: bucketClass.Name,
				},
			},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(res).NotTo(BeNil())

		By("getting the ironcore bucket")
		ironcoreBucket := &storagev1alpha1.Bucket{}
		ironcoreBucketKey := client.ObjectKey{Namespace: ns.Name, Name: res.Bucket.Metadata.Id}
		Expect(k8sClient.Get(ctx, ironcoreBucketKey, ironcoreBucket)).To(Succeed())

		By("generating the bucket events")
		eventGeneratedTime := time.Now()
		eventRecorder := k8sManager.GetEventRecorderFor("test-recorder")
		eventRecorder.Event(ironcoreBucket, corev1.EventTypeNormal, "testing", "this is test bucket event")

		By("listing the bucket events with no filters")
		Eventually(func(g Gomega) []*v1alpha1.Event {
			resp, err := srv.ListEvents(ctx, &iri.ListEventsRequest{})
			g.Expect(err).NotTo(HaveOccurred())
			return resp.Events
		}).Should((ConsistOf(
			HaveField("Spec", SatisfyAll(
				HaveField("InvolvedObjectMeta.Id", Equal(ironcoreBucket.Name)),
				HaveField("Reason", Equal("testing")),
				HaveField("Message", Equal("this is test bucket event")),
				HaveField("Type", Equal(corev1.EventTypeNormal)),
			)),
		)))

		By("listing the bucket events with matching label and time filters")
		resp, err := srv.ListEvents(ctx, &iri.ListEventsRequest{Filter: &iri.EventFilter{
			LabelSelector:  map[string]string{bucketpoolletv1alpha1.BucketUIDLabel: "foobar"},
			EventsFromTime: eventGeneratedTime.Unix(),
			EventsToTime:   time.Now().Unix(),
		}})

		Expect(err).NotTo(HaveOccurred())

		Expect(resp.Events).To(ConsistOf(
			HaveField("Spec", SatisfyAll(
				HaveField("InvolvedObjectMeta.Id", Equal(ironcoreBucket.Name)),
				HaveField("Reason", Equal("testing")),
				HaveField("Message", Equal("this is test bucket event")),
				HaveField("Type", Equal(corev1.EventTypeNormal)),
			)),
		),
		)

		By("listing the bucket events with non matching label filter")
		resp, err = srv.ListEvents(ctx, &iri.ListEventsRequest{Filter: &iri.EventFilter{
			LabelSelector:  map[string]string{"foo": "bar"},
			EventsFromTime: eventGeneratedTime.Unix(),
			EventsToTime:   time.Now().Unix(),
		}})
		Expect(err).NotTo(HaveOccurred())

		Expect(resp.Events).To(BeEmpty())

		By("listing the bucket events with matching label filter and non matching time filter")
		resp, err = srv.ListEvents(ctx, &iri.ListEventsRequest{Filter: &iri.EventFilter{
			LabelSelector:  map[string]string{bucketpoolletv1alpha1.BucketUIDLabel: "foobar"},
			EventsFromTime: eventGeneratedTime.Add(-10 * time.Minute).Unix(),
			EventsToTime:   eventGeneratedTime.Add(-5 * time.Minute).Unix(),
		}})
		Expect(err).NotTo(HaveOccurred())

		Expect(resp.Events).To(BeEmpty())
	})
})
