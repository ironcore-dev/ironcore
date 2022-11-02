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
	"github.com/onmetal/onmetal-api/apiutils/equality"
	machinepoolletpredicate "github.com/onmetal/onmetal-api/machinepoollet/predicate"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// NodePortServiceEndpoints populates Endpoints from a Kubernetes Service of type=NodePort.
// It does so by fetching the service to determine the port and all nodes to determine their addresses.
type NodePortServiceEndpoints struct {
	client    client.Client
	listeners []Listener

	addressesValue atomic.Value
	portValue      atomic.Value

	namespace        string
	serviceName      string
	portName         string
	nodeAddressTypes []corev1.NodeAddressType
}

func (e *NodePortServiceEndpoints) AddListener(listener Listener) {
	e.listeners = append(e.listeners, listener)
}

func (e *NodePortServiceEndpoints) notify() {
	for _, listener := range e.listeners {
		listener.Enqueue()
	}
}

func (e *NodePortServiceEndpoints) GetEndpoints() (addresses []computev1alpha1.MachinePoolAddress, port int32) {
	return e.addressesValue.Load().([]computev1alpha1.MachinePoolAddress), e.portValue.Load().(int32)
}

var nodeAddressTypeToMachinePoolAddressType = map[corev1.NodeAddressType]computev1alpha1.MachinePoolAddressType{
	corev1.NodeExternalDNS: computev1alpha1.MachinePoolExternalDNS,
	corev1.NodeExternalIP:  computev1alpha1.MachinePoolExternalIP,
	corev1.NodeHostName:    computev1alpha1.MachinePoolHostName,
	corev1.NodeInternalDNS: computev1alpha1.MachinePoolInternalIP,
	corev1.NodeInternalIP:  computev1alpha1.MachinePoolInternalIP,
}

func (e *NodePortServiceEndpoints) supportsNodeAddressType(addressType corev1.NodeAddressType) bool {
	for _, supported := range e.nodeAddressTypes {
		if addressType == supported {
			return true
		}
	}
	return false
}

func (e *NodePortServiceEndpoints) determineAddresses(ctx context.Context) ([]computev1alpha1.MachinePoolAddress, error) {
	log := ctrl.LoggerFrom(ctx)

	log.V(1).Info("Listing nodes")
	nodeList := &corev1.NodeList{}
	if err := e.client.List(ctx, nodeList); err != nil {
		return nil, fmt.Errorf("error listing nodes: %w", err)
	}

	res := NewMachinePoolAddressSet()
	for _, node := range nodeList.Items {
		for _, nodeAddress := range node.Status.Addresses {
			if !e.supportsNodeAddressType(nodeAddress.Type) {
				continue
			}

			machinePoolAddressType, ok := nodeAddressTypeToMachinePoolAddressType[nodeAddress.Type]
			if !ok {
				log.V(1).Info("Could not translate node address type to machine pool address type",
					"NodeAddressType", nodeAddress.Type)
				continue
			}

			res.Insert(computev1alpha1.MachinePoolAddress{
				Type:    machinePoolAddressType,
				Address: nodeAddress.Address,
			})
		}
	}

	return res.List(), nil
}

func findPort(ports []corev1.ServicePort, name string) (corev1.ServicePort, bool) {
	for _, port := range ports {
		if port.Name == name {
			port := port
			return port, true
		}
	}
	return corev1.ServicePort{}, false
}

func (e *NodePortServiceEndpoints) determinePort(ctx context.Context) (int32, error) {
	service := &corev1.Service{}
	if err := e.client.Get(ctx, client.ObjectKey{Namespace: e.namespace, Name: e.serviceName}, service); err != nil {
		return 0, fmt.Errorf("error getting service: %w", err)
	}

	if service.Spec.Type != corev1.ServiceTypeNodePort {
		return 0, fmt.Errorf("service %s is not of type %s", service.Name, corev1.ServiceTypeNodePort)
	}

	port, ok := findPort(service.Spec.Ports, e.portName)
	if !ok {
		return 0, fmt.Errorf("port %s not found", e.portName)
	}

	if port.NodePort == 0 {
		return 0, fmt.Errorf("port %s does not have a node port allocated", e.portName)
	}

	return port.NodePort, nil
}

type NodePortServiceEndpointsOptions struct {
	Namespace        string
	ServiceName      string
	PortName         string
	NodeAddressTypes []corev1.NodeAddressType
}

func setNodePortServiceEndpointsOptionsDefaults(o *NodePortServiceEndpointsOptions) {
	if o.NodeAddressTypes == nil {
		o.NodeAddressTypes = []corev1.NodeAddressType{
			corev1.NodeExternalDNS,
			corev1.NodeExternalIP,
			corev1.NodeHostName,
			corev1.NodeInternalDNS,
			corev1.NodeInternalIP,
		}
	}
}

func NewNodePortServiceEndpoints(ctx context.Context, cfg *rest.Config, opts NodePortServiceEndpointsOptions) (*NodePortServiceEndpoints, error) {
	if opts.Namespace == "" {
		return nil, fmt.Errorf("must specify Namespace")
	}
	if opts.ServiceName == "" {
		return nil, fmt.Errorf("must specify ServiceName")
	}
	if opts.PortName == "" {
		return nil, fmt.Errorf("must specify PortName")
	}
	setNodePortServiceEndpointsOptionsDefaults(&opts)

	c, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}

	e := &NodePortServiceEndpoints{
		client:           c,
		namespace:        opts.Namespace,
		serviceName:      opts.ServiceName,
		portName:         opts.PortName,
		nodeAddressTypes: opts.NodeAddressTypes,
	}

	addresses, err := e.determineAddresses(ctx)
	if err != nil {
		return nil, err
	}

	port, err := e.determinePort(ctx)
	if err != nil {
		return nil, err
	}

	e.addressesValue.Store(addresses)
	e.portValue.Store(port)

	return e, nil
}

func (e *NodePortServiceEndpoints) SetupWithManager(mgr ctrl.Manager) error {
	if err := (&nodePortServiceEndpointsServiceReconciler{e}).SetupWithManager(mgr); err != nil {
		return err
	}
	if err := (&nodePortServiceEndpointsNodeReconciler{e}).SetupWithManager(mgr); err != nil {
		return err
	}
	return nil
}

type nodePortServiceEndpointsNodeReconciler struct {
	*NodePortServiceEndpoints
}

func (r *nodePortServiceEndpointsNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	addresses, err := r.determineAddresses(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	if oldV := r.addressesValue.Swap(addresses); !equality.Semantic.DeepEqual(addresses, oldV.([]computev1alpha1.MachinePoolAddress)) {
		r.notify()
	}

	return ctrl.Result{}, nil
}

func (r *nodePortServiceEndpointsNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("nodeportserviceendpoints.node").
		For(
			&corev1.Node{},
			builder.WithPredicates(predicate.Funcs{
				UpdateFunc: func(event event.UpdateEvent) bool {
					oldNode, newNode := event.ObjectOld.(*corev1.Node), event.ObjectNew.(*corev1.Node)
					return !equality.Semantic.DeepEqual(oldNode.Status.Addresses, newNode.Status.Addresses)
				},
			}),
		).
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		Complete(r)
}

type nodePortServiceEndpointsServiceReconciler struct {
	*NodePortServiceEndpoints
}

func (r *nodePortServiceEndpointsServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	port, err := r.determinePort(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	if oldV := r.portValue.Swap(port); oldV == nil || oldV.(int32) != port {
		r.notify()
	}
	return ctrl.Result{}, nil
}

func (r *nodePortServiceEndpointsServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("nodeportserviceendpoints.service").
		For(
			&corev1.Service{},
			builder.WithPredicates(
				machinepoolletpredicate.NamespaceNamePredicate(r.namespace, r.serviceName),
				predicate.Funcs{
					UpdateFunc: func(event event.UpdateEvent) bool {
						oldService, newService := event.ObjectOld.(*corev1.Service), event.ObjectNew.(*corev1.Service)
						oldPort, _ := findPort(oldService.Spec.Ports, r.portName)
						newPort, _ := findPort(newService.Spec.Ports, r.portName)
						return oldPort != newPort
					},
				},
			),
		).
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		Complete(r)
}
