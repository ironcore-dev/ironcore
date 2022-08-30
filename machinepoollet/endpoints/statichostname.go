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
	"fmt"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
)

type StaticHostName struct {
	hostname string
	port     int32
}

func (s *StaticHostName) AddListener(listener Listener) {}

func (s *StaticHostName) GetEndpoints() (addresses []computev1alpha1.MachinePoolAddress, port int32) {
	return []computev1alpha1.MachinePoolAddress{{
		Type:    computev1alpha1.MachinePoolHostName,
		Address: s.hostname,
	}}, s.port
}

func NewStaticHostName(hostname string, port int32) (*StaticHostName, error) {
	if hostname == "" {
		return nil, fmt.Errorf("must specify hostname")
	}
	if port == 0 {
		return nil, fmt.Errorf("must specify port")
	}

	return &StaticHostName{
		hostname: hostname,
		port:     port,
	}, nil
}
