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

package maps

import "github.com/onmetal/controller-utils/set"

func KeySet[M ~map[K]V, K comparable, V any](m M) set.Set[K] {
	s := make(set.Set[K], len(m))
	for k := range m {
		s.Insert(k)
	}
	return s
}

func KeysDifference[M ~map[K]V, K comparable, V any](m1, m2 M) set.Set[K] {
	result := set.New[K]()
	for key := range m1 {
		if _, ok := m2[key]; !ok {
			result.Insert(key)
		}
	}
	return result
}
