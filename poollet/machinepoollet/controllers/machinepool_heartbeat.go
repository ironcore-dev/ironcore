// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"github.com/ironcore-dev/ironcore/iri/apis/machine"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	readyReasonHeartbeatReceived  = "HeartbeatReceived"
	readyReasonRuntimeUnreachable = "RuntimeUnreachable"

	readyMessageHeartbeatReceived = "machine runtime status probe succeeded"
)

// ComputeReadyCondition returns the MachinePoolCondition the heartbeat wants
// to put on the pool, given the current pool generation and the result of
// the most recent IRI Status probe.
func ComputeReadyCondition(generation int64, probeErr error) computev1alpha1.MachinePoolCondition {
	if probeErr != nil {
		return computev1alpha1.MachinePoolCondition{
			Type:               computev1alpha1.MachinePoolReady,
			Status:             corev1.ConditionFalse,
			Reason:             readyReasonRuntimeUnreachable,
			Message:            probeErr.Error(),
			ObservedGeneration: generation,
		}
	}
	return computev1alpha1.MachinePoolCondition{
		Type:               computev1alpha1.MachinePoolReady,
		Status:             corev1.ConditionTrue,
		Reason:             readyReasonHeartbeatReceived,
		Message:            readyMessageHeartbeatReceived,
		ObservedGeneration: generation,
	}
}

// ReadyConditionsDiffer reports whether existing differs from desired in any
// field that justifies a status patch. LastUpdateTime and LastTransitionTime
// are intentionally ignored — they are bookkeeping that should only advance
// when something else also advances. A nil existing always differs.
func ReadyConditionsDiffer(existing *computev1alpha1.MachinePoolCondition, desired computev1alpha1.MachinePoolCondition) bool {
	if existing == nil {
		return true
	}
	return existing.Status != desired.Status ||
		existing.Reason != desired.Reason ||
		existing.Message != desired.Message ||
		existing.ObservedGeneration != desired.ObservedGeneration
}

// MachinePoolHeartbeat is a manager.Runnable that periodically renews the
// pool's Lease in NamespaceMachinePoolLease and updates the Ready condition
// on the MachinePool's status. It is the poollet side of IEP-15.
type MachinePoolHeartbeat struct {
	Client          client.Client
	MachinePoolName string
	MachineRuntime  machine.RuntimeService

	HeartbeatInterval  time.Duration
	LeaseDuration      time.Duration
	StatusProbeTimeout time.Duration

	holderIdentity string
}

// NewMachinePoolHeartbeat constructs a heartbeat runnable. The holderIdentity
// is fixed for the process lifetime: <poolName>_<uuid>.
func NewMachinePoolHeartbeat(
	c client.Client,
	machinePoolName string,
	machineRuntime machine.RuntimeService,
	heartbeatInterval, leaseDuration, statusProbeTimeout time.Duration,
) *MachinePoolHeartbeat {
	return &MachinePoolHeartbeat{
		Client:             c,
		MachinePoolName:    machinePoolName,
		MachineRuntime:     machineRuntime,
		HeartbeatInterval:  heartbeatInterval,
		LeaseDuration:      leaseDuration,
		StatusProbeTimeout: statusProbeTimeout,
		holderIdentity:     fmt.Sprintf("%s_%s", machinePoolName, uuid.NewString()),
	}
}

// Start runs the heartbeat loop until ctx is canceled. It satisfies
// sigs.k8s.io/controller-runtime/pkg/manager.Runnable.
func (h *MachinePoolHeartbeat) Start(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx).WithName("machinepool-heartbeat")
	log.Info("Starting machine pool heartbeat",
		"interval", h.HeartbeatInterval,
		"leaseDuration", h.LeaseDuration,
		"holderIdentity", h.holderIdentity,
	)

	h.tick(ctx, log)

	ticker := time.NewTicker(h.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Info("Stopping machine pool heartbeat")
			return nil
		case <-ticker.C:
			h.tick(ctx, log)
		}
	}
}

func (h *MachinePoolHeartbeat) tick(ctx context.Context, log logr.Logger) {
	probeCtx, cancel := context.WithTimeout(ctx, h.StatusProbeTimeout)
	_, statusErr := h.MachineRuntime.Status(probeCtx, &iri.StatusRequest{})
	cancel()

	if err := h.reconcileLease(ctx); err != nil && ctx.Err() == nil {
		log.Error(err, "Failed to reconcile machine pool lease")
	}
	if err := h.reconcileReadyCondition(ctx, statusErr); err != nil && ctx.Err() == nil {
		log.Error(err, "Failed to reconcile machine pool ready condition")
	}
}

func (h *MachinePoolHeartbeat) reconcileLease(ctx context.Context) error {
	leaseDurationSeconds := int32(h.LeaseDuration.Seconds())
	now := metav1.NewMicroTime(time.Now())

	lease := &coordinationv1.Lease{}
	key := client.ObjectKey{Namespace: computev1alpha1.NamespaceMachinePoolLease, Name: h.MachinePoolName}
	if err := h.Client.Get(ctx, key, lease); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("getting lease: %w", err)
		}
		newLease := &coordinationv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: key.Namespace,
				Name:      key.Name,
			},
			Spec: coordinationv1.LeaseSpec{
				HolderIdentity:       ptr.To(h.holderIdentity),
				LeaseDurationSeconds: ptr.To(leaseDurationSeconds),
				AcquireTime:          ptr.To(now),
				RenewTime:            ptr.To(now),
			},
		}
		if err := h.Client.Create(ctx, newLease); err != nil {
			return fmt.Errorf("creating lease: %w", err)
		}
		return nil
	}

	base := lease.DeepCopy()
	if lease.Spec.HolderIdentity == nil || *lease.Spec.HolderIdentity != h.holderIdentity {
		// Take ownership; the previous owner (likely a previous poollet process for
		// this pool) is gone or stale.
		log := ctrl.LoggerFrom(ctx)
		previousHolder := "<unset>"
		if lease.Spec.HolderIdentity != nil {
			previousHolder = *lease.Spec.HolderIdentity
		}
		log.Info("Taking ownership of stale machine pool lease", "previousHolder", previousHolder, "newHolder", h.holderIdentity)
		lease.Spec.HolderIdentity = ptr.To(h.holderIdentity)
		lease.Spec.AcquireTime = ptr.To(now)
	}
	lease.Spec.LeaseDurationSeconds = ptr.To(leaseDurationSeconds)
	lease.Spec.RenewTime = ptr.To(now)

	if err := h.Client.Patch(ctx, lease, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("patching lease: %w", err)
	}
	return nil
}

func (h *MachinePoolHeartbeat) reconcileReadyCondition(ctx context.Context, statusErr error) error {
	pool := &computev1alpha1.MachinePool{}
	if err := h.Client.Get(ctx, client.ObjectKey{Name: h.MachinePoolName}, pool); err != nil {
		return fmt.Errorf("getting machine pool: %w", err)
	}

	desired := ComputeReadyCondition(pool.Generation, statusErr)
	existing := computev1alpha1.FindMachinePoolCondition(pool.Status.Conditions, computev1alpha1.MachinePoolReady)
	if !ReadyConditionsDiffer(existing, desired) {
		return nil
	}

	base := pool.DeepCopy()
	// Currently only the heartbeat sets conditions, so MergeFrom patch is sufficient.
	pool.Status.Conditions = computev1alpha1.SetMachinePoolCondition(pool.Status.Conditions, desired)
	if err := h.Client.Status().Patch(ctx, pool, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("patching machine pool ready condition: %w", err)
	}
	return nil
}
