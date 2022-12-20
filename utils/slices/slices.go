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

package slices

import "github.com/onmetal/controller-utils/set"

func ToMap[S ~[]V, K comparable, V any](s S, f func(v V) K) map[K]V {
	res := make(map[K]V)
	for _, v := range s {
		res[f(v)] = v
	}
	return res
}

func ToSetFunc[S ~[]V, K comparable, V any](s S, f func(v V) K) set.Set[K] {
	res := set.New[K]()
	for _, v := range s {
		res.Insert(f(v))
	}
	return res
}

func FindFunc[S ~[]V, V any](s S, f func(v V) bool) (V, bool) {
	for _, v := range s {
		if f(v) {
			return v, true
		}
	}
	var zero V
	return zero, false
}

func FindRefFunc[S ~[]V, V any](s S, f func(v V) bool) *V {
	for i, v := range s {
		if f(v) {
			return &s[i]
		}
	}
	return nil
}
