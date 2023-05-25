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

import (
	"k8s.io/apimachinery/pkg/util/sets"
)

func KeysDifference[M ~map[K]V, K comparable, V any](m1, m2 M) sets.Set[K] {
	result := sets.New[K]()
	for key := range m1 {
		if _, ok := m2[key]; !ok {
			result.Insert(key)
		}
	}
	return result
}

func Append[M ~map[K]V, K comparable, V any](m M, key K, value V) map[K]V {
	if m == nil {
		m = make(map[K]V)
	}
	m[key] = value
	return m
}

func AppendMap[M ~map[K]V, K comparable, V any](m M, ms ...M) map[K]V {
	for _, mi := range ms {
		if len(mi) > 0 && m == nil {
			m = make(map[K]V)
		}
		for k, v := range mi {
			m[k] = v
		}
	}
	return m
}
