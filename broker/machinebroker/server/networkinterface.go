// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	"github.com/ironcore-dev/ironcore/utils/generic"
	"golang.org/x/exp/slices"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ironcoreMachineNetworkInterfaceIndex(ironcoreMachine *computev1alpha1.Machine, name string) int {
	return slices.IndexFunc(
		ironcoreMachine.Spec.NetworkInterfaces,
		func(nic computev1alpha1.NetworkInterface) bool {
			return nic.Name == name
		},
	)
}

func (s *Server) bindIronCoreMachineNetworkInterface(
	ctx context.Context,
	ironcoreMachine *computev1alpha1.Machine,
	ironcoreNetworkInterface *networkingv1alpha1.NetworkInterface,
) error {
	baseIronCoreNetworkInterface := ironcoreNetworkInterface.DeepCopy()
	if err := ctrl.SetControllerReference(ironcoreMachine, ironcoreNetworkInterface, s.cluster.Scheme()); err != nil {
		return err
	}
	ironcoreNetworkInterface.Spec.MachineRef = generic.Pointer(s.localObjectReferenceTo(ironcoreMachine))
	return s.cluster.Client().Patch(ctx, ironcoreNetworkInterface, client.StrategicMergeFrom(baseIronCoreNetworkInterface))
}

func (s *Server) aggregateIronCoreNetworkInterface(
	ctx context.Context,
	rd client.Reader,
	ironcoreNic *networkingv1alpha1.NetworkInterface,
) (*AggregateIronCoreNetworkInterface, error) {
	ironcoreNetwork := &networkingv1alpha1.Network{}
	ironcoreNetworkKey := client.ObjectKey{Namespace: s.cluster.Namespace(), Name: ironcoreNic.Spec.NetworkRef.Name}
	if err := rd.Get(ctx, ironcoreNetworkKey, ironcoreNetwork); err != nil {
		return nil, fmt.Errorf("error getting ironcore network %s: %w", ironcoreNic.Name, err)
	}

	return &AggregateIronCoreNetworkInterface{
		Network:          ironcoreNetwork,
		NetworkInterface: ironcoreNic,
	}, nil
}
