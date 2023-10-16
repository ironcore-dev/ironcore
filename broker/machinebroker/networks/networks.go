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
	"hash/fnv"
	"sync"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/machinebroker/cluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Manager struct {
	started   bool
	startedMu sync.Mutex

	cluster cluster.Cluster

	queue workqueue.RateLimitingInterface

	waitersByProviderIDMu sync.Mutex
	waitersByProviderID   map[string]*waiter
}

func NewManager(cluster cluster.Cluster) *Manager {
	return &Manager{
		cluster:             cluster,
		queue:               workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		waitersByProviderID: make(map[string]*waiter),
	}
}

type waiter struct {
	network *networkingv1alpha1.Network
	error   error
	done    chan struct{}
}

func (e *Manager) getNetworkForProviderID(ctx context.Context, providerID string) (*networkingv1alpha1.Network, error) {
	networkList := &networkingv1alpha1.NetworkList{}
	if err := e.cluster.Client().List(ctx, networkList,
		client.InNamespace(e.cluster.Namespace()),
		client.MatchingLabels{machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.MachineBrokerManager},
	); err != nil {
		return nil, fmt.Errorf("error listing networks: %w", err)
	}

	var matching *networkingv1alpha1.Network
	for i := range networkList.Items {
		network := &networkList.Items[i]
		if network.Spec.ProviderID != providerID {
			// Ignore if the providerID doesn't match.
			continue
		}
		if !network.DeletionTimestamp.IsZero() {
			// Ignore if the network is already deleting.
			continue
		}

		if matching == nil || network.CreationTimestamp.Before(&matching.CreationTimestamp) {
			matching = network
		}
	}
	return matching, nil
}

func (e *Manager) providerIDHash(providerID string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(providerID))
	return rand.SafeEncodeString(fmt.Sprint(h.Sum32()))
}

func (e *Manager) getOrCreateNetworkForProviderID(ctx context.Context, log logr.Logger, providerID string) (*networkingv1alpha1.Network, error) {
	network, err := e.getNetworkForProviderID(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("error getting network for providerID: %w", err)
	}
	if network != nil {
		log.V(1).Info("Found existing network for providerID", "Name", network.Name)
		return network, nil
	}

	log.V(1).Info("No network found for providerID, creating a new one")
	network = &networkingv1alpha1.Network{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: e.cluster.Namespace(),
			// TODO: Make this evaluator use a cache and include a collision count in the providerID hash, use
			// name instead of generateName. This makes it possible to providerID fast resyncs similar to
			// Kubernetes' deployment controller.
			GenerateName: fmt.Sprintf("net-%s-", e.providerIDHash(providerID)),
			Annotations: map[string]string{
				commonv1alpha1.ManagedByAnnotation: machinebrokerv1alpha1.MachineBrokerManager,
			},
			Labels: map[string]string{
				machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.MachineBrokerManager,
			},
		},
		Spec: networkingv1alpha1.NetworkSpec{
			ProviderID: providerID,
		},
	}
	if err := e.cluster.Client().Create(ctx, network); err != nil {
		return nil, fmt.Errorf("error creating network: %w", err)
	}
	return network, nil
}

func (e *Manager) setNetworkAsAvailable(ctx context.Context, network *networkingv1alpha1.Network) error {
	baseNetwork := network.DeepCopy()
	network.Status.State = networkingv1alpha1.NetworkStateAvailable
	if err := e.cluster.Client().Status().Patch(ctx, network, client.MergeFrom(baseNetwork)); err != nil {
		return fmt.Errorf("error patching network status: %w", err)
	}
	return nil
}

func (e *Manager) doWork(ctx context.Context, providerID string) (*networkingv1alpha1.Network, error) {
	log := ctrl.LoggerFrom(ctx)
	network, err := e.getOrCreateNetworkForProviderID(ctx, log, providerID)
	if err != nil {
		return nil, fmt.Errorf("error getting / creating network for providerID: %w", err)
	}

	if network.Status.State != networkingv1alpha1.NetworkStateAvailable {
		log.V(1).Info("Setting network state to available")
		if err := e.setNetworkAsAvailable(ctx, network); err != nil {
			return nil, fmt.Errorf("error setting network as available: %w", err)
		}
	}

	log.V(1).Info("Returning available network")
	return network, nil
}

func (e *Manager) processNextWorkItem(ctx context.Context) bool {
	uncastProviderID, quit := e.queue.Get()
	if quit {
		return false
	}
	defer e.queue.Done(uncastProviderID)

	providerID := uncastProviderID.(string)
	network, err := e.doWork(ctx, providerID)
	e.emit(providerID, network, err)
	e.queue.Forget(providerID)
	return true
}

func (e *Manager) Start(ctx context.Context) error {
	e.startedMu.Lock()
	if e.started {
		e.startedMu.Unlock()
		return fmt.Errorf("manager already started")
	}

	var wg sync.WaitGroup
	func() {
		defer e.startedMu.Unlock()

		go func() {
			<-ctx.Done()
			e.queue.ShutDown()
		}()

		const maxConcurrentWork = 10
		wg.Add(maxConcurrentWork)
		for i := 0; i < maxConcurrentWork; i++ {
			go func() {
				defer wg.Done()

				for e.processNextWorkItem(ctx) {
				}
			}()
		}
	}()

	wg.Wait()
	return nil
}

func (e *Manager) getOrCreateWaiter(providerID string) *waiter {
	e.waitersByProviderIDMu.Lock()
	defer e.waitersByProviderIDMu.Unlock()

	w, ok := e.waitersByProviderID[providerID]
	if ok {
		return w
	}

	w = &waiter{done: make(chan struct{})}
	e.waitersByProviderID[providerID] = w
	e.queue.Add(providerID)
	return w
}

func (e *Manager) emit(providerID string, network *networkingv1alpha1.Network, err error) {
	e.waitersByProviderIDMu.Lock()
	defer e.waitersByProviderIDMu.Unlock()

	w, ok := e.waitersByProviderID[providerID]
	if !ok {
		return
	}

	w.network = network
	w.error = err
	close(w.done)
	// Remove providerID from waiters map once the emit call finishes. Otherwise, we won't be able to recreate
	// the network if it has been deleted (e.g. via GC when the network has been released due to the lack of consumers).
	delete(e.waitersByProviderID, providerID)
}

func (e *Manager) GetNetwork(ctx context.Context, providerID string) (*networkingv1alpha1.Network, error) {
	w := e.getOrCreateWaiter(providerID)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-w.done:
		return w.network, w.error
	}
}
