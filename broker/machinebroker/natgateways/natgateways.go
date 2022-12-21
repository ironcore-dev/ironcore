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

package natgateways

import (
	"context"
	"fmt"

	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/apiutils/annotations"
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

const (
	natGatewayIPName = "primary"
)

type NATGateways struct {
	mu commonsync.MutexMap[natGatewayKey]

	cluster cluster.Cluster
}

type natGatewayKey struct {
	networkHandle string
	ip            commonv1alpha1.IP
}

func (k natGatewayKey) String() string {
	return fmt.Sprintf("%s-%s", k.networkHandle, k.ip)
}

func New(cluster cluster.Cluster) *NATGateways {
	return &NATGateways{
		mu:      *commonsync.NewMutexMap[natGatewayKey](),
		cluster: cluster,
	}
}

func (m *NATGateways) filterNATGateways(natGateways []networkingv1alpha1.NATGateway) []networkingv1alpha1.NATGateway {
	var filtered []networkingv1alpha1.NATGateway
	for _, natGateway := range natGateways {
		if !natGateway.DeletionTimestamp.IsZero() {
			continue
		}

		filtered = append(filtered, natGateway)
	}
	return filtered
}

func (m *NATGateways) getNATGatewayByKey(ctx context.Context, key natGatewayKey) (*networkingv1alpha1.NATGateway, *networkingv1alpha1.NATGatewayRouting, bool, error) {
	natGatewayList := &networkingv1alpha1.NATGatewayList{}
	if err := m.cluster.Client().List(ctx, natGatewayList,
		client.InNamespace(m.cluster.Namespace()),
		client.MatchingLabels{
			machinebrokerv1alpha1.ManagerLabel:       machinebrokerv1alpha1.MachineBrokerManager,
			machinebrokerv1alpha1.CreatedLabel:       "true",
			machinebrokerv1alpha1.NetworkHandleLabel: key.networkHandle,
			machinebrokerv1alpha1.IPLabel:            apiutils.EscapeIP(key.ip),
		},
	); err != nil {
		return nil, nil, false, fmt.Errorf("error listing nat gateways by key: %w", err)
	}

	natGateways := m.filterNATGateways(natGatewayList.Items)

	switch len(natGateways) {
	case 0:
		return nil, nil, false, nil
	case 1:
		natGateway := natGateways[0]
		natGatewayRouting := &networkingv1alpha1.NATGatewayRouting{}
		if err := m.cluster.Client().Get(ctx, client.ObjectKeyFromObject(&natGateway), natGatewayRouting); err != nil {
			return nil, nil, false, fmt.Errorf("error getting nat gateway routing: %w", err)
		}
		return &natGateway, natGatewayRouting, true, nil
	default:
		return nil, nil, false, fmt.Errorf("multiple nat gateways found for key %s", key)
	}
}

func (m *NATGateways) createNATGateway(
	ctx context.Context,
	key natGatewayKey,
	network *networkingv1alpha1.Network,
) (resPrefix *networkingv1alpha1.NATGateway, resPrefixRouting *networkingv1alpha1.NATGatewayRouting, retErr error) {
	c := cleaner.New()
	defer cleaner.CleanupOnError(ctx, c, &retErr)

	natGateway := &networkingv1alpha1.NATGateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: m.cluster.Namespace(),
			Name:      m.cluster.IDGen().Generate(),
		},
		Spec: networkingv1alpha1.NATGatewaySpec{
			Type:       networkingv1alpha1.NATGatewayTypePublic,
			IPFamilies: []corev1.IPFamily{key.ip.Family()},
			NetworkRef: corev1.LocalObjectReference{Name: network.Name},
			IPs:        []networkingv1alpha1.NATGatewayIP{{Name: natGatewayIPName}},
		},
	}
	annotations.SetExternallyMangedBy(natGateway, machinebrokerv1alpha1.MachineBrokerManager)
	apiutils.SetManagerLabel(natGateway, machinebrokerv1alpha1.MachineBrokerManager)
	apiutils.SetNetworkHandle(natGateway, network.Spec.Handle)
	apiutils.SetIPLabel(natGateway, key.ip)

	if err := m.cluster.Client().Create(ctx, natGateway); err != nil {
		return nil, nil, fmt.Errorf("error creating nat gateway: %w", err)
	}
	c.Add(cleaner.DeleteObjectIfExistsFunc(m.cluster.Client(), natGateway))

	natGatewayBase := natGateway.DeepCopy()
	natGateway.Status.IPs = []networkingv1alpha1.NATGatewayIPStatus{
		{Name: natGatewayIPName, IP: key.ip},
	}
	if err := m.cluster.Client().Status().Patch(ctx, natGateway, client.MergeFrom(natGatewayBase)); err != nil {
		return nil, nil, fmt.Errorf("error patching nat gateway status: %w", err)
	}

	natGatewayRouting := &networkingv1alpha1.NATGatewayRouting{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: m.cluster.Namespace(),
			Name:      natGateway.Name,
		},
		NetworkRef: commonv1alpha1.LocalUIDReference{
			Name: network.Name,
			UID:  network.UID,
		},
	}
	apiutils.SetManagerLabel(natGatewayRouting, machinebrokerv1alpha1.MachineBrokerManager)
	apiutils.SetNetworkHandle(natGatewayRouting, network.Spec.Handle)
	apiutils.SetIPLabel(natGateway, key.ip)
	if err := ctrl.SetControllerReference(natGateway, natGatewayRouting, m.cluster.Scheme()); err != nil {
		return nil, nil, fmt.Errorf("error setting nat gateway routing to be controlled by nat gateway: %w", err)
	}

	if err := m.cluster.Client().Create(ctx, natGatewayRouting); err != nil {
		return nil, nil, fmt.Errorf("error creating nat gateway routing: %w", err)
	}

	return natGateway, natGatewayRouting, nil
}

func (m *NATGateways) removeNATGatewayRoutingDestination(
	ctx context.Context,
	natGatewayRouting *networkingv1alpha1.NATGatewayRouting,
	networkInterface *networkingv1alpha1.NetworkInterface,
) error {
	idx := slices.IndexFunc(natGatewayRouting.Destinations,
		func(ref networkingv1alpha1.NATGatewayDestination) bool { return ref.UID == networkInterface.UID },
	)
	if idx == -1 {
		return nil
	}

	base := natGatewayRouting.DeepCopy()
	natGatewayRouting.Destinations = slices.Delete(natGatewayRouting.Destinations, idx, idx+1)
	if err := m.cluster.Client().Patch(ctx, natGatewayRouting, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error removing nat gateway routing destination: %w", err)
	}
	return nil
}

func (m *NATGateways) addNATGatewayRoutingDestination(
	ctx context.Context,
	natGatewayRouting *networkingv1alpha1.NATGatewayRouting,
	networkInterface *networkingv1alpha1.NetworkInterface,
	tgt machinebrokerv1alpha1.NATGatewayTarget,
) error {
	idx := slices.IndexFunc(natGatewayRouting.Destinations,
		func(ref networkingv1alpha1.NATGatewayDestination) bool { return ref.UID == networkInterface.UID },
	)
	if idx >= 0 {
		return nil
	}

	base := natGatewayRouting.DeepCopy()
	natGatewayRouting.Destinations = append(natGatewayRouting.Destinations, networkingv1alpha1.NATGatewayDestination{
		Name: networkInterface.Name,
		UID:  networkInterface.UID,
		IPs: []networkingv1alpha1.NATGatewayDestinationIP{
			{
				IP:      tgt.IP,
				Port:    tgt.Port,
				EndPort: tgt.EndPort,
			},
		},
	})
	if err := m.cluster.Client().Patch(ctx, natGatewayRouting, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error adding nat gateway routing destination: %w", err)
	}
	return nil
}

func (m *NATGateways) Create(
	ctx context.Context,
	network *networkingv1alpha1.Network,
	tgt machinebrokerv1alpha1.NATGatewayTarget,
	networkInterface *networkingv1alpha1.NetworkInterface,
) error {
	key := natGatewayKey{
		networkHandle: network.Spec.Handle,
		ip:            tgt.IP,
	}
	m.mu.Lock(key)
	defer m.mu.Unlock(key)

	c := cleaner.New()

	natGateway, natGatewayRouting, ok, err := m.getNATGatewayByKey(ctx, key)
	if err != nil {
		return fmt.Errorf("error getting nat gateway by key %s: %w", key, err)
	}
	if !ok {
		newNATGateway, newNATGatewayRouting, err := m.createNATGateway(ctx, key, network)
		if err != nil {
			return fmt.Errorf("error creating nat gateway: %w", err)
		}

		c.Add(cleaner.DeleteObjectIfExistsFunc(m.cluster.Client(), newNATGateway))
		natGateway = newNATGateway
		natGatewayRouting = newNATGatewayRouting
	}

	if err := m.addNATGatewayRoutingDestination(ctx, natGatewayRouting, networkInterface, tgt); err != nil {
		return err
	}
	c.Add(func(ctx context.Context) error {
		return m.removeNATGatewayRoutingDestination(ctx, natGatewayRouting, networkInterface)
	})

	if err := apiutils.PatchCreatedWithDependent(ctx, m.cluster.Client(), natGateway, networkInterface.GetName()); err != nil {
		return fmt.Errorf("error patching created with dependent: %w", err)
	}
	return nil
}

func (m *NATGateways) Delete(
	ctx context.Context,
	networkHandle string,
	ip commonv1alpha1.IP,
	networkInterface *networkingv1alpha1.NetworkInterface,
) error {
	key := natGatewayKey{
		networkHandle: networkHandle,
		ip:            ip,
	}
	m.mu.Lock(key)
	defer m.mu.Unlock(key)

	natGateway, natGatewayRouting, ok, err := m.getNATGatewayByKey(ctx, key)
	if err != nil {
		return fmt.Errorf("error getting nat gateway by key: %w", err)
	}
	if !ok {
		return nil
	}

	var errs []error
	if err := m.removeNATGatewayRoutingDestination(ctx, natGatewayRouting, networkInterface); err != nil {
		errs = append(errs, fmt.Errorf("error removing nat gateway routing destination: %w", err))
	}
	if err := apiutils.DeleteAndGarbageCollect(ctx, m.cluster.Client(), natGateway, networkInterface.Name); err != nil {
		errs = append(errs, fmt.Errorf("error deleting / garbage collecting: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("error(s) deleting nat gateway: %v", errs)
	}
	return nil
}

func (m *NATGateways) listNATGateways(ctx context.Context, dependent string) ([]networkingv1alpha1.NATGateway, error) {
	natGatewayList := &networkingv1alpha1.NATGatewayList{}
	if err := m.cluster.Client().List(ctx, natGatewayList,
		client.InNamespace(m.cluster.Namespace()),
		client.MatchingLabels{
			machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.MachineBrokerManager,
			machinebrokerv1alpha1.CreatedLabel: "true",
		},
	); err != nil {
		return nil, fmt.Errorf("error listing nat gateways: %w", err)
	}

	if dependent == "" {
		return natGatewayList.Items, nil
	}

	filtered, err := apiutils.FilterObjectListByDependent(natGatewayList.Items, dependent)
	if err != nil {
		return nil, fmt.Errorf("error filtering by dependent: %w", err)
	}
	return filtered, nil
}

func (m *NATGateways) listNATGatewayRoutings(ctx context.Context) ([]networkingv1alpha1.NATGatewayRouting, error) {
	natGatewayRoutingList := &networkingv1alpha1.NATGatewayRoutingList{}
	if err := m.cluster.Client().List(ctx, natGatewayRoutingList,
		client.InNamespace(m.cluster.Namespace()),
		client.MatchingLabels{
			machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.MachineBrokerManager,
		},
	); err != nil {
		return nil, fmt.Errorf("error listing nat gateway routings: %w", err)
	}

	return natGatewayRoutingList.Items, nil
}

func (m *NATGateways) List(ctx context.Context) ([]machinebrokerv1alpha1.NATGateway, error) {
	natGateways, err := m.listNATGateways(ctx, "")
	if err != nil {
		return nil, err
	}

	natGatewayRoutings, err := m.listNATGatewayRoutings(ctx)
	if err != nil {
		return nil, err
	}

	return m.joinNATGatewaysAndRoutings(natGateways, natGatewayRoutings), nil
}

func (m *NATGateways) getNATGatewayIP(natGateway *networkingv1alpha1.NATGateway) (commonv1alpha1.IP, bool) {
	natGatewayIP, ok := utilslices.FindFunc(natGateway.Status.IPs,
		func(ip networkingv1alpha1.NATGatewayIPStatus) bool {
			return ip.Name == natGatewayIPName
		},
	)
	if !ok {
		return commonv1alpha1.IP{}, false
	}
	return natGatewayIP.IP, true
}

func (m *NATGateways) joinNATGatewaysAndRoutings(
	natGateways []networkingv1alpha1.NATGateway,
	natGatewayRoutings []networkingv1alpha1.NATGatewayRouting,
) []machinebrokerv1alpha1.NATGateway {
	natGatewayRoutingByName := utils.ObjectSliceToMapByName(natGatewayRoutings)

	var res []machinebrokerv1alpha1.NATGateway
	for i := range natGateways {
		natGateway := &natGateways[i]

		natGatewayRouting, ok := natGatewayRoutingByName[natGateway.Name]
		if !ok {
			continue
		}

		ip, ok := m.getNATGatewayIP(natGateway)
		if !ok {
			continue
		}

		networkHandle, ok := natGateway.Labels[machinebrokerv1alpha1.NetworkHandleLabel]
		if !ok {
			continue
		}

		destinations := utilslices.Map(
			natGatewayRouting.Destinations,
			func(dest networkingv1alpha1.NATGatewayDestination) machinebrokerv1alpha1.NATGatewayDestination {
				return machinebrokerv1alpha1.NATGatewayDestination{
					ID:      dest.Name,
					Port:    dest.IPs[0].Port, // TODO: Verify that we actually only have a single destination.
					EndPort: dest.IPs[0].EndPort,
				}
			},
		)

		res = append(res, machinebrokerv1alpha1.NATGateway{
			NetworkHandle: networkHandle,
			IP:            ip,
			Destinations:  destinations,
		})
	}
	return res
}

func (m *NATGateways) ListByDependent(ctx context.Context, dependent string) ([]machinebrokerv1alpha1.NATGateway, error) {
	natGateways, err := m.listNATGateways(ctx, dependent)
	if err != nil {
		return nil, err
	}

	natGatewayRoutings, err := m.listNATGatewayRoutings(ctx)
	if err != nil {
		return nil, err
	}

	return m.joinNATGatewaysAndRoutings(natGateways, natGatewayRoutings), nil
}
