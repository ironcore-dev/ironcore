// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package sets

import (
	"golang.org/x/exp/slices"
	"k8s.io/apimachinery/pkg/util/sets"
)

// ListFunc returns an ordered slice of the given set by applying the given less func as comparator.
func ListFunc[T comparable](set sets.Set[T], less func(v1, v2 T) bool) []T {
	items := set.UnsortedList()
	slices.SortFunc(items, less)
	return items
}
