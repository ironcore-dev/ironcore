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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// VirtualIPClaimReconciler reconciles a VirtualIPClaimRef object
type VirtualIPClaimReconciler struct {
	client.Client
	APIReader client.Reader
	Scheme    *runtime.Scheme
}

//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualipclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualipclaims/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualipclaims/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualips,verbs=get;list

// Reconcile is part of the main reconciliation loop for VirtualIPClaimRef types
func (r *VirtualIPClaimReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	claim := &networkingv1alpha1.VirtualIPClaim{}
	if err := r.Get(ctx, req.NamespacedName, claim); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return r.reconcileExists(ctx, log, claim)
}

func (r *VirtualIPClaimReconciler) reconcileExists(ctx context.Context, log logr.Logger, claim *networkingv1alpha1.VirtualIPClaim) (ctrl.Result, error) {
	if !claim.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, claim)
	}
	return r.reconcile(ctx, log, claim)
}

func (r *VirtualIPClaimReconciler) delete(ctx context.Context, log logr.Logger, claim *networkingv1alpha1.VirtualIPClaim) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *VirtualIPClaimReconciler) reconcile(ctx context.Context, log logr.Logger, claim *networkingv1alpha1.VirtualIPClaim) (ctrl.Result, error) {
	log.V(1).Info("Reconciling virtual ip claim")
	if claim.Spec.VirtualIPRef == nil {
		log.V(1).Info("VirtualIP claim is not bound")
		if err := r.patchVirtualIPClaimStatus(ctx, claim, networkingv1alpha1.VirtualIPClaimPhasePending); err != nil {
			return ctrl.Result{}, fmt.Errorf("error setting virtual ip claim to pending: %w", err)
		}
		return ctrl.Result{}, nil
	}

	virtualIP := &networkingv1alpha1.VirtualIP{}
	virtualIPKey := client.ObjectKey{
		Namespace: claim.Namespace,
		Name:      claim.Spec.VirtualIPRef.Name,
	}
	log.V(1).Info("Getting virtual ip for virtual ip claim", "VirtualIPKey", virtualIPKey)
	// We have to use APIReader here as stale data might cause unbinding the already bound virtualIP.
	if err := r.APIReader.Get(ctx, virtualIPKey, virtualIP); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("error getting virtual ip %s for virtual ip claim: %w", virtualIPKey, err)
		}

		log.V(1).Info("VirtualIP claim is lost as the corresponding virtual ip cannot be found", "VirtualIPKey", virtualIPKey)
		if err := r.patchVirtualIPClaimStatus(ctx, claim, networkingv1alpha1.VirtualIPClaimPhaseLost); err != nil {
			return ctrl.Result{}, fmt.Errorf("error setting virtual ip claim to lost: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if virtualIPClaimRef := virtualIP.Spec.ClaimRef; virtualIPClaimRef != nil && virtualIPClaimRef.Name == claim.Name && virtualIPClaimRef.UID == claim.UID {
		log.V(1).Info("VirtualIP is bound to claim", "VirtualIPKey", virtualIPKey)
		if err := r.patchVirtualIPClaimStatus(ctx, claim, networkingv1alpha1.VirtualIPClaimPhaseBound); err != nil {
			return ctrl.Result{}, fmt.Errorf("error setting virtual ip claim to bound: %w", err)
		}
		return ctrl.Result{}, nil
	}

	log.V(1).Info("VirtualIP is not (yet) bound to claim", "VirtualIPKey", virtualIPKey, "ClaimRef", virtualIP.Spec.ClaimRef)
	if err := r.patchVirtualIPClaimStatus(ctx, claim, networkingv1alpha1.VirtualIPClaimPhasePending); err != nil {
		return ctrl.Result{}, fmt.Errorf("error setting virtual ip claim to pending: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *VirtualIPClaimReconciler) patchVirtualIPClaimStatus(ctx context.Context, virtualIPClaim *networkingv1alpha1.VirtualIPClaim, phase networkingv1alpha1.VirtualIPClaimPhase) error {
	base := virtualIPClaim.DeepCopy()
	virtualIPClaim.Status.Phase = phase
	return r.Status().Patch(ctx, virtualIPClaim, client.MergeFrom(base))
}

// SetupWithManager sets up the controller with the Manager.
func (r *VirtualIPClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("virtualipclaim").WithName("setup")

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1alpha1.VirtualIPClaim{}).
		Watches(&source.Kind{Type: &networkingv1alpha1.VirtualIP{}},
			handler.EnqueueRequestsFromMapFunc(func(object client.Object) []ctrl.Request {
				virtualIP := object.(*networkingv1alpha1.VirtualIP)

				claims := &networkingv1alpha1.VirtualIPClaimList{}
				if err := r.List(ctx, claims, &client.ListOptions{
					FieldSelector: fields.OneTermEqualSelector(virtualIPClaimSpecVirtualIPRefNameField, virtualIP.GetName()),
					Namespace:     virtualIP.GetNamespace(),
				}); err != nil {
					log.Error(err, "error listing virtual ip claims matching virtual ip", "VirtualIPKey", client.ObjectKeyFromObject(virtualIP))
					return []ctrl.Request{}
				}

				res := make([]ctrl.Request, 0, len(claims.Items))
				for _, item := range claims.Items {
					res = append(res, ctrl.Request{
						NamespacedName: client.ObjectKeyFromObject(&item),
					})
				}
				return res
			}),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}
