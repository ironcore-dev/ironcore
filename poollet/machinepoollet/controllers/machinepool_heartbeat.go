// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"sync"
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
	"sigs.k8s.io/controller-runtime/pkg/event"
)

const (
	readyReasonHeartbeatReceived  = "HeartbeatReceived"
	readyReasonRuntimeUnreachable = "RuntimeUnreachable"

	readyMessageHeartbeatReceived = "machine runtime status probe succeeded"
)

// ComputeReadyCondition returns the MachinePoolCondition that reflects the
// result of the most recent IRI Status probe for the given pool generation.
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

// MachinePoolReadyState is a thread-safe holder of the most recent IRI
// Status probe result. The MachinePoolHeartbeat publishes results into it;
// the MachinePoolReconciler reads them when computing the Ready condition.
//
// hasResult is false until the first tick has run, which lets the
// reconciler avoid clobbering whatever value is currently on the pool
// (e.g. a Ready=Unknown set by the lifecycle controller) before the
// poollet has any opinion of its own.
type MachinePoolReadyState struct {
	mu        sync.RWMutex
	hasResult bool
	probeErr  error
}

// NewMachinePoolReadyState returns a fresh, empty state.
func NewMachinePoolReadyState() *MachinePoolReadyState {
	return &MachinePoolReadyState{}
}

// Set stores the latest probe result and reports whether it differs from the
// previously stored value. The very first call always reports changed=true.
// Two errors are considered equal if their Error() strings match — that is
// what feeds into the condition's Message field.
func (s *MachinePoolReadyState) Set(probeErr error) (changed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.hasResult {
		s.hasResult = true
		s.probeErr = probeErr
		return true
	}

	if errString(s.probeErr) == errString(probeErr) {
		return false
	}

	s.probeErr = probeErr
	return true
}

// Get returns the latest probe result and whether any tick has stored one.
func (s *MachinePoolReadyState) Get() (probeErr error, hasResult bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.probeErr, s.hasResult
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// MachinePoolHeartbeat is a manager.Runnable that periodically renews the
// pool's Lease in NamespaceMachinePoolLease and publishes the latest IRI
// Status probe result into a shared MachinePoolReadyState. When the probe
// result changes (or on the first tick), it sends a generic event on Events
// so the MachinePoolReconciler — the actual writer of the pool's Status —
// reconciles the new condition.
type MachinePoolHeartbeat struct {
	Client          client.Client
	MachinePoolName string
	MachineRuntime  machine.RuntimeService

	// ReadyState receives the latest probe result on every tick.
	ReadyState *MachinePoolReadyState

	// Events is owned by the caller. The heartbeat sends a single
	// GenericEvent for the pool whenever the probe result changes, so the
	// MachinePoolReconciler picks up the change without waiting for an
	// unrelated reconcile trigger. May be nil in tests where no reconciler
	// consumes the events.
	Events chan<- event.GenericEvent

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
	readyState *MachinePoolReadyState,
	events chan<- event.GenericEvent,
	heartbeatInterval, leaseDuration, statusProbeTimeout time.Duration,
) *MachinePoolHeartbeat {
	return &MachinePoolHeartbeat{
		Client:             c,
		MachinePoolName:    machinePoolName,
		MachineRuntime:     machineRuntime,
		ReadyState:         readyState,
		Events:             events,
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

	h.publishReadyState(ctx, log, statusErr)
}

// publishReadyState stores the latest probe result and, if it differs from
// the previous one, kicks the MachinePoolReconciler via the event channel so
// the new Ready condition is reflected on the pool without waiting for an
// unrelated reconcile trigger.
func (h *MachinePoolHeartbeat) publishReadyState(ctx context.Context, log logr.Logger, statusErr error) {
	if h.ReadyState == nil {
		return
	}

	changed := h.ReadyState.Set(statusErr)
	if !changed || h.Events == nil {
		return
	}

	evt := event.GenericEvent{
		Object: &computev1alpha1.MachinePool{
			ObjectMeta: metav1.ObjectMeta{Name: h.MachinePoolName},
		},
	}
	select {
	case h.Events <- evt:
	case <-ctx.Done():
	default:
		// The reconciler is already enqueued or the channel has no
		// reader yet. Either way, the next reconcile will pick up the
		// updated state from ReadyState — dropping is safe.
		log.V(1).Info("Dropping machine pool ready-state event; channel is full or has no reader")
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
