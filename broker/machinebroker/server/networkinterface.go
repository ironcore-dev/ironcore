// Copyright 2022 IronCore authors
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
