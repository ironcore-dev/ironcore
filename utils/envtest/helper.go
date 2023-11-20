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

package envtest

import (
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

var log = controllerruntime.Log.WithName("test-env")

// mergePaths merges two string slices containing paths.
// This function makes no guarantees about order of the merged slice.
func mergePaths(s1, s2 []string) []string {
	m := make(map[string]struct{})
	for _, s := range s1 {
		m[s] = struct{}{}
	}
	for _, s := range s2 {
		m[s] = struct{}{}
	}
	merged := make([]string, len(m))
	i := 0
	for key := range m {
		merged[i] = key
		i++
	}
	return merged
}

// mergeAPIServices merges two APIService slices using their names.
// This function makes no guarantees about order of the merged slice.
func mergeAPIServices(s1, s2 []*apiregistrationv1.APIService) []*apiregistrationv1.APIService {
	m := make(map[string]*apiregistrationv1.APIService)
	for _, obj := range s1 {
		m[obj.GetName()] = obj
	}
	for _, obj := range s2 {
		m[obj.GetName()] = obj
	}
	merged := make([]*apiregistrationv1.APIService, len(m))
	i := 0
	for _, obj := range m {
		merged[i] = obj.DeepCopy()
		i++
	}
	return merged
}
