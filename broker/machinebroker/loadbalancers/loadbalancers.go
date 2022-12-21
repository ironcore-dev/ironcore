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

package loadbalancers

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

type loadBalancerKey struct {
	networkHandle string
	target        machinebrokerv1alpha1.LoadBalancerTarget
}

func (k loadBalancerKey) Key() string {
	return fmt.Sprintf("%s/%v", k.networkHandle, k.target)
}

type LoadBalancers struct {
	mu commonsync.MutexMap[string]

	cluster cluster.Cluster
}

func New(cluster cluster.Cluster) *LoadBalancers {
	return &LoadBalancers{
		mu:      *commonsync.NewMutexMap[string](),
		cluster: cluster,
	}
}

func (m *LoadBalancers) filterLoadBalancers(
	loadBalancers []networkingv1alpha1.LoadBalancer,
	ports []machinebrokerv1alpha1.LoadBalancerPort,
) []networkingv1alpha1.LoadBalancer {
	portsKey := machinebrokerv1alpha1.LoadBalancerPortsKey(ports)

	var filtered []networkingv1alpha1.LoadBalancer
	for _, loadBalancer := range loadBalancers {
		if loadBalancer.DeletionTimestamp.IsZero() {
			continue
		}

		targetPorts := apiutils.ConvertNetworkingLoadBalancerPortsToLoadBalancerPorts(loadBalancer.Spec.Ports)
		if machinebrokerv1alpha1.LoadBalancerPortsKey(targetPorts) != portsKey {
			continue
		}

		filtered = append(filtered, loadBalancer)
	}
	return filtered
}

func (m *LoadBalancers) getLoadBalancerByKey(ctx context.Context, key loadBalancerKey) (*networkingv1alpha1.LoadBalancer, *networkingv1alpha1.LoadBalancerRouting, bool, error) {
	loadBalancerList := &networkingv1alpha1.LoadBalancerList{}
	if err := m.cluster.Client().List(ctx, loadBalancerList,
		client.InNamespace(m.cluster.Namespace()),
		client.MatchingLabels{
			machinebrokerv1alpha1.ManagerLabel:       machinebrokerv1alpha1.MachineBrokerManager,
			machinebrokerv1alpha1.CreatedLabel:       "true",
			machinebrokerv1alpha1.NetworkHandleLabel: key.networkHandle,
			machinebrokerv1alpha1.IPLabel:            apiutils.EscapeIP(key.target.IP),
		},
	); err != nil {
		return nil, nil, false, fmt.Errorf("error listing load balanceres by key: %w", err)
	}

	loadBalancers := m.filterLoadBalancers(loadBalancerList.Items, key.target.Ports)

	switch len(loadBalancerList.Items) {
	case 0:
		return nil, nil, false, nil
	case 1:
		loadBalancer := loadBalancers[0]
		loadBalancerRouting := &networkingv1alpha1.LoadBalancerRouting{}
		if err := m.cluster.Client().Get(ctx, client.ObjectKeyFromObject(&loadBalancer), loadBalancerRouting); err != nil {
			return nil, nil, false, fmt.Errorf("error getting load balancer routing: %w", err)
		}
		return &loadBalancer, loadBalancerRouting, true, nil
	default:
		return nil, nil, false, fmt.Errorf("multiple load balanceres found for key %v", key)
	}
}

func (m *LoadBalancers) createLoadBalancer(
	ctx context.Context,
	key loadBalancerKey,
	network *networkingv1alpha1.Network,
) (resPrefix *networkingv1alpha1.LoadBalancer, resPrefixRouting *networkingv1alpha1.LoadBalancerRouting, retErr error) {
	c := cleaner.New()
	defer cleaner.CleanupOnError(ctx, c, &retErr)

	ports := apiutils.ConvertLoadBalancerPortsToNetworkingLoadBalancerPorts(key.target.Ports)

	loadBalancer := &networkingv1alpha1.LoadBalancer{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: m.cluster.Namespace(),
			Name:      m.cluster.IDGen().Generate(),
		},
		Spec: networkingv1alpha1.LoadBalancerSpec{
			Type:       networkingv1alpha1.LoadBalancerTypePublic,
			IPFamilies: []corev1.IPFamily{key.target.IP.Family()},
			NetworkRef: corev1.LocalObjectReference{Name: network.Name},
			Ports:      ports,
		},
	}
	annotations.SetExternallyMangedBy(loadBalancer, machinebrokerv1alpha1.MachineBrokerManager)
	apiutils.SetManagerLabel(loadBalancer, machinebrokerv1alpha1.MachineBrokerManager)
	apiutils.SetNetworkHandle(loadBalancer, key.networkHandle)
	apiutils.SetIPLabel(loadBalancer, key.target.IP)

	if err := m.cluster.Client().Create(ctx, loadBalancer); err != nil {
		return nil, nil, fmt.Errorf("error creating load balancer: %w", err)
	}
	c.Add(cleaner.DeleteObjectIfExistsFunc(m.cluster.Client(), loadBalancer))

	baseLoadBalancer := loadBalancer.DeepCopy()
	loadBalancer.Status.IPs = []commonv1alpha1.IP{key.target.IP}
	if err := m.cluster.Client().Status().Patch(ctx, loadBalancer, client.MergeFrom(baseLoadBalancer)); err != nil {
		return nil, nil, fmt.Errorf("error patching load balancer status ips: %w", err)
	}

	loadBalancerRouting := &networkingv1alpha1.LoadBalancerRouting{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: m.cluster.Namespace(),
			Name:      loadBalancer.Name,
		},
		NetworkRef: commonv1alpha1.LocalUIDReference{
			Name: network.Name,
			UID:  network.UID,
		},
	}
	apiutils.SetManagerLabel(loadBalancerRouting, machinebrokerv1alpha1.MachineBrokerManager)
	apiutils.SetNetworkHandle(loadBalancerRouting, key.networkHandle)
	apiutils.SetIPLabel(loadBalancerRouting, key.target.IP)
	if err := ctrl.SetControllerReference(loadBalancer, loadBalancerRouting, m.cluster.Scheme()); err != nil {
		return nil, nil, fmt.Errorf("error setting load balancer routing to be controlled by load balancer: %w", err)
	}

	if err := m.cluster.Client().Create(ctx, loadBalancerRouting); err != nil {
		return nil, nil, fmt.Errorf("error creating load balancer routing: %w", err)
	}

	return loadBalancer, loadBalancerRouting, nil
}

func (m *LoadBalancers) removeLoadBalancerRoutingDestination(
	ctx context.Context,
	loadBalancerRouting *networkingv1alpha1.LoadBalancerRouting,
	obj client.Object,
) error {
	idx := slices.IndexFunc(loadBalancerRouting.Destinations,
		func(ref commonv1alpha1.LocalUIDReference) bool { return ref.UID == obj.GetUID() },
	)
	if idx == -1 {
		return nil
	}

	base := loadBalancerRouting.DeepCopy()
	loadBalancerRouting.Destinations = slices.Delete(loadBalancerRouting.Destinations, idx, idx+1)
	if err := m.cluster.Client().Patch(ctx, loadBalancerRouting, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error removing load balancer routing destination: %w", err)
	}
	return nil
}

func (m *LoadBalancers) addLoadBalancerRoutingDestination(
	ctx context.Context,
	loadBalancerRouting *networkingv1alpha1.LoadBalancerRouting,
	obj client.Object,
) error {
	idx := slices.IndexFunc(loadBalancerRouting.Destinations,
		func(ref commonv1alpha1.LocalUIDReference) bool { return ref.UID == obj.GetUID() },
	)
	if idx >= 0 {
		return nil
	}

	base := loadBalancerRouting.DeepCopy()
	loadBalancerRouting.Destinations = append(loadBalancerRouting.Destinations, commonv1alpha1.LocalUIDReference{
		Name: obj.GetName(),
		UID:  obj.GetUID(),
	})
	if err := m.cluster.Client().Patch(ctx, loadBalancerRouting, client.MergeFrom(base)); err != nil {
		return fmt.Errorf("error adding load balancer routing destination: %w", err)
	}
	return nil
}

func (m *LoadBalancers) Create(
	ctx context.Context,
	network *networkingv1alpha1.Network,
	tgt machinebrokerv1alpha1.LoadBalancerTarget,
	networkInterface *networkingv1alpha1.NetworkInterface,
) error {
	key := loadBalancerKey{
		networkHandle: network.Spec.Handle,
		target:        tgt,
	}
	m.mu.Lock(key.Key())
	defer m.mu.Unlock(key.Key())

	c := cleaner.New()

	loadBalancer, loadBalancerRouting, ok, err := m.getLoadBalancerByKey(ctx, key)
	if err != nil {
		return fmt.Errorf("error getting load balancer by key %v: %w", key, err)
	}
	if !ok {
		newLoadBalancer, newLoadBalancerRouting, err := m.createLoadBalancer(ctx, key, network)
		if err != nil {
			return fmt.Errorf("error creating load balancer: %w", err)
		}

		c.Add(cleaner.DeleteObjectIfExistsFunc(m.cluster.Client(), newLoadBalancer))
		loadBalancer = newLoadBalancer
		loadBalancerRouting = newLoadBalancerRouting
	}

	if err := m.addLoadBalancerRoutingDestination(ctx, loadBalancerRouting, networkInterface); err != nil {
		return err
	}
	c.Add(func(ctx context.Context) error {
		return m.removeLoadBalancerRoutingDestination(ctx, loadBalancerRouting, networkInterface)
	})

	if err := apiutils.PatchCreatedWithDependent(ctx, m.cluster.Client(), loadBalancer, networkInterface.GetName()); err != nil {
		return fmt.Errorf("error patching created with dependent: %w", err)
	}
	return nil
}

func (m *LoadBalancers) Delete(
	ctx context.Context,
	networkHandle string,
	tgt machinebrokerv1alpha1.LoadBalancerTarget,
	obj client.Object,
) error {
	key := loadBalancerKey{
		networkHandle: networkHandle,
		target:        tgt,
	}
	m.mu.Lock(key.Key())
	defer m.mu.Unlock(key.Key())

	loadBalancer, loadBalancerRouting, ok, err := m.getLoadBalancerByKey(ctx, key)
	if err != nil {
		return fmt.Errorf("error getting load balancer by key: %w", err)
	}
	if !ok {
		return nil
	}

	var errs []error
	if err := m.removeLoadBalancerRoutingDestination(ctx, loadBalancerRouting, obj); err != nil {
		errs = append(errs, fmt.Errorf("error removing load balancer routing destination: %w", err))
	}
	if err := apiutils.DeleteAndGarbageCollect(ctx, m.cluster.Client(), loadBalancer, obj.GetName()); err != nil {
		errs = append(errs, fmt.Errorf("error deleting / garbage collecting: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("error(s) deleting load balancer: %v", errs)
	}
	return nil
}

func (m *LoadBalancers) listLoadBalancers(ctx context.Context, dependent string) ([]networkingv1alpha1.LoadBalancer, error) {
	loadBalancerList := &networkingv1alpha1.LoadBalancerList{}
	if err := m.cluster.Client().List(ctx, loadBalancerList,
		client.InNamespace(m.cluster.Namespace()),
		client.MatchingLabels{
			machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.MachineBrokerManager,
			machinebrokerv1alpha1.CreatedLabel: "true",
		},
	); err != nil {
		return nil, fmt.Errorf("error listing load balanceres: %w", err)
	}

	if dependent == "" {
		return loadBalancerList.Items, nil
	}

	filtered, err := apiutils.FilterObjectListByDependent(loadBalancerList.Items, dependent)
	if err != nil {
		return nil, fmt.Errorf("error filtering by dependent: %w", err)
	}
	return filtered, nil
}

func (m *LoadBalancers) listLoadBalancerRoutings(ctx context.Context) ([]networkingv1alpha1.LoadBalancerRouting, error) {
	loadBalancerRoutingList := &networkingv1alpha1.LoadBalancerRoutingList{}
	if err := m.cluster.Client().List(ctx, loadBalancerRoutingList,
		client.InNamespace(m.cluster.Namespace()),
		client.MatchingLabels{
			machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.MachineBrokerManager,
		},
	); err != nil {
		return nil, fmt.Errorf("error listing load balancer routings: %w", err)
	}

	return loadBalancerRoutingList.Items, nil
}

func (m *LoadBalancers) List(ctx context.Context) ([]machinebrokerv1alpha1.LoadBalancer, error) {
	loadBalancers, err := m.listLoadBalancers(ctx, "")
	if err != nil {
		return nil, err
	}

	loadBalancerRoutings, err := m.listLoadBalancerRoutings(ctx)
	if err != nil {
		return nil, err
	}

	return m.joinLoadBalancersAndRoutings(loadBalancers, loadBalancerRoutings), nil
}

func (m *LoadBalancers) joinLoadBalancersAndRoutings(
	loadBalancers []networkingv1alpha1.LoadBalancer,
	loadBalancerRoutings []networkingv1alpha1.LoadBalancerRouting,
) []machinebrokerv1alpha1.LoadBalancer {
	loadBalancerRoutingByName := utils.ObjectSliceToMapByName(loadBalancerRoutings)

	var res []machinebrokerv1alpha1.LoadBalancer
	for i := range loadBalancers {
		loadBalancer := &loadBalancers[i]

		ip, err := apiutils.GetIPLabel(loadBalancer)
		if err != nil {
			continue // TODO: Should we handle this case better?
		}

		loadBalancerRouting, ok := loadBalancerRoutingByName[loadBalancer.Name]
		if !ok {
			continue
		}

		networkHandle, ok := loadBalancer.Labels[machinebrokerv1alpha1.NetworkHandleLabel]
		if !ok {
			continue
		}

		destinations := utilslices.Map(
			loadBalancerRouting.Destinations,
			func(dest commonv1alpha1.LocalUIDReference) string {
				return dest.Name
			},
		)

		res = append(res, machinebrokerv1alpha1.LoadBalancer{
			NetworkHandle: networkHandle,
			IP:            ip,
			Ports:         apiutils.ConvertNetworkingLoadBalancerPortsToLoadBalancerPorts(loadBalancer.Spec.Ports),
			Destinations:  destinations,
		})
	}
	return res
}

func (m *LoadBalancers) ListByDependent(ctx context.Context, dependent string) ([]machinebrokerv1alpha1.LoadBalancer, error) {
	loadBalancers, err := m.listLoadBalancers(ctx, dependent)
	if err != nil {
		return nil, err
	}

	loadBalancerRoutings, err := m.listLoadBalancerRoutings(ctx)
	if err != nil {
		return nil, err
	}

	return m.joinLoadBalancersAndRoutings(loadBalancers, loadBalancerRoutings), nil
}
