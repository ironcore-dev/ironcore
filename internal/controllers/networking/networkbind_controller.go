// Copyright 2023 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package networking

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	networkingclient "github.com/onmetal/onmetal-api/internal/client/networking"
	clientutils "github.com/onmetal/onmetal-api/utils/client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type peeringStatusData struct {
	phase  networkingv1alpha1.NetworkPeeringPhase
	handle string
}

type NetworkBindReconciler struct {
	record.EventRecorder
	client.Client
}

//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networks,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networks/status,verbs=get;update;patch

func (r *NetworkBindReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	network := &networkingv1alpha1.Network{}
	if err := r.Get(ctx, req.NamespacedName, network); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, network)
}

func (r *NetworkBindReconciler) reconcileExists(ctx context.Context, log logr.Logger, network *networkingv1alpha1.Network) (ctrl.Result, error) {
	if !network.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	return r.reconcile(ctx, log, network)
}

func (r *NetworkBindReconciler) reconcile(ctx context.Context, log logr.Logger, network *networkingv1alpha1.Network) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	peeringUIDByName := make(map[string]types.UID)
	peeringStatusDataByName := make(map[string]peeringStatusData)

	for _, peering := range network.Spec.Peerings {
		uid, data, err := r.reconcilePeering(ctx, log, network, peering)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("[network peering %s] %w", peering.Name, err)
		}

		switch {
		case uid != "":
			peeringUIDByName[peering.Name] = uid
		case data != nil:
			peeringStatusDataByName[peering.Name] = *data
		default:
		}
	}

	if len(peeringUIDByName) > 0 {
		log.V(1).Info("UID updates require spec update", "UIDUpdates", peeringUIDByName)
		if err := r.updateSpec(ctx, log, network, peeringUIDByName); err != nil {
			return ctrl.Result{}, fmt.Errorf("error updating spec: %w", err)
		}
		return ctrl.Result{}, nil
	}

	result, err := r.updateStatus(ctx, log, network, peeringStatusDataByName)
	if err != nil {
		return result, err
	}

	log.V(1).Info("Reconciled")
	return ctrl.Result{}, nil
}

func (r *NetworkBindReconciler) updateStatus(ctx context.Context, log logr.Logger, network *networkingv1alpha1.Network, peeringStatusDataByName map[string]peeringStatusData) (ctrl.Result, error) {
	base := network.DeepCopy()
	handledPeeringStatusDataNames := sets.New[string]()
	newStatusPeerings := make([]networkingv1alpha1.NetworkPeeringStatus, 0, len(peeringStatusDataByName))
	now := metav1.Now()
	for i := range network.Status.Peerings {
		peeringStatus := &network.Status.Peerings[i]
		data, ok := peeringStatusDataByName[peeringStatus.Name]
		if !ok {
			continue
		}

		lastPhaseTransitionTime := peeringStatus.LastPhaseTransitionTime
		if data.phase != peeringStatus.Phase {
			lastPhaseTransitionTime = &now
		}

		newStatusPeerings = append(newStatusPeerings, networkingv1alpha1.NetworkPeeringStatus{
			Name:                    peeringStatus.Name,
			NetworkHandle:           data.handle,
			Phase:                   data.phase,
			LastPhaseTransitionTime: lastPhaseTransitionTime,
		})
		handledPeeringStatusDataNames.Insert(peeringStatus.Name)
	}
	for name, data := range peeringStatusDataByName {
		if handledPeeringStatusDataNames.Has(name) {
			continue
		}
		newStatusPeerings = append(newStatusPeerings, networkingv1alpha1.NetworkPeeringStatus{
			Name:                    name,
			NetworkHandle:           data.handle,
			Phase:                   data.phase,
			LastPhaseTransitionTime: &now,
		})
	}
	network.Status.Peerings = newStatusPeerings

	log.V(1).Info("Updating network status")
	if err := r.Status().Patch(ctx, network, client.StrategicMergeFrom(base)); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating network status: %w", err)
	}
	log.V(1).Info("Updated network status")

	return ctrl.Result{}, nil
}

func (r *NetworkBindReconciler) updateSpec(
	ctx context.Context,
	log logr.Logger,
	network *networkingv1alpha1.Network,
	peeringUIDByName map[string]types.UID,
) error {
	base := network.DeepCopy()
	for i := range network.Spec.Peerings {
		peering := &network.Spec.Peerings[i]
		uid, ok := peeringUIDByName[peering.Name]
		if !ok {
			continue
		}

		peering.NetworkRef.UID = uid
	}
	log.V(1).Info("Updating spec")
	if err := r.Patch(ctx, network, client.StrategicMergeFrom(base)); err != nil {
		return fmt.Errorf("error patching spec: %w", err)
	}

	log.V(1).Info("Updated spec")
	return nil
}

func (r *NetworkBindReconciler) reconcilePeering(
	ctx context.Context,
	log logr.Logger,
	network *networkingv1alpha1.Network,
	peering networkingv1alpha1.NetworkPeering,
) (types.UID, *peeringStatusData, error) {
	if network.Status.State != networkingv1alpha1.NetworkStateAvailable {
		return "", &peeringStatusData{phase: networkingv1alpha1.NetworkPeeringPhasePending}, nil
	}

	networkKey := client.ObjectKeyFromObject(network)

	targetNetwork := &networkingv1alpha1.Network{}
	targetNetworkRef := peering.NetworkRef
	targetNetworkNamespace := targetNetworkRef.Namespace
	if targetNetworkNamespace == "" {
		targetNetworkNamespace = network.Namespace
	}
	targetNetworkKey := client.ObjectKey{Namespace: targetNetworkNamespace, Name: targetNetworkRef.Name}
	log = log.WithValues("TargetNetworkKey", targetNetworkKey)

	log.V(2).Info("Getting target network")
	if err := r.Get(ctx, targetNetworkKey, targetNetwork); err != nil {
		if !apierrors.IsNotFound(err) {
			return "", nil, fmt.Errorf("error getting network %s: %w", targetNetworkKey, err)
		}

		log.V(1).Info("Target network not found")
		return "", &peeringStatusData{phase: networkingv1alpha1.NetworkPeeringPhasePending}, nil
	}

	if targetNetworkRef.UID != "" && targetNetworkRef.UID != targetNetwork.UID {
		log.V(1).Info("Target network UID mismatch", "ExpectedUID", targetNetworkRef.UID, "ActualUID", targetNetwork.UID)
		return "", &peeringStatusData{phase: networkingv1alpha1.NetworkPeeringPhasePending}, nil
	}

	for _, targetPeering := range targetNetwork.Spec.Peerings {
		targetPeeringNetworkRef := targetPeering.NetworkRef
		targetPeeringNetworkNamespace := targetPeeringNetworkRef.Namespace
		if targetPeeringNetworkNamespace == "" {
			targetPeeringNetworkNamespace = targetNetwork.Namespace
		}
		targetPeeringNetworkKey := client.ObjectKey{Namespace: targetPeeringNetworkNamespace, Name: targetPeeringNetworkRef.Name}

		if targetPeeringNetworkKey != networkKey {
			continue
		}

		if targetPeeringNetworkRef.UID != "" && targetPeeringNetworkRef.UID != network.UID {
			log.V(1).Info("Target peering network ref UID mismatch", "ExpectedUID", targetPeeringNetworkRef.UID, "ActualUID", network.UID)
			return "", &peeringStatusData{phase: networkingv1alpha1.NetworkPeeringPhasePending}, nil
		}

		if targetNetworkRef.UID == "" {
			log.V(1).Info("Target network ref UID has to be assigned", "TargetNetworkUID", targetNetwork.UID)
			return targetNetwork.UID, nil, nil
		}

		if targetPeeringNetworkRef.UID == "" {
			log.V(1).Info("Target peering network ref UID has to be assigned")
			return "", &peeringStatusData{phase: networkingv1alpha1.NetworkPeeringPhasePending}, nil
		}

		if targetNetwork.Status.State != networkingv1alpha1.NetworkStateAvailable {
			log.V(1).Info("Target network is not available yet")
			return "", &peeringStatusData{phase: networkingv1alpha1.NetworkPeeringPhasePending}, nil
		}

		log.V(1).Info("Target network peering matches")
		handle := targetNetwork.Spec.Handle
		return "", &peeringStatusData{phase: networkingv1alpha1.NetworkPeeringPhaseBound, handle: handle}, nil
	}

	log.V(1).Info("No matching target peering found")
	return "", &peeringStatusData{phase: networkingv1alpha1.NetworkPeeringPhasePending}, nil
}

func (r *NetworkBindReconciler) enqueuePeeringReferencedNetworks() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		network := obj.(*networkingv1alpha1.Network)
		reqs := sets.New[ctrl.Request]()
		for _, peering := range network.Spec.Peerings {
			ref := peering.NetworkRef
			refNamespace := ref.Namespace
			if refNamespace == "" {
				refNamespace = network.Namespace
			}

			refKey := client.ObjectKey{Namespace: refNamespace, Name: ref.Name}
			reqs.Insert(ctrl.Request{NamespacedName: refKey})
		}
		return reqs.UnsortedList()
	})
}

func (r *NetworkBindReconciler) enqueuePeeringUsingNetworks(ctx context.Context, log logr.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		network := obj.(*networkingv1alpha1.Network)

		usingNetworkList := &networkingv1alpha1.NetworkList{}
		if err := r.List(ctx, usingNetworkList,
			client.MatchingFields{networkingclient.NetworkPeeringKeysField: networkingclient.NetworkPeeringKey(network)},
		); err != nil {
			log.Error(err, "Error listing using networks")
			return nil
		}

		return clientutils.ReconcileRequestsFromObjectSlice(usingNetworkList.Items)
	})
}

func (r *NetworkBindReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("networkbind").WithName("setup")
	ctx = ctrl.LoggerInto(ctx, log)

	return ctrl.NewControllerManagedBy(mgr).
		Named("networkbind").
		For(&networkingv1alpha1.Network{}).
		Watches(
			&source.Kind{Type: &networkingv1alpha1.Network{}},
			r.enqueuePeeringReferencedNetworks(),
		).
		Watches(
			&source.Kind{Type: &networkingv1alpha1.Network{}},
			r.enqueuePeeringUsingNetworks(ctx, log),
		).
		Complete(r)
}
