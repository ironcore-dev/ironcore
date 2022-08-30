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

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/onmetal/onmetal-api/machinepoollet/endpoints"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

type EndpointsType string

const (
	EndpointsTypeNone           EndpointsType = "None"
	EndpointsTypeNodePort       EndpointsType = "NodePort"
	EndpointsTypeLoadBalancer   EndpointsType = "LoadBalancer"
	EndpointsTypeStaticHostname EndpointsType = "StaticHostName"
)

var (
	AvailableEndpointsTypes = []EndpointsType{
		EndpointsTypeNone,
		EndpointsTypeNodePort,
		EndpointsTypeLoadBalancer,
		EndpointsTypeStaticHostname,
	}
	AvailableNodeAddressTypes = []string{
		string(corev1.NodeHostName),
		string(corev1.NodeInternalIP),
		string(corev1.NodeInternalDNS),
		string(corev1.NodeExternalIP),
		string(corev1.NodeExternalDNS),
	}
)

type EndpointsOptions struct {
	Type             EndpointsType
	Namespace        string
	ServiceName      string
	PortName         string
	NodeAddressTypes []string
	StaticHostname   string
	StaticPort       int32
}

func NewEndpointsOptions() *EndpointsOptions {
	return &EndpointsOptions{
		Type: EndpointsTypeNone,
	}
}

func (o *EndpointsOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar((*string)(&o.Type), "endpoints-type", string(EndpointsTypeNone), fmt.Sprintf("Type to use for endpoints reconciliation. "+
		"Available endpoint types are: %v", AvailableEndpointsTypes))
	fs.StringVar(&o.Namespace, "endpoints-namespace", "", "Namespace for the endpoints controller. "+
		"If unspecified while running in Kubernetes this is auto-determined.")
	fs.StringVar(&o.ServiceName, "endpoints-service-name", "", "Name of the service to inspect for endpoints.")
	fs.StringVar(&o.PortName, "endpoints-port-name", "", "Name of the service port to inspect for endpoints.")
	fs.StringSliceVar(&o.NodeAddressTypes, "endpoints-node-address-types", AvailableNodeAddressTypes, "Node address types to use.")
	fs.StringVar(&o.StaticHostname, "static-host-name", o.StaticHostname, "Static hostname to report.")
	fs.Int32Var(&o.StaticPort, "static-port", o.StaticPort, "Static port to report.")
}

func (o *EndpointsOptions) getNamespace() string {
	if o.Namespace != "" {
		return o.Namespace
	}

	namespace, _ := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	return strings.TrimSpace(string(namespace))
}

func (o *EndpointsOptions) nodeAddressTypes() []corev1.NodeAddressType {
	res := make([]corev1.NodeAddressType, 0, len(o.NodeAddressTypes))
	for _, addressType := range o.NodeAddressTypes {
		res = append(res, corev1.NodeAddressType(addressType))
	}
	return res
}

func (o *EndpointsOptions) NewEndpoints(ctx context.Context, cfg *rest.Config) (endpoints.Endpoints, error) {
	switch o.Type {
	case EndpointsTypeNone:
		return endpoints.NoEndpoints{}, nil
	case EndpointsTypeLoadBalancer:
		return endpoints.NewLoadBalancerServiceEndpoints(ctx, cfg, endpoints.LoadBalancerServiceEndpointsOptions{
			Namespace:   o.getNamespace(),
			ServiceName: o.ServiceName,
			PortName:    o.PortName,
		})
	case EndpointsTypeNodePort:
		return endpoints.NewNodePortServiceEndpoints(ctx, cfg, endpoints.NodePortServiceEndpointsOptions{
			Namespace:        o.getNamespace(),
			ServiceName:      o.ServiceName,
			PortName:         o.PortName,
			NodeAddressTypes: o.nodeAddressTypes(),
		})
	case EndpointsTypeStaticHostname:
		return endpoints.NewStaticHostName(o.StaticHostname, o.StaticPort)
	default:
		return nil, fmt.Errorf("unknown endpoints type %q", o.Type)
	}
}

// SetupEndpointsWithManager sets up endpoints.Endpoints with the ctrl.Manager if necessary.
func SetupEndpointsWithManager(eps endpoints.Endpoints, mgr ctrl.Manager) error {
	type WithManagerSetup interface {
		SetupWithManager(mgr ctrl.Manager) error
	}
	if withSetup, ok := eps.(WithManagerSetup); ok {
		return withSetup.SetupWithManager(mgr)
	}
	return nil
}
