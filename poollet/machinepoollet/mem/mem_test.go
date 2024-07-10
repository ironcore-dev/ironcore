// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package mem_test

import (
	"context"
	"fmt"
	"time"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	fakemachine "github.com/ironcore-dev/ironcore/iri/testing/machine"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/controllers"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/mcm"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/mem"
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

var _ = Describe("MachineEventMapper", func() {
	var srv = &fakemachine.FakeRuntimeService{}
	ns, mp, mc := SetupTest()

	BeforeEach(func(ctx SpecContext) {
		*srv = *fakemachine.NewFakeRuntimeService()
		srv.SetMachineClasses([]*fakemachine.FakeMachineClassStatus{
			{
				MachineClassStatus: iri.MachineClassStatus{
					MachineClass: &iri.MachineClass{
						Name: mc.Name,
						Capabilities: &iri.MachineClassCapabilities{
							CpuMillis:   mc.Capabilities.CPU().MilliValue(),
							MemoryBytes: mc.Capabilities.Memory().Value(),
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

		machineClassMapper := mcm.NewGeneric(srv, mcm.GenericOptions{
			RelistPeriod: 2 * time.Second,
		})
		Expect(k8sManager.Add(machineClassMapper)).To(Succeed())

		machineEventMapper := mem.NewMachineEventMapper(k8sManager.GetClient(), srv, k8sManager.GetEventRecorderFor("test"), mem.MachineEventMapperOptions{
			RelistPeriod: 2 * time.Second,
		})
		Expect(k8sManager.Add(machineEventMapper)).To(Succeed())

		Expect((&controllers.MachineReconciler{
			EventRecorder:         &record.FakeRecorder{},
			Client:                k8sManager.GetClient(),
			MachineRuntime:        srv,
			MachineRuntimeName:    fakemachine.FakeRuntimeName,
			MachineRuntimeVersion: fakemachine.FakeVersion,
			MachineClassMapper:    machineClassMapper,
			MachinePoolName:       mp.Name,
			DownwardAPILabels: map[string]string{
				fooDownwardAPILabel: fmt.Sprintf("metadata.annotations['%s']", fooAnnotation),
			},
		}).SetupWithManager(k8sManager)).To(Succeed())

		mgrCtx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)

		go func() {
			defer GinkgoRecover()
			Expect(k8sManager.Start(mgrCtx)).To(Succeed(), "failed to start manager")
		}()
	})

	It("should get event list for machine", func(ctx SpecContext) {
		By("creating a machine")
		const fooAnnotationValue = "bar"
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    ns.Name,
				GenerateName: "machine-",
				Annotations: map[string]string{
					fooAnnotation: fooAnnotationValue,
				},
			},
			Spec: computev1alpha1.MachineSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: mc.Name},
				MachinePoolRef:  &corev1.LocalObjectReference{Name: mp.Name},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())
		By("waiting for the runtime to report the machine, volume and network interface")
		Eventually(srv).Should(SatisfyAll(
			HaveField("Machines", HaveLen(1)),
		))

		By("setting an event for iri machine")
		_, iriMachine := GetSingleMapEntry(srv.Machines)
		machineEvent := &fakemachine.FakeMachineEvents{
			MachineEvents: iri.MachineEvents{
				InvolvedObjectMeta: &v1alpha1.ObjectMetadata{
					Id:     iriMachine.Metadata.Id,
					Labels: iriMachine.Metadata.Labels,
				},
				Events: []*iri.Event{{
					Spec: &iri.EventSpec{
						Reason:    "testing",
						Message:   "this is test event",
						Type:      "Normal",
						EventTime: time.Now().Unix(),
					}}},
			},
		}
		srv.SetEvents(iriMachine.Metadata.Id, machineEvent)

		By("validating event has been emitted for correct mchine")
		machineEventList := &corev1.EventList{}
		selectorField := fields.Set{}
		selectorField["involvedObject.name"] = machine.GetName()
		Eventually(func(g Gomega) []corev1.Event {
			err := k8sClient.List(ctx, machineEventList,
				client.InNamespace(ns.Name), client.MatchingFieldsSelector{Selector: selectorField.AsSelector()},
			)
			g.Expect(err).NotTo(HaveOccurred())
			return machineEventList.Items
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
