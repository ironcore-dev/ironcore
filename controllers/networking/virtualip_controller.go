/*
 * Copyright (c) 2021 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package networking

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	onmetalapiclient "github.com/onmetal/onmetal-api/client"
	"github.com/onmetal/onmetal-api/controllers/networking/events"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// VirtualIPReconciler reconciles a VirtualIP object
type VirtualIPReconciler struct {
	record.EventRecorder
	client.Client
	APIReader client.Reader
	Scheme    *runtime.Scheme
	// BindTimeout is the maximum duration until a VirtualIP's Bound condition is considered to be timed out.
	BindTimeout time.Duration
}

//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualips,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualips/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualips/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networkinterfaces,verbs=get;list

// Reconcile is part of the main reconciliation loop for VirtualIP types
func (r *VirtualIPReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	virtualIP := &networkingv1alpha1.VirtualIP{}
	if err := r.Get(ctx, req.NamespacedName, virtualIP); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return r.reconcileExists(ctx, log, virtualIP)
}

func (r *VirtualIPReconciler) reconcileExists(ctx context.Context, log logr.Logger, virtualIP *networkingv1alpha1.VirtualIP) (ctrl.Result, error) {
	if !virtualIP.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, virtualIP)
	}
	return r.reconcile(ctx, log, virtualIP)
}

func (r *VirtualIPReconciler) delete(ctx context.Context, log logr.Logger, virtualIP *networkingv1alpha1.VirtualIP) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *VirtualIPReconciler) phaseTransitionTimedOut(timestamp *metav1.Time) bool {
	if timestamp.IsZero() {
		return false
	}
	return timestamp.Add(r.BindTimeout).Before(time.Now())
}

func (r *VirtualIPReconciler) setTargetRefUID(ctx context.Context, virtualIP *networkingv1alpha1.VirtualIP, uid types.UID) error {
	base := virtualIP.DeepCopy()
	virtualIP.Spec.TargetRef.UID = uid
	if err := r.Patch(ctx, virtualIP, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error setting target ref uid: %w", err)
	}
	return nil
}

func (r *VirtualIPReconciler) reconcile(ctx context.Context, log logr.Logger, virtualIP *networkingv1alpha1.VirtualIP) (ctrl.Result, error) {
	log.V(1).Info("Reconciling virtual ip")
	if virtualIP.Spec.TargetRef == nil {
		return r.reconcileUnbound(ctx, log, virtualIP)
	}

	return r.reconcileBound(ctx, log, virtualIP)
}

func (r *VirtualIPReconciler) reconcileBound(ctx context.Context, log logr.Logger, virtualIP *networkingv1alpha1.VirtualIP) (ctrl.Result, error) {
	nic := &networkingv1alpha1.NetworkInterface{}
	nicKey := client.ObjectKey{
		Namespace: virtualIP.Namespace,
		Name:      virtualIP.Spec.TargetRef.Name,
	}
	log = log.WithValues("NetworkInterfaceKey", nicKey, "TargetType", "NetworkInterface")
	log.V(1).Info("VirtualIP references target")
	// We have to use APIReader here as stale data might cause unbinding a virtual ip for a short duration.
	err := r.APIReader.Get(ctx, nicKey, nic)
	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, fmt.Errorf("error getting network interface %s: %w", nicKey, err)
	}

	if err == nil && virtualIP.Spec.TargetRef.UID == "" {
		log = log.WithValues("TargetUID", nic.UID)
		log.V(1).Info("Setting target ref uid")
		if err := r.setTargetRefUID(ctx, virtualIP, nic.UID); err != nil {
			return ctrl.Result{}, err
		}

		log.V(1).Info("Set target ref uid")
		return ctrl.Result{}, nil
	}

	nicExists := err == nil
	validReferences := nicExists && r.validReferences(virtualIP, nic)
	phase := virtualIP.Status.Phase
	phaseLastTransitionTime := virtualIP.Status.LastPhaseTransitionTime

	log = log.WithValues(
		"NetworkInterfaceExists", nicExists,
		"ValidReferences", validReferences,
		"Phase", phase,
		"PhaseLastTransitionTime", phaseLastTransitionTime,
	)

	if !nicExists {
		r.Eventf(virtualIP, corev1.EventTypeWarning, events.FailedBindingNetworkInterface, "Network interface %s not found", nicKey.Name)
	}

	switch {
	case validReferences:
		log.V(1).Info("Setting virtual ip to bound")
		if err := r.patchStatus(ctx, virtualIP, networkingv1alpha1.VirtualIPPhaseBound); err != nil {
			return ctrl.Result{}, fmt.Errorf("error binding virtualip: %w", err)
		}

		log.V(1).Info("Successfully set virtual ip to bound.")
		return ctrl.Result{}, nil
	case !validReferences && phase == networkingv1alpha1.VirtualIPPhasePending && r.phaseTransitionTimedOut(phaseLastTransitionTime):
		log.V(1).Info("Bind is not ok and timed out, releasing virtual ip")
		if err := r.release(ctx, virtualIP); err != nil {
			return ctrl.Result{}, fmt.Errorf("error releasing virtualip: %w", err)
		}

		log.V(1).Info("Successfully released virtual ip")
		return ctrl.Result{}, nil
	default:
		log.V(1).Info("Bind is not ok and not yet timed out, setting to pending")
		if err := r.patchStatus(ctx, virtualIP, networkingv1alpha1.VirtualIPPhasePending); err != nil {
			return ctrl.Result{}, fmt.Errorf("error setting virtualip to pending: %w", err)
		}

		log.V(1).Info("Successfully set virtual ip to pending")
		return r.requeueAfterBoundTimeout(virtualIP), nil
	}
}

func (r *VirtualIPReconciler) reconcileUnbound(ctx context.Context, log logr.Logger, virtualIP *networkingv1alpha1.VirtualIP) (ctrl.Result, error) {
	log.V(1).Info("Reconcile unbound")

	if virtualIP.Status.IP == nil {
		log.V(1).Info("No ip assigned, marking as phase unbound")
		if err := r.patchStatus(ctx, virtualIP, networkingv1alpha1.VirtualIPPhaseUnbound); err != nil {
			return ctrl.Result{}, err
		}

		log.V(1).Info("Successfully marked as phase unbound")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("IP assigned, searching for suitable requester")
	nic, err := r.getMatchingNetworkInterface(ctx, virtualIP)
	if err != nil {
		return ctrl.Result{}, err
	}

	if nic == nil {
		log.V(1).Info("No requester found, setting to phase unbound")
		if err := r.patchStatus(ctx, virtualIP, networkingv1alpha1.VirtualIPPhaseUnbound); err != nil {
			return ctrl.Result{}, err
		}
		log.V(1).Info("Successfully set to phase unbound")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("NetworkInterfaceKey", client.ObjectKeyFromObject(nic))
	log.V(1).Info("Found a matching requester, assigning to it")
	if err := r.assign(ctx, virtualIP, nic); err != nil {
		return ctrl.Result{}, err
	}

	log.V(1).Info("Successfully assigned")
	return ctrl.Result{}, nil
}

func (r *VirtualIPReconciler) assign(ctx context.Context, virtualIP *networkingv1alpha1.VirtualIP, nic *networkingv1alpha1.NetworkInterface) error {
	base := virtualIP.DeepCopy()
	virtualIP.Spec.TargetRef = &commonv1alpha1.LocalUIDReference{Name: nic.Name, UID: nic.UID}
	if err := r.Patch(ctx, virtualIP, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error assigning to network interface: %w", err)
	}
	return nil
}

func (r *VirtualIPReconciler) getMatchingNetworkInterface(ctx context.Context, virtualIP *networkingv1alpha1.VirtualIP) (*networkingv1alpha1.NetworkInterface, error) {
	nicList := &networkingv1alpha1.NetworkInterfaceList{}
	if err := r.List(ctx, nicList,
		client.InNamespace(virtualIP.Namespace),
		client.MatchingFields{onmetalapiclient.NetworkInterfaceVirtualIPNames: virtualIP.Name},
	); err != nil {
		return nil, fmt.Errorf("error listing suitable requesters: %w", err)
	}

	var matches []networkingv1alpha1.NetworkInterface
	for _, nic := range nicList.Items {
		if !nic.DeletionTimestamp.IsZero() {
			continue
		}

		if r.networkInterfaceReferencesVirtualIP(virtualIP, &nic) {
			matches = append(matches, nic)
		}
	}
	if len(matches) == 0 {
		return nil, nil
	}
	match := matches[rand.Intn(len(matches))]
	return &match, nil
}

func (r *VirtualIPReconciler) networkInterfaceReferencesVirtualIP(virtualIP *networkingv1alpha1.VirtualIP, nic *networkingv1alpha1.NetworkInterface) bool {
	nicVirtualIP := nic.Spec.VirtualIP
	if nicVirtualIP == nil {
		return false
	}

	return networkingv1alpha1.NetworkInterfaceVirtualIPName(nic.Name, *nicVirtualIP) == virtualIP.Name
}

func (r *VirtualIPReconciler) requeueAfterBoundTimeout(virtualIP *networkingv1alpha1.VirtualIP) ctrl.Result {
	boundTimeoutExpirationDuration := time.Until(virtualIP.Status.LastPhaseTransitionTime.Add(r.BindTimeout)).Round(time.Second)
	if boundTimeoutExpirationDuration <= 0 {
		return ctrl.Result{Requeue: true}
	}
	return ctrl.Result{RequeueAfter: boundTimeoutExpirationDuration}
}

func (r *VirtualIPReconciler) validReferences(virtualIP *networkingv1alpha1.VirtualIP, nic *networkingv1alpha1.NetworkInterface) bool {
	targetRef := virtualIP.Spec.TargetRef
	if targetRef.UID != nic.UID {
		return false
	}

	nicVirtualIP := nic.Spec.VirtualIP
	if nicVirtualIP == nil {
		return false
	}

	switch {
	case nicVirtualIP.Ephemeral != nil:
		return virtualIP.Name == nic.Name
	case nicVirtualIP.VirtualIPRef != nil:
		return virtualIP.Name == nicVirtualIP.VirtualIPRef.Name
	default:
		return false
	}
}

func (r *VirtualIPReconciler) release(ctx context.Context, virtualIP *networkingv1alpha1.VirtualIP) error {
	baseVirtualIP := virtualIP.DeepCopy()
	virtualIP.Spec.TargetRef = nil
	return r.Patch(ctx, virtualIP, client.MergeFrom(baseVirtualIP))
}

func (r *VirtualIPReconciler) patchStatus(ctx context.Context, virtualIP *networkingv1alpha1.VirtualIP, phase networkingv1alpha1.VirtualIPPhase) error {
	now := metav1.Now()
	virtualIPBase := virtualIP.DeepCopy()

	if virtualIP.Status.Phase != phase {
		virtualIP.Status.LastPhaseTransitionTime = &now
	}
	virtualIP.Status.Phase = phase

	return r.Status().Patch(ctx, virtualIP, client.MergeFrom(virtualIPBase))
}

const (
	virtualIPSpecTargetRefNameField = ".spec.targetRef.name"
)

// SetupWithManager sets up the controller with the Manager.
func (r *VirtualIPReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("virtualip").WithName("setup")

	if err := mgr.GetFieldIndexer().IndexField(ctx, &networkingv1alpha1.VirtualIP{}, virtualIPSpecTargetRefNameField, func(object client.Object) []string {
		virtualIP := object.(*networkingv1alpha1.VirtualIP)
		targetRef := virtualIP.Spec.TargetRef
		if targetRef == nil {
			return []string{""}
		}
		return []string{targetRef.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1alpha1.VirtualIP{}).
		Watches(
			&source.Kind{Type: &networkingv1alpha1.NetworkInterface{}},
			r.enqueueByTargetNameReferencingNetworkInterface(ctx, log),
		).
		Watches(
			&source.Kind{Type: &networkingv1alpha1.NetworkInterface{}},
			r.enqueueByNameEqualNetworkInterfaceVirtualIPName(ctx, log),
		).
		Complete(r)
}

func (r *VirtualIPReconciler) enqueueByTargetNameReferencingNetworkInterface(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		nic := obj.(*networkingv1alpha1.NetworkInterface)

		virtualIPs := &networkingv1alpha1.VirtualIPList{}
		if err := r.List(ctx, virtualIPs, client.InNamespace(nic.Namespace), client.MatchingFields{
			virtualIPSpecTargetRefNameField: nic.Name,
		}); err != nil {
			log.Error(err, "Error listing virtual ips targeting network interface")
			return []ctrl.Request{}
		}

		res := make([]ctrl.Request, 0, len(virtualIPs.Items))
		for _, item := range virtualIPs.Items {
			res = append(res, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      item.GetName(),
					Namespace: item.GetNamespace(),
				},
			})
		}
		return res
	})
}

func (r *VirtualIPReconciler) enqueueByNameEqualNetworkInterfaceVirtualIPName(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		nic := obj.(*networkingv1alpha1.NetworkInterface)

		nicVirtualIP := nic.Spec.VirtualIP
		if nicVirtualIP == nil {
			return nil
		}

		nicVirtualIPName := networkingv1alpha1.NetworkInterfaceVirtualIPName(nic.Name, *nicVirtualIP)
		if nicVirtualIPName == "" {
			return nil
		}

		return []ctrl.Request{{NamespacedName: client.ObjectKey{Namespace: nic.Namespace, Name: nicVirtualIPName}}}
	})
}
