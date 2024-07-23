// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"time"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	irievent "github.com/ironcore-dev/ironcore/iri/apis/event/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	machinepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ListEvents", func() {
	ns, srv := SetupTest()
	machineClass := SetupMachineClass()

	It("should correctly list events", func(ctx SpecContext) {
		By("creating machine")
		Expect(computev1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
		Expect(networkingv1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())

		k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
		})
		Expect(err).ToNot(HaveOccurred())

		res, err := srv.CreateMachine(ctx, &iri.CreateMachineRequest{
			Machine: &iri.Machine{
				Metadata: &irimeta.ObjectMetadata{
					Labels: map[string]string{
						machinepoolletv1alpha1.MachineUIDLabel: "foobar",
					},
				},
				Spec: &iri.MachineSpec{
					Power: iri.Power_POWER_ON,
					Image: &iri.ImageSpec{
						Image: "example.org/foo:latest",
					},
					Class: machineClass.Name,
					NetworkInterfaces: []*iri.NetworkInterface{
						{
							Name:      "primary-nic",
							NetworkId: "network-id",
							Ips:       []string{"10.0.0.1"},
						},
					},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(res).NotTo(BeNil())

		By("getting the ironcore machine")
		ironcoreMachine := &computev1alpha1.Machine{}
		ironcoreMachineKey := client.ObjectKey{Namespace: ns.Name, Name: res.Machine.Metadata.Id}
		Expect(k8sClient.Get(ctx, ironcoreMachineKey, ironcoreMachine)).To(Succeed())

		By("generating the machine events")
		eventGeneratedTime := time.Now()
		eventRecorder := k8sManager.GetEventRecorderFor("test-recorder")
		eventRecorder.Event(ironcoreMachine, corev1.EventTypeNormal, "testing", "this is test event")

		By("listing the machine events with no filters")
		Eventually(func(g Gomega) []*irievent.Event {
			resp, err := srv.ListEvents(ctx, &iri.ListEventsRequest{})
			g.Expect(err).NotTo(HaveOccurred())
			return resp.Events
		}).Should(ConsistOf(
			HaveField("Spec", SatisfyAll(
				HaveField("InvolvedObjectMeta.Id", Equal(ironcoreMachine.Name)),
				HaveField("Reason", Equal("testing")),
				HaveField("Message", Equal("this is test event")),
				HaveField("Type", Equal(corev1.EventTypeNormal)),
			)),
		),
		)

		By("listing the machine events with matching label and time filters")
		resp, err := srv.ListEvents(ctx, &iri.ListEventsRequest{Filter: &iri.EventFilter{
			LabelSelector:  map[string]string{machinepoolletv1alpha1.MachineUIDLabel: "foobar"},
			EventsFromTime: eventGeneratedTime.Unix(),
			EventsToTime:   time.Now().Unix(),
		}})

		Expect(err).NotTo(HaveOccurred())

		Expect(resp.Events).To(ConsistOf(
			HaveField("Spec", SatisfyAll(
				HaveField("InvolvedObjectMeta.Id", Equal(ironcoreMachine.Name)),
				HaveField("Reason", Equal("testing")),
				HaveField("Message", Equal("this is test event")),
				HaveField("Type", Equal(corev1.EventTypeNormal)),
			)),
		),
		)

		By("listing the machine events with non matching label filter")
		resp, err = srv.ListEvents(ctx, &iri.ListEventsRequest{Filter: &iri.EventFilter{
			LabelSelector:  map[string]string{"foo": "bar"},
			EventsFromTime: eventGeneratedTime.Unix(),
			EventsToTime:   time.Now().Unix(),
		}})
		Expect(err).NotTo(HaveOccurred())

		Expect(resp.Events).To(BeEmpty())

		By("listing the machine events with matching label filter and non matching time filter")
		resp, err = srv.ListEvents(ctx, &iri.ListEventsRequest{Filter: &iri.EventFilter{
			LabelSelector:  map[string]string{machinepoolletv1alpha1.MachineUIDLabel: "foobar"},
			EventsFromTime: eventGeneratedTime.Add(-10 * time.Minute).Unix(),
			EventsToTime:   eventGeneratedTime.Add(-5 * time.Minute).Unix(),
		}})
		Expect(err).NotTo(HaveOccurred())

		Expect(resp.Events).To(BeEmpty())
	})
})
