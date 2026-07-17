// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	computeclient "github.com/ironcore-dev/ironcore/internal/client/compute"
	"github.com/ironcore-dev/ironcore/iri/apis/machine"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	fakemachine "github.com/ironcore-dev/ironcore/iri/testing/machine"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/controllers"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/mcm"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlconfig "sigs.k8s.io/controller-runtime/pkg/config"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
	"sigs.k8s.io/controller-runtime/pkg/event"
	metricserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

// errorInjectingRuntime wraps a real fake runtime and lets a test toggle
// errors on the Status method only. All other calls pass through.
type errorInjectingRuntime struct {
	machine.RuntimeService
	statusErr atomic.Pointer[error]
}

func (r *errorInjectingRuntime) Status(ctx context.Context, req *iri.StatusRequest) (*iri.StatusResponse, error) {
	if errPtr := r.statusErr.Load(); errPtr != nil && *errPtr != nil {
		return nil, *errPtr
	}
	return r.RuntimeService.Status(ctx, req)
}

func (r *errorInjectingRuntime) setStatusErr(err error) {
	if err == nil {
		r.statusErr.Store(nil)
		return
	}
	r.statusErr.Store(&err)
}

var _ = Describe("MachinePoolHeartbeat", func() {
	var (
		mp     *computev1alpha1.MachinePool
		runner *errorInjectingRuntime
	)

	BeforeEach(func(ctx SpecContext) {
		By("creating a pool for this spec")
		mp = &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "heartbeat-mp-"},
		}
		Expect(k8sClient.Create(ctx, mp)).To(Succeed())
		DeferCleanup(k8sClient.Delete, mp)

		fake := fakemachine.NewFakeRuntimeService()
		runner = &errorInjectingRuntime{RuntimeService: fake}

		// Stand up a manager just for this spec — keep it isolated from the
		// shared SetupTest manager so we control which runnables are added.
		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme:     scheme.Scheme,
			Metrics:    metricserver.Options{BindAddress: "0"},
			Controller: ctrlconfig.Controller{SkipNameValidation: ptr.To(true)},
		})
		Expect(err).NotTo(HaveOccurred())

		// The reconciler needs the same indexer the production wiring sets up.
		Expect(computeclient.SetupMachineSpecMachinePoolRefNameFieldIndexer(ctx, mgr.GetFieldIndexer())).To(Succeed())

		machineClassMapper := mcm.NewGeneric(runner, mcm.GenericOptions{
			RelistPeriod: 2 * time.Second,
		})
		Expect(mgr.Add(machineClassMapper)).To(Succeed())

		readyState := controllers.NewMachinePoolReadyState()
		heartbeatEvents := make(chan event.GenericEvent, 1)

		Expect((&controllers.MachinePoolReconciler{
			Client:             mgr.GetClient(),
			MachinePoolName:    mp.Name,
			MachineRuntime:     runner,
			MachineClassMapper: machineClassMapper,
			TopologyLabels:     map[commonv1alpha1.TopologyLabel]string{},
			ReadyState:         readyState,
			HeartbeatEvents:    heartbeatEvents,
		}).SetupWithManager(mgr)).To(Succeed())

		hb := controllers.NewMachinePoolHeartbeat(
			mgr.GetClient(), mp.Name, runner,
			readyState, heartbeatEvents,
			500*time.Millisecond, // interval
			3*time.Second,        // lease duration
			200*time.Millisecond, // status probe timeout
		)
		Expect(mgr.Add(hb)).To(Succeed())

		mgrCtx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)
		go func() {
			defer GinkgoRecover()
			Expect(mgr.Start(mgrCtx)).To(Succeed())
		}()
	})

	It("creates and renews the lease and sets Ready=True on success", func() {
		lease := &coordinationv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: computev1alpha1.NamespaceMachinePoolLease,
				Name:      mp.Name,
			},
		}

		By("waiting for the lease to be created with the expected shape")
		Eventually(Object(lease)).Should(SatisfyAll(
			HaveField("Spec.HolderIdentity", Not(BeNil())),
			HaveField("Spec.HolderIdentity", PointTo(HavePrefix(mp.Name+"_"))),
			HaveField("Spec.LeaseDurationSeconds", Not(BeNil())),
			HaveField("Spec.LeaseDurationSeconds", PointTo(Equal(int32(3)))),
			HaveField("Spec.RenewTime", Not(BeNil())),
		))

		pool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{Name: mp.Name},
		}

		By("waiting for the Ready condition to be set to True")
		Eventually(Object(pool)).Should(
			HaveField("Status.Conditions", ContainElement(SatisfyAll(
				HaveField("Type", computev1alpha1.MachinePoolReady),
				HaveField("Status", corev1.ConditionTrue),
				HaveField("Reason", Equal("HeartbeatReceived")),
			))),
		)

		By("capturing the current renewTime")
		var first time.Time
		Eventually(func(g Gomega) {
			g.Expect(Object(lease)()).To(HaveField("Spec.RenewTime", Not(BeNil())))
			first = lease.Spec.RenewTime.Time
		}).Should(Succeed())

		By("observing that the lease gets renewed")
		Eventually(func(g Gomega) {
			g.Expect(Object(lease)()).To(HaveField("Spec.RenewTime", Not(BeNil())))
			g.Expect(lease.Spec.RenewTime.Time.After(first)).To(BeTrue(), "lease should have been renewed")
		}).Should(Succeed())
	})

	It("does not bump the pool's resourceVersion on no-op heartbeats", func() {
		pool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{Name: mp.Name},
		}

		By("waiting for the first Ready=True patch and capturing the resourceVersion")
		Eventually(Object(pool)).Should(
			HaveField("Status.Conditions", ContainElement(SatisfyAll(
				HaveField("Type", computev1alpha1.MachinePoolReady),
				HaveField("Status", corev1.ConditionTrue),
			))),
		)
		initialRV := pool.ResourceVersion

		By("verifying the resourceVersion does not change across subsequent ticks")
		Consistently(Object(pool)).Should(
			HaveField("ResourceVersion", Equal(initialRV)),
		)
	})

	It("flips Ready to False when the runtime probe errors", func() {
		pool := &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{Name: mp.Name},
		}

		By("waiting for Ready=True before injecting failures")
		Eventually(Object(pool)).Should(
			HaveField("Status.Conditions", ContainElement(SatisfyAll(
				HaveField("Type", computev1alpha1.MachinePoolReady),
				HaveField("Status", corev1.ConditionTrue),
			))),
		)

		By("injecting a Status error into the fake runtime")
		runner.setStatusErr(errors.New("simulated runtime down"))

		By("observing Ready flip to False with RuntimeUnreachable")
		Eventually(Object(pool)).Should(
			HaveField("Status.Conditions", ContainElement(SatisfyAll(
				HaveField("Type", computev1alpha1.MachinePoolReady),
				HaveField("Status", corev1.ConditionFalse),
				HaveField("Reason", Equal("RuntimeUnreachable")),
				HaveField("Message", ContainSubstring("simulated runtime down")),
			))),
		)
	})
})
