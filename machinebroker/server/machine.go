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
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/machinebroker/apiutils"
	ori "github.com/onmetal/onmetal-api/ori/apis/compute/v1alpha1"
)

func (s *Server) convertOnmetalMachine(machine *computev1alpha1.Machine) (*ori.Machine, error) {
	id := machine.Labels[machinebrokerv1alpha1.MachineIDLabel]

	metadata, err := apiutils.GetMetadataAnnotation(machine)
	if err != nil {
		return nil, err
	}

	labels, err := apiutils.GetLabelsAnnotation(machine)
	if err != nil {
		return nil, err
	}

	annotations, err := apiutils.GetAnnotationsAnnotation(machine)
	if err != nil {
		return nil, err
	}

	return &ori.Machine{
		Id:          id,
		Metadata:    metadata,
		Annotations: annotations,
		Labels:      labels,
	}, nil
}
