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

package networks

import (
	"context"
	"fmt"

	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/common/cleaner"
	commonsync "github.com/onmetal/onmetal-api/broker/common/sync"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/machinebroker/apiutils"
	"github.com/onmetal/onmetal-api/broker/machinebroker/cluster"
	"github.com/onmetal/onmetal-api/broker/machinebroker/transaction"
	"github.com/onmetal/onmetal-api/utils/annotations"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Networks struct {
	mu commonsync.MutexMap[string]

	cluster cluster.Cluster
}

func New(cluster cluster.Cluster) *Networks {
	return &Networks{
		mu:      *commonsync.NewMutexMap[string](),
		cluster: cluster,
	}
}

func (m *Networks) filterNetworks(networks []networkingv1alpha1.Network) []networkingv1alpha1.Network {
	var filtered []networkingv1alpha1.Network
	for _, network := range networks {
		if !network.DeletionTimestamp.IsZero() {
			continue
		}

		filtered = append(filtered, network)
	}
	return filtered
}

func (m *Networks) getNetworkByHandle(ctx context.Context, handle string) (*networkingv1alpha1.Network, bool, error) {
	networkList := &networkingv1alpha1.NetworkList{}
	if err := m.cluster.Client().List(ctx, networkList,
		client.InNamespace(m.cluster.Namespace()),
		client.MatchingLabels{
			machinebrokerv1alpha1.ManagerLabel:       machinebrokerv1alpha1.MachineBrokerManager,
			machinebrokerv1alpha1.CreatedLabel:       "true",
			machinebrokerv1alpha1.NetworkHandleLabel: handle,
		},
	); err != nil {
		return nil, false, fmt.Errorf("error listing networks by handle: %w", err)
	}

	networks := m.filterNetworks(networkList.Items)

	switch len(networks) {
	case 0:
		return nil, false, nil
	case 1:
		network := networks[0]
		return &network, true, nil
	default:
		return nil, false, fmt.Errorf("multiple networks found for handle %s", handle)
	}
}

func (m *Networks) createNetwork(ctx context.Context, handle string) (res *networkingv1alpha1.Network, retErr error) {
	c := cleaner.New()
	defer cleaner.CleanupOnError(ctx, c, &retErr)

	network := &networkingv1alpha1.Network{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: m.cluster.Namespace(),
			Name:      m.cluster.IDGen().Generate(),
		},
		Spec: networkingv1alpha1.NetworkSpec{
			Handle: handle,
		},
	}
	annotations.SetExternallyMangedBy(network, machinebrokerv1alpha1.MachineBrokerManager)
	apiutils.SetManagerLabel(network, machinebrokerv1alpha1.MachineBrokerManager)
	apiutils.SetNetworkHandle(network, handle)

	if err := m.cluster.Client().Create(ctx, network); err != nil {
		return nil, fmt.Errorf("error creating network: %w", err)
	}
	c.Add(cleaner.DeleteObjectIfExistsFunc(m.cluster.Client(), network))

	base := network.DeepCopy()
	network.Status.State = networkingv1alpha1.NetworkStateAvailable
	if err := m.cluster.Client().Status().Patch(ctx, network, client.MergeFrom(base)); err != nil {
		return nil, fmt.Errorf("error setting network to available: %w", err)
	}

	return network, nil
}

func (m *Networks) BeginCreate(ctx context.Context, handle string) (*networkingv1alpha1.Network, transaction.Transaction[client.Object], error) {
	log := ctrl.LoggerFrom(ctx)

	var network *networkingv1alpha1.Network
	t, err := transaction.Build(func(cb transaction.Callback[client.Object]) error {
		m.mu.Lock(handle)
		defer m.mu.Unlock(handle)

		c := cleaner.New()

		n, ok, err := m.getNetworkByHandle(ctx, handle)
		if err != nil {
			return fmt.Errorf("error getting network by handle %s: %w", handle, err)
		}
		if !ok {
			newNetwork, err := m.createNetwork(ctx, handle)
			if err != nil {
				return fmt.Errorf("error creating network: %w", err)
			}

			c.Add(cleaner.DeleteObjectIfExistsFunc(m.cluster.Client(), newNetwork))
			n = newNetwork
		}
		network = n

		obj, rollback := cb()
		if rollback {
			return c.Cleanup(ctx)
		}
		if err := apiutils.PatchCreatedWithDependent(ctx, m.cluster.Client(), network, obj.GetName()); err != nil {
			if err := c.Cleanup(ctx); err != nil {
				log.Error(err, "Error cleaning up")
			}
			return err
		}
		return nil
	})
	return network, t, err
}

func (m *Networks) Delete(ctx context.Context, handle, id string) error {
	m.mu.Lock(handle)
	defer m.mu.Unlock(handle)

	network, ok, err := m.getNetworkByHandle(ctx, handle)
	if err != nil {
		return fmt.Errorf("error getting network by handle: %w", err)
	}
	if !ok {
		return nil
	}

	return apiutils.DeleteAndGarbageCollect(ctx, m.cluster.Client(), network, id)
}
