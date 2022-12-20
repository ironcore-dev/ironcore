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

package aliasprefixes

import (
	"context"
	"fmt"

	"github.com/onmetal/controller-utils/set"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/common/cleaner"
	commonsync "github.com/onmetal/onmetal-api/broker/common/sync"
	"github.com/onmetal/onmetal-api/broker/common/utils"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/machinebroker/apiutils"
	"github.com/onmetal/onmetal-api/broker/machinebroker/cluster"
	utilslices "github.com/onmetal/onmetal-api/utils/slices"
	"golang.org/x/exp/slices"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type aliasPrefixKey struct {
	networkHandle string
	prefix        commonv1alpha1.IPPrefix
}

func (k aliasPrefixKey) String() string {
	return fmt.Sprintf("%s/%s", k.networkHandle, k.prefix.String())
}

type AliasPrefixes struct {
	mu commonsync.MutexMap[aliasPrefixKey]

	cluster cluster.Cluster
}

func New(cluster cluster.Cluster) *AliasPrefixes {
	return &AliasPrefixes{
		mu:      *commonsync.NewMutexMap[aliasPrefixKey](),
		cluster: cluster,
	}
}

func (m *AliasPrefixes) filterAliasPrefixes(aliasPrefixes []networkingv1alpha1.AliasPrefix) []networkingv1alpha1.AliasPrefix {
	var filtered []networkingv1alpha1.AliasPrefix
	for _, aliasPrefix := range aliasPrefixes {
		if aliasPrefix.DeletionTimestamp.IsZero() {
			continue
		}

		filtered = append(filtered, aliasPrefix)
	}
	return filtered
}

func (m *AliasPrefixes) getAliasPrefixByKey(ctx context.Context, key aliasPrefixKey) (*networkingv1alpha1.AliasPrefix, *networkingv1alpha1.AliasPrefixRouting, bool, error) {
	aliasPrefixList := &networkingv1alpha1.AliasPrefixList{}
	if err := m.cluster.Client().List(ctx, aliasPrefixList,
		client.InNamespace(m.cluster.Namespace()),
		client.MatchingLabels{
			machinebrokerv1alpha1.ManagerLabel:       machinebrokerv1alpha1.MachineBrokerManager,
			machinebrokerv1alpha1.CreatedLabel:       "true",
			machinebrokerv1alpha1.NetworkHandleLabel: key.networkHandle,
			machinebrokerv1alpha1.PrefixLabel:        apiutils.EscapePrefix(key.prefix),
		},
	); err != nil {
		return nil, nil, false, fmt.Errorf("error listing alias prefixes by key: %w", err)
	}

	aliasPrefixes := m.filterAliasPrefixes(aliasPrefixList.Items)

	switch len(aliasPrefixes) {
	case 0:
		return nil, nil, false, nil
	case 1:
		aliasPrefix := aliasPrefixes[0]
		aliasPrefixRouting := &networkingv1alpha1.AliasPrefixRouting{}
		if err := m.cluster.Client().Get(ctx, client.ObjectKeyFromObject(&aliasPrefix), aliasPrefixRouting); err != nil {
			return nil, nil, false, fmt.Errorf("error getting alias prefix routing: %w", err)
		}
		return &aliasPrefix, aliasPrefixRouting, true, nil
	default:
		return nil, nil, false, fmt.Errorf("multiple alias prefixes found for key %s", key)
	}
}

func (m *AliasPrefixes) createAliasPrefix(
	ctx context.Context,
	key aliasPrefixKey,
	network *networkingv1alpha1.Network,
) (resPrefix *networkingv1alpha1.AliasPrefix, resPrefixRouting *networkingv1alpha1.AliasPrefixRouting, retErr error) {
	c := cleaner.New()
	defer cleaner.CleanupOnError(ctx, c, &retErr)

	aliasPrefix := &networkingv1alpha1.AliasPrefix{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: m.cluster.Namespace(),
			Name:      m.cluster.IDGen().Generate(),
		},
		Spec: networkingv1alpha1.AliasPrefixSpec{
			NetworkRef: corev1.LocalObjectReference{Name: network.Name},
			Prefix: networkingv1alpha1.PrefixSource{
				Value: &key.prefix,
			},
		},
	}
	apiutils.SetManagerLabel(aliasPrefix, machinebrokerv1alpha1.MachineBrokerManager)
	apiutils.SetNetworkHandle(aliasPrefix, key.networkHandle)
	apiutils.SetPrefixLabel(aliasPrefix, key.prefix)

	if err := m.cluster.Client().Create(ctx, aliasPrefix); err != nil {
		return nil, nil, fmt.Errorf("error creating alias prefix: %w", err)
	}
	c.Add(cleaner.DeleteObjectIfExistsFunc(m.cluster.Client(), aliasPrefix))

	aliasPrefixRouting := &networkingv1alpha1.AliasPrefixRouting{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: m.cluster.Namespace(),
			Name:      aliasPrefix.Name,
		},
		NetworkRef: commonv1alpha1.LocalUIDReference{
			Name: network.Name,
			UID:  network.UID,
		},
	}
	apiutils.SetManagerLabel(aliasPrefixRouting, machinebrokerv1alpha1.MachineBrokerManager)
	apiutils.SetNetworkHandle(aliasPrefixRouting, key.networkHandle)
	apiutils.SetPrefixLabel(aliasPrefixRouting, key.prefix)
	if err := ctrl.SetControllerReference(aliasPrefix, aliasPrefixRouting, m.cluster.Scheme()); err != nil {
		return nil, nil, fmt.Errorf("error setting alias prefix routing to be controlled by alias prefix: %w", err)
	}

	if err := m.cluster.Client().Create(ctx, aliasPrefixRouting); err != nil {
		return nil, nil, fmt.Errorf("error creating alias prefix routing: %w", err)
	}

	return aliasPrefix, aliasPrefixRouting, nil
}

func (m *AliasPrefixes) removeAliasPrefixRoutingDestination(
	ctx context.Context,
	aliasPrefixRouting *networkingv1alpha1.AliasPrefixRouting,
	obj client.Object,
) error {
	idx := slices.IndexFunc(aliasPrefixRouting.Destinations,
		func(ref commonv1alpha1.LocalUIDReference) bool { return ref.UID == obj.GetUID() },
	)
	if idx == -1 {
		return nil
	}

	base := aliasPrefixRouting.DeepCopy()
	aliasPrefixRouting.Destinations = slices.Delete(aliasPrefixRouting.Destinations, idx, idx+1)
	if err := m.cluster.Client().Patch(ctx, aliasPrefixRouting, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error removing alias prefix routing destination: %w", err)
	}
	return nil
}

func (m *AliasPrefixes) addAliasPrefixRoutingDestination(
	ctx context.Context,
	aliasPrefixRouting *networkingv1alpha1.AliasPrefixRouting,
	obj client.Object,
) error {
	idx := slices.IndexFunc(aliasPrefixRouting.Destinations,
		func(ref commonv1alpha1.LocalUIDReference) bool { return ref.UID == obj.GetUID() },
	)
	if idx >= 0 {
		return nil
	}

	base := aliasPrefixRouting.DeepCopy()
	aliasPrefixRouting.Destinations = append(aliasPrefixRouting.Destinations, commonv1alpha1.LocalUIDReference{
		Name: obj.GetName(),
		UID:  obj.GetUID(),
	})
	if err := m.cluster.Client().Patch(ctx, aliasPrefixRouting, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error adding alias prefix routing destination: %w", err)
	}
	return nil
}

func (m *AliasPrefixes) Create(
	ctx context.Context,
	network *networkingv1alpha1.Network,
	prefix commonv1alpha1.IPPrefix,
	networkInterface *networkingv1alpha1.NetworkInterface,
) error {
	key := aliasPrefixKey{
		networkHandle: network.Spec.Handle,
		prefix:        prefix,
	}
	m.mu.Lock(key)
	defer m.mu.Unlock(key)

	c := cleaner.New()

	aliasPrefix, aliasPrefixRouting, ok, err := m.getAliasPrefixByKey(ctx, key)
	if err != nil {
		return fmt.Errorf("error getting alias prefix by key %s: %w", key, err)
	}
	if !ok {
		newAliasPrefix, newAliasPrefixRouting, err := m.createAliasPrefix(ctx, key, network)
		if err != nil {
			return fmt.Errorf("error creating alias prefix: %w", err)
		}

		c.Add(cleaner.DeleteObjectIfExistsFunc(m.cluster.Client(), newAliasPrefix))
		aliasPrefix = newAliasPrefix
		aliasPrefixRouting = newAliasPrefixRouting
	}

	if err := m.addAliasPrefixRoutingDestination(ctx, aliasPrefixRouting, networkInterface); err != nil {
		return err
	}
	c.Add(func(ctx context.Context) error {
		return m.removeAliasPrefixRoutingDestination(ctx, aliasPrefixRouting, networkInterface)
	})

	if err := apiutils.PatchCreatedWithDependent(ctx, m.cluster.Client(), aliasPrefix, networkInterface.GetName()); err != nil {
		return fmt.Errorf("error patching created with dependent: %w", err)
	}
	return nil
}

func (m *AliasPrefixes) Delete(ctx context.Context, networkHandle string, prefix commonv1alpha1.IPPrefix, obj client.Object) error {
	key := aliasPrefixKey{
		networkHandle: networkHandle,
		prefix:        prefix,
	}
	m.mu.Lock(key)
	defer m.mu.Unlock(key)

	aliasPrefix, aliasPrefixRouting, ok, err := m.getAliasPrefixByKey(ctx, key)
	if err != nil {
		return fmt.Errorf("error getting alias prefix by key: %w", err)
	}
	if !ok {
		return nil
	}

	var errs []error
	if err := m.removeAliasPrefixRoutingDestination(ctx, aliasPrefixRouting, obj); err != nil {
		errs = append(errs, fmt.Errorf("error removing alias prefix routing destination: %w", err))
	}
	if err := apiutils.DeleteAndGarbageCollect(ctx, m.cluster.Client(), aliasPrefix, obj.GetName()); err != nil {
		errs = append(errs, fmt.Errorf("error deleting / garbage collecting: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("error(s) deleting alias prefix: %v", errs)
	}
	return nil
}

type AliasPrefix struct {
	NetworkHandle string
	Prefix        commonv1alpha1.IPPrefix
	Destinations  set.Set[string]
}

func (m *AliasPrefixes) listAliasPrefixes(ctx context.Context, dependent string) ([]networkingv1alpha1.AliasPrefix, error) {
	aliasPrefixList := &networkingv1alpha1.AliasPrefixList{}
	if err := m.cluster.Client().List(ctx, aliasPrefixList,
		client.InNamespace(m.cluster.Namespace()),
		client.MatchingLabels{
			machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.MachineBrokerManager,
			machinebrokerv1alpha1.CreatedLabel: "true",
		},
	); err != nil {
		return nil, fmt.Errorf("error listing alias prefixes: %w", err)
	}

	if dependent == "" {
		return aliasPrefixList.Items, nil
	}

	filtered, err := apiutils.FilterObjectListByDependent(aliasPrefixList.Items, dependent)
	if err != nil {
		return nil, fmt.Errorf("error filtering by dependent: %w", err)
	}
	return filtered, nil
}

func (m *AliasPrefixes) listAliasPrefixRoutings(ctx context.Context) ([]networkingv1alpha1.AliasPrefixRouting, error) {
	aliasPrefixRoutingList := &networkingv1alpha1.AliasPrefixRoutingList{}
	if err := m.cluster.Client().List(ctx, aliasPrefixRoutingList,
		client.InNamespace(m.cluster.Namespace()),
		client.MatchingLabels{
			machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.MachineBrokerManager,
		},
	); err != nil {
		return nil, fmt.Errorf("error listing alias prefix routings: %w", err)
	}

	return aliasPrefixRoutingList.Items, nil
}

func (m *AliasPrefixes) List(ctx context.Context) ([]AliasPrefix, error) {
	aliasPrefixes, err := m.listAliasPrefixes(ctx, "")
	if err != nil {
		return nil, err
	}

	aliasPrefixRoutings, err := m.listAliasPrefixRoutings(ctx)
	if err != nil {
		return nil, err
	}

	return m.joinAliasPrefixesAndRoutings(aliasPrefixes, aliasPrefixRoutings), nil
}

func (m *AliasPrefixes) joinAliasPrefixesAndRoutings(
	aliasPrefixes []networkingv1alpha1.AliasPrefix,
	aliasPrefixRoutings []networkingv1alpha1.AliasPrefixRouting,
) []AliasPrefix {
	aliasPrefixRoutingByName := utils.ObjectSliceToMapByName(aliasPrefixRoutings)

	var res []AliasPrefix
	for i := range aliasPrefixes {
		aliasPrefix := &aliasPrefixes[i]

		prefixSrc := aliasPrefix.Spec.Prefix
		if prefixSrc.Value == nil {
			continue
		}

		aliasPrefixRouting, ok := aliasPrefixRoutingByName[aliasPrefix.Name]
		if !ok {
			continue
		}

		networkHandle, ok := aliasPrefix.Labels[machinebrokerv1alpha1.NetworkHandleLabel]
		if !ok {
			continue
		}

		destinations := utilslices.ToSetFunc(
			aliasPrefixRouting.Destinations,
			func(dest commonv1alpha1.LocalUIDReference) string {
				return dest.Name
			},
		)

		res = append(res, AliasPrefix{
			NetworkHandle: networkHandle,
			Prefix:        *prefixSrc.Value,
			Destinations:  destinations,
		})
	}
	return res
}

func (m *AliasPrefixes) ListByDependent(ctx context.Context, dependent string) ([]AliasPrefix, error) {
	aliasPrefixes, err := m.listAliasPrefixes(ctx, dependent)
	if err != nil {
		return nil, err
	}

	aliasPrefixRoutings, err := m.listAliasPrefixRoutings(ctx)
	if err != nil {
		return nil, err
	}

	return m.joinAliasPrefixesAndRoutings(aliasPrefixes, aliasPrefixRoutings), nil
}
