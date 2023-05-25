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

package server

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/common/cleaner"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/broker/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/broker/machinebroker/apiutils"
	"github.com/onmetal/onmetal-api/broker/machinebroker/transaction"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func (s *Server) loggerFrom(ctx context.Context, keysWithValues ...interface{}) logr.Logger {
	return ctrl.LoggerFrom(ctx, keysWithValues...)
}

func (s *Server) listManagedAndCreated(ctx context.Context, list client.ObjectList) error {
	return s.cluster.Client().List(ctx, list,
		client.InNamespace(s.cluster.Namespace()),
		client.MatchingLabels{
			machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.MachineBrokerManager,
			machinebrokerv1alpha1.CreatedLabel: "true",
		},
	)
}

func (s *Server) listWithPurpose(ctx context.Context, list client.ObjectList, purpose string) error {
	return s.cluster.Client().List(ctx, list,
		client.InNamespace(s.cluster.Namespace()),
		client.MatchingLabels{
			machinebrokerv1alpha1.PurposeLabel: purpose,
		},
	)
}

func (s *Server) getManagedAndCreated(ctx context.Context, name string, obj client.Object) error {
	key := client.ObjectKey{Namespace: s.cluster.Namespace(), Name: name}
	if err := s.cluster.Client().Get(ctx, key, obj); err != nil {
		return err
	}
	if !apiutils.IsManagedBy(obj, machinebrokerv1alpha1.MachineBrokerManager) || !apiutils.IsCreated(obj) {
		gvk, err := apiutil.GVKForObject(obj, s.cluster.Client().Scheme())
		if err != nil {
			return err
		}

		return apierrors.NewNotFound(schema.GroupResource{
			Group:    gvk.Group,
			Resource: gvk.Kind, // Yes, kind is good enough here
		}, key.Name)
	}
	return nil
}

func (s *Server) setupCleaner(ctx context.Context, log logr.Logger, retErr *error) (c *cleaner.Cleaner, cleanup func()) {
	c = cleaner.New()
	cleanup = func() {
		if *retErr != nil {
			select {
			case <-ctx.Done():
				log.Info("Cannot do cleanup since context expired")
				return
			default:
				if err := c.Cleanup(ctx); err != nil {
					log.Error(err, "Error cleaning up")
				}
			}
		}
	}
	return c, cleanup
}

func (s *Server) convertOnmetalIPSourcesToIPs(ipSources []networkingv1alpha1.IPSource) ([]string, error) {
	res := make([]string, len(ipSources))
	for i, ipSource := range ipSources {
		if ipSource.Value == nil {
			return nil, fmt.Errorf("ip source %d does not specify an ip literal", i)
		}
		res[i] = ipSource.Value.String()
	}
	return res, nil
}

func (s *Server) convertOnmetalPrefixSourcesToPrefixes(prefixSources []networkingv1alpha1.PrefixSource) ([]string, error) {
	res := make([]string, len(prefixSources))
	for i, prefixSource := range prefixSources {
		if prefixSource.Value == nil {
			return nil, fmt.Errorf("prefix source %d does not specify a prefix literal", i)
		}
		res[i] = prefixSource.Value.String()
	}
	return res, nil
}

func (s *Server) convertOnmetalLoadBalancerType(typ networkingv1alpha1.LoadBalancerType) (ori.LoadBalancerType, error) {
	switch typ {
	case networkingv1alpha1.LoadBalancerTypePublic:
		return ori.LoadBalancerType_PUBLIC, nil
	case networkingv1alpha1.LoadBalancerTypeInternal:
		return ori.LoadBalancerType_INTERNAL, nil
	default:
		return 0, fmt.Errorf("unrecognized load balancer type %q", typ)
	}
}

func (s *Server) convertOnmetalProtocol(protocol corev1.Protocol) (ori.Protocol, error) {
	switch protocol {
	case corev1.ProtocolTCP:
		return ori.Protocol_TCP, nil
	case corev1.ProtocolSCTP:
		return ori.Protocol_SCTP, nil
	case corev1.ProtocolUDP:
		return ori.Protocol_UDP, nil
	default:
		return 0, fmt.Errorf("unrecognized protocol %q", protocol)
	}
}

func (s *Server) convertOnmetalLoadBalancerTargetPort(port machinebrokerv1alpha1.LoadBalancerPort) (*ori.LoadBalancerPort, error) {
	protocol, err := s.convertOnmetalProtocol(port.Protocol)
	if err != nil {
		return nil, err
	}

	return &ori.LoadBalancerPort{
		Protocol: protocol,
		Port:     port.Port,
		EndPort:  port.EndPort,
	}, nil
}

func (s *Server) convertOnmetalLoadBalancerTargets(loadBalancerTargets []machinebrokerv1alpha1.LoadBalancerTarget) ([]*ori.LoadBalancerTargetSpec, error) {
	res := make([]*ori.LoadBalancerTargetSpec, len(loadBalancerTargets))
	for i, loadBalancerTarget := range loadBalancerTargets {
		typ, err := s.convertOnmetalLoadBalancerType(loadBalancerTarget.LoadBalancerType)
		if err != nil {
			return nil, err
		}

		ports := make([]*ori.LoadBalancerPort, len(loadBalancerTarget.Ports))
		for j, port := range loadBalancerTarget.Ports {
			p, err := s.convertOnmetalLoadBalancerTargetPort(port)
			if err != nil {
				return nil, err
			}

			ports[j] = p
		}

		res[i] = &ori.LoadBalancerTargetSpec{
			LoadBalancerType: typ,
			Ip:               loadBalancerTarget.IP.String(),
			Ports:            ports,
		}
	}
	return res, nil
}

func (s *Server) convertOnmetalNATGatewayTarget(natGatewayTarget machinebrokerv1alpha1.NATGatewayTarget) (*ori.NATSpec, error) {
	return &ori.NATSpec{
		Ip:      natGatewayTarget.IP.String(),
		Port:    natGatewayTarget.Port,
		EndPort: natGatewayTarget.EndPort,
	}, nil
}

func (s *Server) convertOnmetalNATGatewayTargets(natGatewayTargets []machinebrokerv1alpha1.NATGatewayTarget) ([]*ori.NATSpec, error) {
	res := make([]*ori.NATSpec, len(natGatewayTargets))
	for i, natGatewayTarget := range natGatewayTargets {
		nat, err := s.convertOnmetalNATGatewayTarget(natGatewayTarget)
		if err != nil {
			return nil, err
		}

		res[i] = nat
	}
	return res, nil
}

func (s *Server) getOnmetalIPsIPFamilies(ips []commonv1alpha1.IP) []corev1.IPFamily {
	res := make([]corev1.IPFamily, len(ips))
	for i, ip := range ips {
		res[i] = ip.Family()
	}
	return res
}

func (s *Server) onmetalIPsToOnmetalIPSources(ips []commonv1alpha1.IP) []networkingv1alpha1.IPSource {
	res := make([]networkingv1alpha1.IPSource, len(ips))
	for i := range ips {
		res[i] = networkingv1alpha1.IPSource{
			Value: &ips[i],
		}
	}
	return res
}

func (s *Server) onmetalPrefixesToOnmetalPrefixSources(prefixes []commonv1alpha1.IPPrefix) []networkingv1alpha1.PrefixSource {
	res := make([]networkingv1alpha1.PrefixSource, len(prefixes))
	for i := range prefixes {
		res[i] = networkingv1alpha1.PrefixSource{
			Value: &prefixes[i],
		}
	}
	return res
}

func (s *Server) parseIPs(ipStrings []string) ([]commonv1alpha1.IP, error) {
	var ips []commonv1alpha1.IP
	for _, ipString := range ipStrings {
		ip, err := commonv1alpha1.ParseIP(ipString)
		if err != nil {
			return nil, fmt.Errorf("error parsing ip %q: %w", ipString, err)
		}

		ips = append(ips, ip)
	}
	return ips, nil
}

func (s *Server) convertORIProtocol(protocol ori.Protocol) (corev1.Protocol, error) {
	switch protocol {
	case ori.Protocol_TCP:
		return corev1.ProtocolTCP, nil
	case ori.Protocol_UDP:
		return corev1.ProtocolUDP, nil
	case ori.Protocol_SCTP:
		return corev1.ProtocolSCTP, nil
	default:
		return "", fmt.Errorf("unknown protocol %d", protocol)
	}
}

func (s *Server) convertORILoadBalancerType(typ ori.LoadBalancerType) (networkingv1alpha1.LoadBalancerType, error) {
	switch typ {
	case ori.LoadBalancerType_PUBLIC:
		return networkingv1alpha1.LoadBalancerTypePublic, nil
	case ori.LoadBalancerType_INTERNAL:
		return networkingv1alpha1.LoadBalancerTypeInternal, nil
	default:
		return "", fmt.Errorf("unknown load balancer type %d", typ)
	}
}

func (s *Server) parseIPPrefixes(prefixStrings []string) ([]commonv1alpha1.IPPrefix, error) {
	var ipPrefixes []commonv1alpha1.IPPrefix
	for _, prefixString := range prefixStrings {
		ipPrefix, err := commonv1alpha1.ParseIPPrefix(prefixString)
		if err != nil {
			return nil, fmt.Errorf("error parsing ip prefix %q: %w", prefixString, err)
		}

		ipPrefixes = append(ipPrefixes, ipPrefix)
	}
	return ipPrefixes, nil
}

func RollbackTransactionIgnoreClosedFunc[E any](t transaction.Transaction[E]) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return transaction.IgnoreClosedError(t.Rollback())
	}
}
