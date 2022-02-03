/*
 * Copyright (c) 2022 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tolerations

import (
	apiequality "k8s.io/apimachinery/pkg/api/equality"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
)

// Based on https://pkg.go.dev/k8s.io/kubernetes/pkg/util/tolerations#MergeTolerations with our own Tolerations
// MergeTolerations merges two sets of tolerations into one. If one toleration is a superset of
// another, only the superset is kept.
func MergeTolerations(first, second []commonv1alpha1.Toleration) []commonv1alpha1.Toleration {
	all := append(first, second...)
	var merged []commonv1alpha1.Toleration

Next:
	for i, t := range all {
		for _, t2 := range merged {
			if isSuperset(&t2, &t) {
				continue Next // t is redundant; ignore it
			}
		}
		if i+1 < len(all) {
			for _, t2 := range all[i+1:] {
				if !apiequality.Semantic.DeepEqual(&t2, &t) && isSuperset(&t2, &t) { // if there's a superset later in `all`
					continue Next // t is redundant; ignore it
				}
			}
		}
		merged = append(merged, t) // If some tolerations are equal, the first occurrence will be appended.
	}

	return merged
}

// isSuperset checks whether super tolerates a superset of sub.
func isSuperset(super, sub *commonv1alpha1.Toleration) bool {
	if super.Key != sub.Key &&
		// An empty key with Exists operator means matching all keys & values.
		!(super.Key == "" && super.Operator == commonv1alpha1.TolerationOpExists) {
		return false
	}

	// An empty effect means matching all effects.
	if super.Effect != "" && sub.Effect != super.Effect {
		return false
	}

	switch super.Operator {
	case commonv1alpha1.TolerationOpEqual, "": // empty operator means Equal
		return (sub.Operator == commonv1alpha1.TolerationOpEqual || sub.Operator == "") &&
			super.Value == sub.Value
	case commonv1alpha1.TolerationOpExists:
		return true
	default: // false operator
		return false
	}
}
