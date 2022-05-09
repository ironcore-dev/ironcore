// Copyright 2022 OnMetal authors
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
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	onmetalapiclientutils "github.com/onmetal/onmetal-api/clientutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	aliasPrefixFieldOwner = client.FieldOwner(networkingv1alpha1.Resource("aliasprefixes").String())
)

type AliasPrefixReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *AliasPrefixReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	aliasPrefix := &networkingv1alpha1.AliasPrefix{}
	if err := r.Get(ctx, req.NamespacedName, aliasPrefix); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileExists(ctx, log, aliasPrefix)
}

func (r *AliasPrefixReconciler) reconcileExists(ctx context.Context, log logr.Logger, aliasPrefix *networkingv1alpha1.AliasPrefix) (ctrl.Result, error) {
	if !aliasPrefix.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, aliasPrefix)
	}
	return r.reconcile(ctx, log, aliasPrefix)
}

func (r *AliasPrefixReconciler) delete(ctx context.Context, log logr.Logger, aliasPrefix *networkingv1alpha1.AliasPrefix) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *AliasPrefixReconciler) applyPrefix(ctx context.Context, log logr.Logger, aliasPrefix *networkingv1alpha1.AliasPrefix) (*commonv1alpha1.IPPrefix, error) {
	prefixSrc := aliasPrefix.Spec.Prefix
	switch {
	case prefixSrc.Value != nil:
		log.V(1).Info("Static prefix")
		return prefixSrc.Value, nil
	case prefixSrc.Ephemeral != nil:
		log.V(1).Info("Ephemeral prefix")
		prefix := &ipamv1alpha1.Prefix{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: aliasPrefix.Namespace,
				Name:      aliasPrefix.Name,
			},
		}
		if err := onmetalapiclientutils.ControlledCreateOrGet(ctx, r.Client, aliasPrefix, prefix, func() error {
			template := prefixSrc.Ephemeral.PrefixTemplate
			prefix.Labels = template.Labels
			prefix.Annotations = template.Annotations
			prefix.Spec = template.Spec
			return nil
		}); err != nil {
			return nil, fmt.Errorf("error managing prefix: %w", err)
		}

		readiness := ipamv1alpha1.GetPrefixReadiness(prefix)
		if readiness == ipamv1alpha1.ReadinessSucceeded {
			log.V(1).Info("Ephemeral prefix is ready")
			return prefix.Spec.Prefix, nil
		}
		log.V(1).Info("Ephemeral prefix is not yet ready", "Readiness", readiness)
		return nil, nil
	default:
		return nil, fmt.Errorf("invalid prefix source %#v", prefixSrc)
	}
}

func (r *AliasPrefixReconciler) reconcile(ctx context.Context, log logr.Logger, aliasPrefix *networkingv1alpha1.AliasPrefix) (ctrl.Result, error) {
	log.V(1).Info("Reconcile")

	log.V(1).Info("Applying prefix")
	prefix, err := r.applyPrefix(ctx, log, aliasPrefix)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error applying prefix: %w", err)
	}
	log.V(1).Info("Successfully applied prefix", "Prefix", prefix)

	log.V(1).Info("Patching status")
	if err := r.patchStatus(ctx, aliasPrefix, prefix); err != nil {
		return ctrl.Result{}, fmt.Errorf("error patching status: %w", err)
	}
	log.V(1).Info("Successfully patched status")

	nicSelector := aliasPrefix.Spec.NetworkInterfaceSelector
	if nicSelector == nil {
		log.V(1).Info("Network interface selector is empty, assuming prefix is managed by external process")
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Network interface selector is present, managing routing")

	log.V(1).Info("Finding destinations")
	destinations, err := r.findDestinations(ctx, log, aliasPrefix)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error finding destinations: %w", err)
	}
	log.V(1).Info("Successfully found destinations", "Destinations", destinations)

	log.V(1).Info("Applying routing")
	if err := r.applyRouting(ctx, aliasPrefix, destinations); err != nil {
		return ctrl.Result{}, fmt.Errorf("error applying routing: %w", err)
	}
	log.V(1).Info("Successfully applied routing")
	return ctrl.Result{}, nil
}

func (r *AliasPrefixReconciler) patchStatus(ctx context.Context, aliasPrefix *networkingv1alpha1.AliasPrefix, prefix *commonv1alpha1.IPPrefix) error {
	base := aliasPrefix.DeepCopy()
	aliasPrefix.Status.Prefix = prefix
	if err := r.Status().Patch(ctx, aliasPrefix, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error patching status: %w", err)
	}
	return nil
}

func (r *AliasPrefixReconciler) findDestinations(ctx context.Context, log logr.Logger, aliasPrefix *networkingv1alpha1.AliasPrefix) ([]commonv1alpha1.LocalUIDReference, error) {
	sel, err := metav1.LabelSelectorAsSelector(aliasPrefix.Spec.NetworkInterfaceSelector)
	if err != nil {
		return nil, err
	}

	nicList := &networkingv1alpha1.NetworkInterfaceList{}
	if err := r.List(ctx, nicList,
		client.InNamespace(aliasPrefix.Namespace),
		client.MatchingLabelsSelector{Selector: sel},
		client.MatchingFields{networkInterfaceSpecNetworkRefNameField: aliasPrefix.Spec.NetworkRef.Name},
	); err != nil {
		return nil, fmt.Errorf("error listing network interfaces: %w", err)
	}

	destinations := make([]commonv1alpha1.LocalUIDReference, 0, len(nicList.Items))
	for _, nic := range nicList.Items {
		destinations = append(destinations, commonv1alpha1.LocalUIDReference{Name: nic.Name, UID: nic.UID})
	}
	return destinations, nil
}

func (r *AliasPrefixReconciler) applyRouting(ctx context.Context, aliasPrefix *networkingv1alpha1.AliasPrefix, destinations []commonv1alpha1.LocalUIDReference) error {
	aliasPrefixRouting := &networkingv1alpha1.AliasPrefixRouting{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AliasPrefixRouting",
			APIVersion: networkingv1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: aliasPrefix.Namespace,
			Name:      aliasPrefix.Name,
		},
		Destinations: destinations,
	}
	if err := ctrl.SetControllerReference(aliasPrefix, aliasPrefixRouting, r.Scheme); err != nil {
		return fmt.Errorf("error setting controller reference: %w", err)
	}
	if err := r.Patch(ctx, aliasPrefixRouting, client.Apply, aliasPrefixFieldOwner, client.ForceOwnership); err != nil {
		return fmt.Errorf("error applying alias prefix routing: %w", err)
	}
	return nil
}

const (
	networkInterfaceSpecNetworkRefNameField = ".spec.networkRef.name"
	aliasPrefixSpecNetworkRefNameField      = ".spec.networkRef.name"
)

func (r *AliasPrefixReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()
	log := ctrl.Log.WithName("aliasprefix").WithName("setup")

	if err := mgr.GetFieldIndexer().IndexField(ctx, &networkingv1alpha1.NetworkInterface{}, networkInterfaceSpecNetworkRefNameField, func(obj client.Object) []string {
		nic := obj.(*networkingv1alpha1.NetworkInterface)
		return []string{nic.Spec.NetworkRef.Name}
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &networkingv1alpha1.AliasPrefix{}, aliasPrefixSpecNetworkRefNameField, func(obj client.Object) []string {
		aliasPrefix := obj.(*networkingv1alpha1.AliasPrefix)
		return []string{aliasPrefix.Spec.NetworkRef.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1alpha1.AliasPrefix{}).
		Owns(&networkingv1alpha1.AliasPrefixRouting{}).
		Watches(
			&source.Kind{Type: &networkingv1alpha1.NetworkInterface{}},
			r.enqueueByAliasPrefixMatchingNetworkInterface(log, ctx),
		).
		Complete(r)
}

func (r *AliasPrefixReconciler) enqueueByAliasPrefixMatchingNetworkInterface(log logr.Logger, ctx context.Context) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []ctrl.Request {
		nic := obj.(*networkingv1alpha1.NetworkInterface)
		log = log.WithValues("NetworkInterfaceKey", client.ObjectKeyFromObject(nic))

		aliasPrefixList := &networkingv1alpha1.AliasPrefixList{}
		if err := r.List(ctx, aliasPrefixList,
			client.InNamespace(nic.Namespace),
			client.MatchingFields{aliasPrefixSpecNetworkRefNameField: nic.Spec.NetworkRef.Name},
		); err != nil {
			log.Error(err, "Error listing alias prefixes for network")
			return nil
		}

		var res []ctrl.Request
		for _, aliasPrefix := range aliasPrefixList.Items {
			aliasPrefixKey := client.ObjectKeyFromObject(&aliasPrefix)
			log := log.WithValues("AliasPrefixKey", aliasPrefixKey)
			nicSelector := aliasPrefix.Spec.NetworkInterfaceSelector
			if nicSelector == nil {
				return nil
			}

			sel, err := metav1.LabelSelectorAsSelector(nicSelector)
			if err != nil {
				log.Error(err, "Invalid network interface selector")
				continue
			}

			if sel.Matches(labels.Set(nic.Labels)) {
				res = append(res, ctrl.Request{NamespacedName: aliasPrefixKey})
			}
		}
		return res
	})
}
