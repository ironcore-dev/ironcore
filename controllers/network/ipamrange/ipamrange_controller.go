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

package ipamrange

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/go-logr/logr"
	common "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	"github.com/onmetal/onmetal-api/apis/core"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"github.com/onmetal/onmetal-api/pkg/utils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	finalizerName = core.LabelDomain + "/ipamrange"
	fieldOwner    = client.FieldOwner(finalizerName)
)

// Reconciler reconciles a IPAMRange object
type Reconciler struct {
	client.Client
	cache *IPAMCache
}

//+kubebuilder:rbac:groups=network.onmetal.de,resources=ipamranges,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=network.onmetal.de,resources=ipamranges/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=network.onmetal.de,resources=ipamranges/finalizers,verbs=update

// Reconcile
// if parent -> handle request part and set range
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	ipamRange := &networkv1alpha1.IPAMRange{}
	if err := r.Get(ctx, req.NamespacedName, ipamRange); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !ipamRange.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, ipamRange)
	}
	return r.reconcileExists(ctx, log, ipamRange)
}

var relevantIPAMChanges = predicate.Funcs{
	UpdateFunc: func(event event.UpdateEvent) bool {
		oldIpamRange, newIpamRange := event.ObjectOld.(*networkv1alpha1.IPAMRange), event.ObjectNew.(*networkv1alpha1.IPAMRange)

		type RelevantStatus struct {
			CIDRs            []networkv1alpha1.CIDRAllocationStatus
			AllocationState  []string
			RoundRobinState  []string
			PendingRequest   *networkv1alpha1.IPAMPendingRequest
			PendingDeletions []networkv1alpha1.CIDRAllocationStatus
		}

		oldStatus := RelevantStatus{
			CIDRs:            oldIpamRange.Status.CIDRs,
			AllocationState:  oldIpamRange.Status.AllocationState,
			RoundRobinState:  oldIpamRange.Status.RoundRobinState,
			PendingRequest:   oldIpamRange.Status.PendingRequest,
			PendingDeletions: oldIpamRange.Status.PendingDeletions,
		}
		newStatus := RelevantStatus{
			CIDRs:            newIpamRange.Status.CIDRs,
			AllocationState:  newIpamRange.Status.AllocationState,
			RoundRobinState:  newIpamRange.Status.RoundRobinState,
			PendingRequest:   newIpamRange.Status.PendingRequest,
			PendingDeletions: newIpamRange.Status.PendingDeletions,
		}

		return !equality.Semantic.DeepEqual(oldIpamRange.Spec, newIpamRange.Spec) ||
			!equality.Semantic.DeepEqual(oldStatus, newStatus)
	},
}

const parentField = ".spec.parent.name"

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr manager.Manager) error {
	ctx := context.Background()
	r.cache = NewIPAMCache(r.Client)

	if err := mgr.GetFieldIndexer().IndexField(ctx, &networkv1alpha1.IPAMRange{}, parentField, func(object client.Object) []string {
		ipamRange := object.(*networkv1alpha1.IPAMRange)
		if ipamRange.Spec.Parent == nil {
			return nil
		}
		return []string{ipamRange.Spec.Parent.Name}
	}); err != nil {
		return fmt.Errorf("could not index field %s: %w", parentField, err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkv1alpha1.IPAMRange{}).
		Watches(
			// Parents trigger their children.
			&source.Kind{Type: &networkv1alpha1.IPAMRange{}},
			handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
				ipamRange := object.(*networkv1alpha1.IPAMRange)
				if ipamRange.Spec.Parent == nil {
					return nil
				}

				childrenList := &networkv1alpha1.IPAMRangeList{}
				if err := mgr.GetClient().List(ctx, childrenList, client.MatchingFields{parentField: ipamRange.Name}); err != nil {
					return nil
				}

				requests := make([]reconcile.Request, 0, len(childrenList.Items))
				for _, child := range childrenList.Items {
					requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&child)})
				}
				return requests
			}),
			builder.WithPredicates(relevantIPAMChanges),
		).
		Watches(
			// Children trigger their parents.
			&source.Kind{Type: &networkv1alpha1.IPAMRange{}},
			handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
				ipamRange := object.(*networkv1alpha1.IPAMRange)
				if ipamRange.Spec.Parent == nil {
					return nil
				}

				return []reconcile.Request{
					{
						NamespacedName: types.NamespacedName{
							Namespace: ipamRange.Namespace,
							Name:      ipamRange.Spec.Parent.Name,
						},
					},
				}
			}),
			builder.WithPredicates(relevantIPAMChanges),
		).
		Complete(r)
}

func (r *Reconciler) reconcileExists(ctx context.Context, log logr.Logger, obj *networkv1alpha1.IPAMRange) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(obj, finalizerName) {
		controllerutil.AddFinalizer(obj, finalizerName)
		if err := r.Update(ctx, obj); err != nil {
			return ctrl.Result{}, fmt.Errorf("error setting finalizer: %w", err)
		}
	}

	newObjectKey := utils.NewObjectKey(obj)
	current, err := r.cache.getRange(ctx, log, newObjectKey, obj)
	if err != nil {
		return utils.Requeue(err)
	}
	defer r.cache.release(log, newObjectKey)
	if current.error != "" {
		return r.invalid(ctx, log, current, current.error)
	}
	if obj.Spec.Parent != nil {
		// handle IPAMRange as request first
		result, err := r.reconcileChildRequest(ctx, log, current)
		if err != nil || !result.IsZero() {
			return result, err
		}
	} else {
		result, err := r.reconcileRootRequest(ctx, log, current)
		if err != nil || !result.IsZero() {
			return result, err
		}
	}
	if current.ipam != nil {
		return r.reconcileRange(ctx, log, current)
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) delete(ctx context.Context, log logr.Logger, obj *networkv1alpha1.IPAMRange) (ctrl.Result, error) {
	newObjectKey := utils.NewObjectKey(obj)
	current, err := r.cache.getRange(ctx, log, newObjectKey, obj)
	if err != nil {
		return utils.Requeue(err)
	}
	defer r.cache.release(log, newObjectKey)
	return r.deleteRequest(ctx, log, current)
}

// TODO: generalize state handling
func (r *Reconciler) invalid(ctx context.Context, log logr.Logger, ipr *IPAM, message string, args ...interface{}) (ctrl.Result, error) {
	return r.setStatus(ctx, log, ipr, common.StateInvalid, message, args...)
}

func (r *Reconciler) setStatus(ctx context.Context, log logr.Logger, ipr *IPAM, state string, msg string, args ...interface{}) (ctrl.Result, error) {
	if err := updateStatus(ctx, log, r, ipr, func(obj *networkv1alpha1.IPAMRange) {
		obj.Status.State = state
		obj.Status.Message = fmt.Sprintf(msg, args...)
	}, nil); err != nil {
		return utils.Requeue(err)
	}
	return utils.Succeeded()
}

func updateStatus(ctx context.Context, log logr.Logger, clt client.Client, ipr *IPAM, update func(ipamRange *networkv1alpha1.IPAMRange), cache func()) error {
	newObj := ipr.object.DeepCopy()
	update(newObj)
	if !reflect.DeepEqual(newObj, ipr.object) {
		log.Info("updating status", "state", newObj.Status.State, "message", newObj.Status.Message)
		if err := clt.Status().Patch(ctx, newObj, client.MergeFrom(ipr.object)); err != nil {
			return err
		}
		// update self in cache
		if ipr.ipam == nil {
			ipr.updateFrom(log, newObj)
		} else {
			// attention; cache updated need to be done externally
			ipr.object = newObj
			if cache != nil {
				cache()
			}
		}
	}
	return nil
}
