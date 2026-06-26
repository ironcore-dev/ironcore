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
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlconfig "sigs.k8s.io/controller-runtime/pkg/config"
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

	It("creates and renews the lease and sets Ready=True on success", func(ctx SpecContext) {
		leaseKey := client.ObjectKey{
			Namespace: computev1alpha1.NamespaceMachinePoolLease,
			Name:      mp.Name,
		}

		By("waiting for the lease to be created with the expected shape")
		Eventually(func(g Gomega) {
			lease := &coordinationv1.Lease{}
			g.Expect(k8sClient.Get(ctx, leaseKey, lease)).To(Succeed())
			g.Expect(lease.Spec.HolderIdentity).NotTo(BeNil())
			g.Expect(*lease.Spec.HolderIdentity).To(HavePrefix(mp.Name + "_"))
			g.Expect(lease.Spec.LeaseDurationSeconds).NotTo(BeNil())
			g.Expect(*lease.Spec.LeaseDurationSeconds).To(Equal(int32(3)))
			g.Expect(lease.Spec.RenewTime).NotTo(BeNil())
		}).Should(Succeed())

		By("waiting for the Ready condition to be set to True")
		Eventually(func(g Gomega) {
			pool := &computev1alpha1.MachinePool{}
			g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: mp.Name}, pool)).To(Succeed())
			cond := computev1alpha1.FindMachinePoolCondition(pool.Status.Conditions, computev1alpha1.MachinePoolReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(corev1.ConditionTrue))
			g.Expect(cond.Reason).To(Equal("HeartbeatReceived"))
		}).Should(Succeed())

		By("capturing the current renewTime")
		var first time.Time
		Eventually(func(g Gomega) {
			lease := &coordinationv1.Lease{}
			g.Expect(k8sClient.Get(ctx, leaseKey, lease)).To(Succeed())
			g.Expect(lease.Spec.RenewTime).NotTo(BeNil())
			first = lease.Spec.RenewTime.Time
		}).Should(Succeed())

		By("observing that the lease gets renewed")
		Eventually(func(g Gomega) {
			lease := &coordinationv1.Lease{}
			g.Expect(k8sClient.Get(ctx, leaseKey, lease)).To(Succeed())
			g.Expect(lease.Spec.RenewTime).NotTo(BeNil())
			g.Expect(lease.Spec.RenewTime.Time.After(first)).To(BeTrue(), "lease should have been renewed")
		}).Should(Succeed())
	})

	It("does not bump the pool's resourceVersion on no-op heartbeats", func(ctx SpecContext) {
		By("waiting for the first Ready=True patch and capturing the resourceVersion")
		var initialRV string
		Eventually(func(g Gomega) {
			pool := &computev1alpha1.MachinePool{}
			g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: mp.Name}, pool)).To(Succeed())
			cond := computev1alpha1.FindMachinePoolCondition(pool.Status.Conditions, computev1alpha1.MachinePoolReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(corev1.ConditionTrue))
			initialRV = pool.ResourceVersion
		}).Should(Succeed())

		By("verifying the resourceVersion does not change across subsequent ticks")
		Consistently(func(g Gomega) {
			pool := &computev1alpha1.MachinePool{}
			g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: mp.Name}, pool)).To(Succeed())
			g.Expect(pool.ResourceVersion).To(Equal(initialRV))
		}, 2*time.Second, 200*time.Millisecond).Should(Succeed())
	})

	It("flips Ready to False when the runtime probe errors", func(ctx SpecContext) {
		By("waiting for Ready=True before injecting failures")
		Eventually(func(g Gomega) {
			pool := &computev1alpha1.MachinePool{}
			g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: mp.Name}, pool)).To(Succeed())
			cond := computev1alpha1.FindMachinePoolCondition(pool.Status.Conditions, computev1alpha1.MachinePoolReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(corev1.ConditionTrue))
		}).Should(Succeed())

		By("injecting a Status error into the fake runtime")
		runner.setStatusErr(errors.New("simulated runtime down"))

		By("observing Ready flip to False with RuntimeUnreachable")
		Eventually(func(g Gomega) {
			pool := &computev1alpha1.MachinePool{}
			g.Expect(k8sClient.Get(ctx, client.ObjectKey{Name: mp.Name}, pool)).To(Succeed())
			cond := computev1alpha1.FindMachinePoolCondition(pool.Status.Conditions, computev1alpha1.MachinePoolReady)
			g.Expect(cond).NotTo(BeNil())
			g.Expect(cond.Status).To(Equal(corev1.ConditionFalse))
			g.Expect(cond.Reason).To(Equal("RuntimeUnreachable"))
			g.Expect(cond.Message).To(ContainSubstring("simulated runtime down"))
		}).Should(Succeed())
	})
})
