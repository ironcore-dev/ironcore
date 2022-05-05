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
	"time"

	"github.com/go-logr/logr"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	apiequality "github.com/onmetal/onmetal-api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// VirtualIPReconciler reconciles a VirtualIP object
type VirtualIPReconciler struct {
	client.Client
	APIReader client.Reader
	Scheme    *runtime.Scheme
	// BindTimeout is the maximum duration until a VirtualIP's Bound condition is considered to be timed out.
	BindTimeout time.Duration
}

//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualips,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualips/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualips/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualipclaims,verbs=get;list

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

func (r *VirtualIPReconciler) reconcile(ctx context.Context, log logr.Logger, virtualIP *networkingv1alpha1.VirtualIP) (ctrl.Result, error) {
	log.V(1).Info("Reconciling virtual ip")
	if virtualIP.Spec.ClaimRef == nil {
		log.V(1).Info("VirtualIP is not bound and not referencing any claim")
		if err := r.patchVirtualIPStatus(ctx, virtualIP, networkingv1alpha1.VirtualIPPhaseUnbound); err != nil {
			return ctrl.Result{}, err
		}

		log.V(1).Info("Successfully marked virtual ip as unbound")
		return ctrl.Result{}, nil
	}

	virtualIPClaim := &networkingv1alpha1.VirtualIPClaim{}
	virtualIPClaimKey := client.ObjectKey{
		Namespace: virtualIP.Namespace,
		Name:      virtualIP.Spec.ClaimRef.Name,
	}
	log = log.WithValues("VirtualIPClaimKey", virtualIPClaimKey)
	log.V(1).Info("VirtualIP references claim")
	// We have to use APIReader here as stale data might cause unbinding a virtual ip for a short duration.
	err := r.APIReader.Get(ctx, virtualIPClaimKey, virtualIPClaim)
	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, fmt.Errorf("error getting virtual ip claim %s: %w", virtualIPClaimKey, err)
	}

	virtualIPClaimExists := err == nil
	validReferences := virtualIPClaimExists && r.validReferences(virtualIP, virtualIPClaim)
	virtualIPPhase := virtualIP.Status.Phase
	virtualIPPhaseLastTransitionTime := virtualIP.Status.LastPhaseTransitionTime
	virtualIPIP := virtualIP.Status.IP

	bindOK := virtualIPIP.IsValid() && validReferences

	log = log.WithValues(
		"VirtualIPClaimExists", virtualIPClaimExists,
		"ValidReferences", validReferences,
		"VirtualIPIP", virtualIPIP,
		"VirtualIPPhase", virtualIPPhase,
		"VirtualIPPhaseLastTransitionTime", virtualIPPhaseLastTransitionTime,
	)
	switch {
	case bindOK:
		log.V(1).Info("Setting virtual ip to bound")
		if err := r.patchVirtualIPStatus(ctx, virtualIP, networkingv1alpha1.VirtualIPPhaseBound); err != nil {
			return ctrl.Result{}, fmt.Errorf("error binding virtualip: %w", err)
		}

		log.V(1).Info("Successfully set virtual ip to bound.")
		return ctrl.Result{}, nil
	case !bindOK && virtualIPPhase == networkingv1alpha1.VirtualIPPhasePending && r.phaseTransitionTimedOut(virtualIPPhaseLastTransitionTime):
		log.V(1).Info("Bind is not ok and timed out, releasing virtual ip")
		if err := r.releaseVirtualIP(ctx, virtualIP); err != nil {
			return ctrl.Result{}, fmt.Errorf("error releasing virtualip: %w", err)
		}

		log.V(1).Info("Successfully released virtual ip")
		return ctrl.Result{}, nil
	default:
		log.V(1).Info("Bind is not ok and not yet timed out, setting to pending")
		if err := r.patchVirtualIPStatus(ctx, virtualIP, networkingv1alpha1.VirtualIPPhasePending); err != nil {
			return ctrl.Result{}, fmt.Errorf("error setting virtualip to pending: %w", err)
		}

		log.V(1).Info("Successfully set virtual ip to pending")
		return r.requeueAfterBoundTimeout(virtualIP), nil
	}
}

func (r *VirtualIPReconciler) requeueAfterBoundTimeout(virtualIP *networkingv1alpha1.VirtualIP) ctrl.Result {
	boundTimeoutExpirationDuration := time.Until(virtualIP.Status.LastPhaseTransitionTime.Add(r.BindTimeout)).Round(time.Second)
	if boundTimeoutExpirationDuration <= 0 {
		return ctrl.Result{Requeue: true}
	}
	return ctrl.Result{RequeueAfter: boundTimeoutExpirationDuration}
}

func (r *VirtualIPReconciler) validReferences(virtualIP *networkingv1alpha1.VirtualIP, virtualIPClaim *networkingv1alpha1.VirtualIPClaim) bool {
	virtualIPRef := virtualIPClaim.Spec.VirtualIPRef
	if virtualIPRef.Name != virtualIP.Name {
		return false
	}

	claimRef := virtualIP.Spec.ClaimRef
	if claimRef == nil {
		return false
	}
	return claimRef.Name == virtualIPClaim.Name && claimRef.UID == virtualIPClaim.UID
}

func (r *VirtualIPReconciler) releaseVirtualIP(ctx context.Context, virtualIP *networkingv1alpha1.VirtualIP) error {
	baseVirtualIP := virtualIP.DeepCopy()
	virtualIP.Spec.ClaimRef = nil
	return r.Patch(ctx, virtualIP, client.MergeFrom(baseVirtualIP))
}

func (r *VirtualIPReconciler) patchVirtualIPStatus(ctx context.Context, virtualIP *networkingv1alpha1.VirtualIP, phase networkingv1alpha1.VirtualIPPhase) error {
	now := metav1.Now()
	virtualIPBase := virtualIP.DeepCopy()

	if virtualIP.Status.Phase != phase {
		virtualIP.Status.LastPhaseTransitionTime = &now
	}
	virtualIP.Status.Phase = phase

	return r.Status().Patch(ctx, virtualIP, client.MergeFrom(virtualIPBase))
}

const (
	virtualIPSpecVirtualIPClaimNameRefField = ".spec.claimRef.name"
)

// SetupWithManager sets up the controller with the Manager.
func (r *VirtualIPReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("virtualip").WithName("setup")

	if err := mgr.GetFieldIndexer().IndexField(ctx, &networkingv1alpha1.VirtualIP{}, virtualIPSpecVirtualIPClaimNameRefField, func(object client.Object) []string {
		virtualIP := object.(*networkingv1alpha1.VirtualIP)
		claimRef := virtualIP.Spec.ClaimRef
		if claimRef == nil {
			return nil
		}
		return []string{claimRef.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(
			&networkingv1alpha1.VirtualIP{},
			builder.WithPredicates(predicate.Funcs{
				UpdateFunc: func(event event.UpdateEvent) bool {
					oldVirtualIP, newVirtualIP := event.ObjectOld.(*networkingv1alpha1.VirtualIP), event.ObjectNew.(*networkingv1alpha1.VirtualIP)
					return !apiequality.Semantic.DeepEqual(oldVirtualIP.Spec, newVirtualIP.Spec) ||
						oldVirtualIP.Status.IP != newVirtualIP.Status.IP ||
						oldVirtualIP.Status.Phase != newVirtualIP.Status.Phase
				},
			}),
		).
		Watches(&source.Kind{Type: &networkingv1alpha1.VirtualIPClaim{}},
			handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
				virtualIPClaim := obj.(*networkingv1alpha1.VirtualIPClaim)

				virtualIPs := &networkingv1alpha1.VirtualIPList{}
				if err := r.List(ctx, virtualIPs, client.InNamespace(virtualIPClaim.Namespace), client.MatchingFields{
					virtualIPSpecVirtualIPClaimNameRefField: virtualIPClaim.Name,
				}); err != nil {
					log.Error(err, "Error listing claims using virtual ip")
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
			}),
			builder.WithPredicates(predicate.Funcs{
				UpdateFunc: func(event event.UpdateEvent) bool {
					oldVirtualIPClaim, newVirtualIPClaim := event.ObjectOld.(*networkingv1alpha1.VirtualIPClaim), event.ObjectNew.(*networkingv1alpha1.VirtualIPClaim)
					return !apiequality.Semantic.DeepEqual(oldVirtualIPClaim.Spec, newVirtualIPClaim.Spec) ||
						oldVirtualIPClaim.Status.Phase != newVirtualIPClaim.Status.Phase
				},
			}),
		).
		Complete(r)
}
