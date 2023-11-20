// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

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
