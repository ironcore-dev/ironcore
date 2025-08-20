// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package maps

import (
	"encoding/json"
	"fmt"
)

// Pop gets the value associated with the key (if any) and deletes it from the map.
func Pop[M ~map[K]V, K comparable, V any](m M, key K) (V, bool) {
	v, ok := m[key]
	delete(m, key)
	return v, ok
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

func Clone[M ~map[K]V, K comparable, V any](m M) M {
	clone := make(M)
	for k, v := range m {
		clone[k] = v
	}
	return clone
}

// MustMarshalJSON is a helper function that marshals the given value to JSON and panics if an error occurs.
func MustMarshalJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func UnmarshalLabels(labelsString string, present bool) (map[string]string, error) {
	if present {
		var labels map[string]string
		if err := json.Unmarshal([]byte(labelsString), &labels); err != nil {
			return nil, fmt.Errorf("error unmarshaling labels: %w", err)
		}
		return labels, nil
	}
	return map[string]string{}, nil
}
