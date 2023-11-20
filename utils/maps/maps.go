// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package maps

import (
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"k8s.io/apimachinery/pkg/util/sets"
)

// Pop gets the value associated with the key (if any) and deletes it from the map.
func Pop[M ~map[K]V, K comparable, V any](m M, key K) (V, bool) {
	v, ok := m[key]
	delete(m, key)
	return v, ok
}

// GetSingle returns the single key-value-pair if the map has length 1. Otherwise, it returns zero values and false.
func GetSingle[M ~map[K]V, K comparable, V any](m M) (K, V, bool) {
	if len(m) != 1 {
		var (
			zeroKey   K
			zeroValue V
		)
		return zeroKey, zeroValue, false
	}
	for k, v := range m {
		return k, v, true
	}
	panic("GetSingle: concurrent map modification: map reported a single element but was modified upon access")
}

// PopAny pops a key-value-pair from the map, if any.
// The returned boolean indicates whether a pair was present.
func PopAny[M ~map[K]V, K comparable, V any](m M) (K, V, bool) {
	for k, v := range m {
		delete(m, k)
		return k, v, true
	}
	var (
		zeroKey   K
		zeroValue V
	)
	return zeroKey, zeroValue, false
}

// SortedKeys returns the maps keys sorted.
func SortedKeys[M ~map[K]V, K interface {
	comparable
	constraints.Ordered
}, V any](m M) []K {
	keys := maps.Keys(m)
	slices.Sort(keys)
	return keys
}

// SortedKeysFunc returns the maps keys sorted by the given less func.
func SortedKeysFunc[M ~map[K]V, K comparable, V any](m M, less func(a, b K) bool) []K {
	keys := maps.Keys(m)
	slices.SortFunc(keys, less)
	return keys
}

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

// Deleted returns a map with the given keys deleted.
// If the map is empty, nil is returned.
func Deleted[M ~map[K]V, K comparable, V any](m M, keys ...K) map[K]V {
	if m == nil {
		return nil
	}
	for _, k := range keys {
		delete(m, k)
		if len(m) == 0 {
			return nil
		}
	}
	return m
}
