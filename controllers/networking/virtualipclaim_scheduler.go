/*
 * Copyright (c) 2022 by the OnMetal authors.
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
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type VirtualIPClaimScheduler struct {
	client.Client
	record.EventRecorder
}

//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualipclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualipclaims/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualipclaims/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualips,verbs=get;list;watch;update;patch

func (s *VirtualIPClaimScheduler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	virtualIPClaim := &networkingv1alpha1.VirtualIPClaim{}
	if err := s.Get(ctx, req.NamespacedName, virtualIPClaim); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return s.reconcileExists(ctx, log, virtualIPClaim)
}

func (s *VirtualIPClaimScheduler) reconcileExists(ctx context.Context, log logr.Logger, claim *networkingv1alpha1.VirtualIPClaim) (ctrl.Result, error) {
	if !claim.DeletionTimestamp.IsZero() {
		return s.delete(ctx, log, claim)
	}
	return s.reconcile(ctx, log, claim)
}

func (s *VirtualIPClaimScheduler) delete(ctx context.Context, log logr.Logger, claim *networkingv1alpha1.VirtualIPClaim) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (s *VirtualIPClaimScheduler) bind(ctx context.Context, virtualIPClaim *networkingv1alpha1.VirtualIPClaim, virtualIP *networkingv1alpha1.VirtualIP) error {
	baseVirtualIP := virtualIP.DeepCopy()
	virtualIP.Spec.ClaimRef = &commonv1alpha1.LocalUIDReference{
		Name: virtualIPClaim.Name,
		UID:  virtualIPClaim.UID,
	}
	if err := s.Patch(ctx, virtualIP, client.MergeFrom(baseVirtualIP)); err != nil {
		return fmt.Errorf("could not assign virtualIP claim to virtualIP %s: %w", client.ObjectKeyFromObject(virtualIP), err)
	}

	baseClaim := virtualIPClaim.DeepCopy()
	virtualIPClaim.Spec.VirtualIPRef = &corev1.LocalObjectReference{Name: virtualIP.Name}
	if err := s.Patch(ctx, virtualIPClaim, client.MergeFrom(baseClaim)); err != nil {
		return fmt.Errorf("could not assign virtualIP %s to virtualIP claim: %w", client.ObjectKeyFromObject(virtualIP), err)
	}

	return nil
}

func (s *VirtualIPClaimScheduler) reconcile(ctx context.Context, log logr.Logger, virtualIPClaim *networkingv1alpha1.VirtualIPClaim) (ctrl.Result, error) {
	log.V(1).Info("Reconciling virtual ip claim")
	if virtualIPRef := virtualIPClaim.Spec.VirtualIPRef; virtualIPRef != nil {
		virtualIP := &networkingv1alpha1.VirtualIP{}
		virtualIPKey := client.ObjectKey{Namespace: virtualIPClaim.Namespace, Name: virtualIPRef.Name}
		log = log.WithValues("VirtualIPKey", virtualIPKey)
		log.V(1).Info("Getting virtual ip specified by claim")
		if err := s.Get(ctx, virtualIPKey, virtualIP); err != nil {
			if !apierrors.IsNotFound(err) {
				return ctrl.Result{}, fmt.Errorf("error getting virtual ip %s specified by claim: %w", virtualIPKey, err)
			}

			log.V(1).Info("VirtualIP specified by claim does not exist")
			return ctrl.Result{}, nil
		}

		if virtualIP.Spec.ClaimRef != nil {
			log.V(1).Info("VirtualIP already specifies a claim ref", "ClaimRef", virtualIP.Spec.ClaimRef)
			return ctrl.Result{}, nil
		}

		log.V(1).Info("VirtualIP is not yet bound, trying to bind it")
		if err := s.bind(ctx, virtualIPClaim, virtualIP); err != nil {
			return ctrl.Result{}, fmt.Errorf("error binding: %w", err)
		}

		log.V(1).Info("Successfully bound virtualIP and claim")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Listing suitable virtual ips")
	sel, err := metav1.LabelSelectorAsSelector(virtualIPClaim.Spec.Selector)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("invalid label selector: %w", err)
	}

	virtualIPList := &networkingv1alpha1.VirtualIPList{}
	if err := s.List(ctx, virtualIPList,
		client.InNamespace(virtualIPClaim.Namespace),
		client.MatchingLabelsSelector{Selector: sel},
	); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to list matching virtual ips: %w", err)
	}

	log.V(1).Info("Searching for suitable virtual ips for claim")
	virtualIP := s.findVirtualIPForClaim(virtualIPList.Items, virtualIPClaim)
	if virtualIP == nil {
		s.Event(virtualIPClaim, corev1.EventTypeNormal, "FailedScheduling", "no matching virtual ip found for claim")
		log.V(1).Info("Could not find a matching virtual ip for claim")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("VirtualIPKey", client.ObjectKeyFromObject(virtualIP))
	log.V(1).Info("Found matching virtual ip, binding virtual ip and claim")
	if err := s.bind(ctx, virtualIPClaim, virtualIP); err != nil {
		return ctrl.Result{}, fmt.Errorf("error binding virtual ip %s to claim: %w", client.ObjectKeyFromObject(virtualIP), err)
	}

	log.V(1).Info("Successfully bound virtual ip to claim", "VirtualIP", client.ObjectKeyFromObject(virtualIP))
	return ctrl.Result{}, nil
}

func (s *VirtualIPClaimScheduler) findVirtualIPForClaim(virtualIPs []networkingv1alpha1.VirtualIP, claim *networkingv1alpha1.VirtualIPClaim) *networkingv1alpha1.VirtualIP {
	var matchingVirtualIP *networkingv1alpha1.VirtualIP
	for _, vip := range virtualIPs {
		if !vip.DeletionTimestamp.IsZero() {
			continue
		}

		if claimRef := vip.Spec.ClaimRef; claimRef != nil {
			if claimRef.Name != claim.Name {
				continue
			}
			if vip.Spec.ClaimRef.UID != claim.UID {
				continue
			}
			// If we hit a VirtualIP that matches exactly our claim we need to return immediately to avoid over-claiming
			// VirtualIPs in the cluster.
			vip := vip
			return &vip
		}

		if !s.virtualIPSatisfiesClaim(&vip, claim) {
			continue
		}
		vip := vip
		matchingVirtualIP = &vip
	}
	return matchingVirtualIP
}

func (s *VirtualIPClaimScheduler) virtualIPSatisfiesClaim(virtualIP *networkingv1alpha1.VirtualIP, claim *networkingv1alpha1.VirtualIPClaim) bool {
	return virtualIP.Status.IP.IsValid()
}

// SetupWithManager sets up the controller with the Manager.
func (s *VirtualIPClaimScheduler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("virtualipclaim-scheduler").WithName("setup")
	return ctrl.NewControllerManagedBy(mgr).
		Named("virtualipclaim-scheduler").
		For(&networkingv1alpha1.VirtualIPClaim{}, builder.WithPredicates(predicate.NewPredicateFuncs(func(object client.Object) bool {
			// Only reconcile claims which haven't been bound
			claim := object.(*networkingv1alpha1.VirtualIPClaim)
			return claim.Status.Phase == networkingv1alpha1.VirtualIPClaimPhasePending
		}))).
		Watches(&source.Kind{Type: &networkingv1alpha1.VirtualIP{}}, s.enqueueByVirtualIP(ctx, log)).
		Complete(s)
}

func (s *VirtualIPClaimScheduler) enqueueByVirtualIP(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(
		func(object client.Object) []ctrl.Request {
			virtualIP := object.(*networkingv1alpha1.VirtualIP)
			if claimRef := virtualIP.Spec.ClaimRef; claimRef != nil {
				claim := &networkingv1alpha1.VirtualIPClaim{}
				claimKey := client.ObjectKey{Namespace: virtualIP.Namespace, Name: claimRef.Name}
				if err := s.Get(ctx, claimKey, claim); err != nil {
					if !apierrors.IsNotFound(err) {
						log.Error(err, "Failed to get claim referenced by virtual ip", "VirtualIPClaimRef", claimKey)
						return nil
					}

					log.V(1).Info("Claim referenced by virtual ip does not exist", "VirtualIPClaimRef", claimKey)
					return nil
				}

				if claim.Spec.VirtualIPRef != nil {
					return nil
				}
				log.V(1).Info("Enqueueing claim that has already been accepted by its virtual ip", "VirtualIPClaimRef", claimKey)
				return []ctrl.Request{{NamespacedName: claimKey}}
			}

			virtualIPClaims := &networkingv1alpha1.VirtualIPClaimList{}
			if err := s.List(ctx, virtualIPClaims,
				client.InNamespace(virtualIP.Namespace),
				client.MatchingFields{
					virtualIPClaimSpecVirtualIPRefNameField: "",
				},
			); err != nil {
				log.Error(err, "Could not list empty virtual ip claims", "Namespace", virtualIP.Namespace)
				return nil
			}

			var requests []ctrl.Request
			for _, claim := range virtualIPClaims.Items {
				if s.virtualIPSatisfiesClaim(virtualIP, &claim) {
					requests = append(requests, ctrl.Request{
						NamespacedName: client.ObjectKeyFromObject(&claim),
					})
				}
			}
			return requests
		})
}
