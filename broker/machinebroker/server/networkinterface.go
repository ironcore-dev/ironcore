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

	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	"github.com/onmetal/onmetal-api/utils/generic"
	"golang.org/x/exp/slices"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func onmetalMachineNetworkInterfaceIndex(onmetalMachine *computev1alpha1.Machine, name string) int {
	return slices.IndexFunc(
		onmetalMachine.Spec.NetworkInterfaces,
		func(nic computev1alpha1.NetworkInterface) bool {
			return nic.Name == name
		},
	)
}

func (s *Server) bindOnmetalMachineNetworkInterface(
	ctx context.Context,
	onmetalMachine *computev1alpha1.Machine,
	onmetalNetworkInterface *networkingv1alpha1.NetworkInterface,
) error {
	baseOnmetalNetworkInterface := onmetalNetworkInterface.DeepCopy()
	if err := ctrl.SetControllerReference(onmetalMachine, onmetalNetworkInterface, s.cluster.Scheme()); err != nil {
		return err
	}
	onmetalNetworkInterface.Spec.MachineRef = generic.Pointer(s.localObjectReferenceTo(onmetalMachine))
	return s.cluster.Client().Patch(ctx, onmetalNetworkInterface, client.StrategicMergeFrom(baseOnmetalNetworkInterface))
}

func (s *Server) aggregateOnmetalNetworkInterface(
	ctx context.Context,
	rd client.Reader,
	onmetalNic *networkingv1alpha1.NetworkInterface,
) (*AggregateOnmetalNetworkInterface, error) {
	onmetalNetwork := &networkingv1alpha1.Network{}
	onmetalNetworkKey := client.ObjectKey{Namespace: s.cluster.Namespace(), Name: onmetalNic.Spec.NetworkRef.Name}
	if err := rd.Get(ctx, onmetalNetworkKey, onmetalNetwork); err != nil {
		return nil, fmt.Errorf("error getting onmetal network %s: %w", onmetalNic.Name, err)
	}

	return &AggregateOnmetalNetworkInterface{
		Network:          onmetalNetwork,
		NetworkInterface: onmetalNic,
	}, nil
}
