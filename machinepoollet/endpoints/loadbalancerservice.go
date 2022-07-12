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

package endpoints

import (
	"context"
	"fmt"
	"sync/atomic"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	"github.com/onmetal/onmetal-api/equality"
	machinepoolletpredicate "github.com/onmetal/onmetal-api/machinepoollet/predicate"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type addressesAndPort struct {
	addresses []computev1alpha1.MachinePoolAddress
	port      int32
}

// LoadBalancerServiceEndpoints is an implementor of Endpoints that works by watching target Service of type=LoadBalancer.
type LoadBalancerServiceEndpoints struct {
	client client.Client

	listeners []Listener

	value atomic.Value

	namespace   string
	serviceName string
	portName    string
}

// LoadBalancerServiceEndpointsOptions are options for initializing LoadBalancerServiceEndpoints.
type LoadBalancerServiceEndpointsOptions struct {
	Namespace   string
	ServiceName string
	PortName    string
}

// NewLoadBalancerServiceEndpoints creates a new LoadBalancerServiceEndpoints that populates endpoints by a given Service of type=LoadBalancer.
func NewLoadBalancerServiceEndpoints(ctx context.Context, cfg *rest.Config, opts LoadBalancerServiceEndpointsOptions) (*LoadBalancerServiceEndpoints, error) {
	if opts.Namespace == "" {
		return nil, fmt.Errorf("must specify Namespace")
	}
	if opts.ServiceName == "" {
		return nil, fmt.Errorf("must specify ServiceName")
	}
	if opts.PortName == "" {
		return nil, fmt.Errorf("must specify PortName")
	}

	c, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}

	e := &LoadBalancerServiceEndpoints{
		client:      c,
		namespace:   opts.Namespace,
		serviceName: opts.ServiceName,
		portName:    opts.PortName,
	}

	addresses, port, err := e.loadAddressesAndPort(ctx)
	if err != nil {
		return nil, err
	}

	e.value.Store(addressesAndPort{addresses, port})

	return e, nil
}

// AddListener implements Notifier.
func (e *LoadBalancerServiceEndpoints) AddListener(listener Listener) {
	e.listeners = append(e.listeners, listener)
}

func (e *LoadBalancerServiceEndpoints) notify() {
	for _, listener := range e.listeners {
		listener.Enqueue()
	}
}

// GetEndpoints implements Endpoints.
func (e *LoadBalancerServiceEndpoints) GetEndpoints() (addresses []computev1alpha1.MachinePoolAddress, port int32) {
	val := e.value.Load().(addressesAndPort)
	return val.addresses, val.port
}

// SetupWithManager sets up the active components (continuous watching) of the LoadBalancerServiceEndpoints with the manager.
func (e *LoadBalancerServiceEndpoints) SetupWithManager(mgr ctrl.Manager) error {
	return (&loadBalancerServiceEndpointsReconciler{e}).SetupWithManager(mgr)
}

func (e *LoadBalancerServiceEndpoints) loadAddressesAndPort(ctx context.Context) ([]computev1alpha1.MachinePoolAddress, int32, error) {
	service := &corev1.Service{}
	if err := e.client.Get(ctx, client.ObjectKey{Namespace: e.namespace, Name: e.serviceName}, service); err != nil {
		return nil, 0, fmt.Errorf("error getting service: %w", err)
	}

	if service.Spec.Type != corev1.ServiceTypeLoadBalancer {
		return nil, 0, fmt.Errorf("service %s is not of type %s", service.Name, corev1.ServiceTypeLoadBalancer)
	}

	port, ok := findPort(service.Spec.Ports, e.portName)
	if !ok {
		return nil, 0, fmt.Errorf("service %s does not have a port %s", service.Name, e.portName)
	}

	res := NewMachinePoolAddressSet()
	for _, ingress := range service.Status.LoadBalancer.Ingress {
		switch {
		case ingress.IP != "":
			res.Insert(computev1alpha1.MachinePoolAddress{
				Type:    computev1alpha1.MachinePoolExternalIP,
				Address: ingress.IP,
			})
		case ingress.Hostname != "":
			res.Insert(computev1alpha1.MachinePoolAddress{
				Type:    computev1alpha1.MachinePoolExternalDNS,
				Address: ingress.Hostname,
			})
		}
	}

	return res.List(), port.Port, nil
}

type loadBalancerServiceEndpointsReconciler struct {
	*LoadBalancerServiceEndpoints
}

func (r *loadBalancerServiceEndpointsReconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	addresses, port, err := r.loadAddressesAndPort(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	if old := r.value.Swap(addressesAndPort{addresses, port}).(addressesAndPort); old.port != port || !equality.Semantic.DeepEqual(old.addresses, addresses) {
		r.notify()
	}

	return ctrl.Result{}, nil
}

func (r *loadBalancerServiceEndpointsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("loadbalancerserviceendpoints").
		For(
			&corev1.Service{},
			builder.WithPredicates(
				machinepoolletpredicate.NamespaceNamePredicate(r.namespace, r.serviceName),
				predicate.Funcs{
					UpdateFunc: func(event event.UpdateEvent) bool {
						oldSvc, newSvc := event.ObjectOld.(*corev1.Service), event.ObjectNew.(*corev1.Service)
						oldPort, _ := findPort(oldSvc.Spec.Ports, r.portName)
						newPort, _ := findPort(newSvc.Spec.Ports, r.portName)
						return oldPort.Port != newPort.Port || !equality.Semantic.DeepEqual(oldSvc.Status.LoadBalancer, newSvc.Status.LoadBalancer)
					},
				},
			),
		).
		Complete(r)
}
