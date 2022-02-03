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

// package taints is from https://pkg.go.dev/k8s.io/kubernetes/pkg/util/taints with our own Taint type and without any function relating to corev1.Node
package taints

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"

	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
)

// parseTaint parses a taint from a string, whose form must be either
// '<key>=<value>:<effect>', '<key>:<effect>', or '<key>'.
func parseTaint(st string) (commonv1alpha1.Taint, error) {
	var taint commonv1alpha1.Taint

	var key string
	var value string
	var effect commonv1alpha1.TaintEffect

	parts := strings.Split(st, ":")
	switch len(parts) {
	case 1:
		key = parts[0]
	case 2:
		effect = commonv1alpha1.TaintEffect(parts[1])
		if err := validateTaintEffect(effect); err != nil {
			return taint, err
		}

		partsKV := strings.Split(parts[0], "=")
		if len(partsKV) > 2 {
			return taint, fmt.Errorf("invalid taint spec: %v", st)
		}
		key = partsKV[0]
		if len(partsKV) == 2 {
			value = partsKV[1]
			if errs := validation.IsValidLabelValue(value); len(errs) > 0 {
				return taint, fmt.Errorf("invalid taint spec: %v, %s", st, strings.Join(errs, "; "))
			}
		}
	default:
		return taint, fmt.Errorf("invalid taint spec: %v", st)
	}

	if errs := validation.IsQualifiedName(key); len(errs) > 0 {
		return taint, fmt.Errorf("invalid taint spec: %v, %s", st, strings.Join(errs, "; "))
	}

	taint.Key = key
	taint.Value = value
	taint.Effect = effect

	return taint, nil
}

func validateTaintEffect(effect commonv1alpha1.TaintEffect) error {
	if effect != commonv1alpha1.TaintEffectNoSchedule { // the only valid effect right now, different from k8s
		return fmt.Errorf("invalid taint effect: %v, unsupported taint effect", effect)
	}

	return nil
}

// ParseTaints takes a spec which is an array and creates slices for new taints to be added, taints to be deleted.
// It also validates the spec. For example, the form `<key>` may be used to remove a taint, but not to add one.
func ParseTaints(spec []string) ([]commonv1alpha1.Taint, []commonv1alpha1.Taint, error) {
	var taints, taintsToRemove []commonv1alpha1.Taint
	uniqueTaints := map[commonv1alpha1.TaintEffect]sets.String{}

	for _, taintSpec := range spec {
		if strings.HasSuffix(taintSpec, "-") {
			taintToRemove, err := parseTaint(strings.TrimSuffix(taintSpec, "-"))
			if err != nil {
				return nil, nil, err
			}
			taintsToRemove = append(taintsToRemove, commonv1alpha1.Taint{Key: taintToRemove.Key, Effect: taintToRemove.Effect})
		} else {
			newTaint, err := parseTaint(taintSpec)
			if err != nil {
				return nil, nil, err
			}
			// validate that the taint has an effect, which is required to add the taint
			if len(newTaint.Effect) == 0 {
				return nil, nil, fmt.Errorf("invalid taint spec: %v", taintSpec)
			}
			// validate if taint is unique by <key, effect>
			if len(uniqueTaints[newTaint.Effect]) > 0 && uniqueTaints[newTaint.Effect].Has(newTaint.Key) {
				return nil, nil, fmt.Errorf("duplicated taints with the same key and effect: %v", newTaint)
			}
			// add taint to existingTaints for uniqueness check
			if len(uniqueTaints[newTaint.Effect]) == 0 {
				uniqueTaints[newTaint.Effect] = sets.String{}
			}
			uniqueTaints[newTaint.Effect].Insert(newTaint.Key)

			taints = append(taints, newTaint)
		}
	}
	return taints, taintsToRemove, nil
}

// deleteTaints deletes the given taints from the node's taintlist.
func deleteTaints(taintsToRemove []commonv1alpha1.Taint, newTaints *[]commonv1alpha1.Taint) ([]error, bool) {
	allErrs := []error{}
	var removed bool
	for _, taintToRemove := range taintsToRemove {
		removed = false // nolint:ineffassign
		if len(taintToRemove.Effect) > 0 {
			*newTaints, removed = DeleteTaint(*newTaints, &taintToRemove)
		} else {
			*newTaints, removed = DeleteTaintsByKey(*newTaints, taintToRemove.Key)
		}
		if !removed {
			allErrs = append(allErrs, fmt.Errorf("taint %q not found", taintToRemove.ToString()))
		}
	}
	return allErrs, removed
}

// addTaints adds the newTaints list to existing ones and updates the newTaints List.
// TODO: This needs a rewrite to take only the new values instead of appended newTaints list to be consistent.
func addTaints(oldTaints []commonv1alpha1.Taint, newTaints *[]commonv1alpha1.Taint) bool {
	for _, oldTaint := range oldTaints {
		existsInNew := false
		for _, taint := range *newTaints {
			if taint.MatchTaint(&oldTaint) {
				existsInNew = true
				break
			}
		}
		if !existsInNew {
			*newTaints = append(*newTaints, oldTaint)
		}
	}
	return len(oldTaints) != len(*newTaints)
}

// CheckIfTaintsAlreadyExists checks if the node already has taints that we want to add and returns a string with taint keys.
func CheckIfTaintsAlreadyExists(oldTaints []commonv1alpha1.Taint, taints []commonv1alpha1.Taint) string {
	var existingTaintList = make([]string, 0)
	for _, taint := range taints {
		for _, oldTaint := range oldTaints {
			if taint.Key == oldTaint.Key && taint.Effect == oldTaint.Effect {
				existingTaintList = append(existingTaintList, taint.Key)
			}
		}
	}
	return strings.Join(existingTaintList, ",")
}

// DeleteTaintsByKey removes all the taints that have the same key to given taintKey
func DeleteTaintsByKey(taints []commonv1alpha1.Taint, taintKey string) ([]commonv1alpha1.Taint, bool) {
	newTaints := []commonv1alpha1.Taint{}
	deleted := false
	for i := range taints {
		if taintKey == taints[i].Key {
			deleted = true
			continue
		}
		newTaints = append(newTaints, taints[i])
	}
	return newTaints, deleted
}

// DeleteTaint removes all the taints that have the same key and effect to given taintToDelete.
func DeleteTaint(taints []commonv1alpha1.Taint, taintToDelete *commonv1alpha1.Taint) ([]commonv1alpha1.Taint, bool) {
	newTaints := []commonv1alpha1.Taint{}
	deleted := false
	for i := range taints {
		if taintToDelete.MatchTaint(&taints[i]) {
			deleted = true
			continue
		}
		newTaints = append(newTaints, taints[i])
	}
	return newTaints, deleted
}

// TaintExists checks if the given taint exists in list of taints. Returns true if exists false otherwise.
func TaintExists(taints []commonv1alpha1.Taint, taintToFind *commonv1alpha1.Taint) bool {
	for _, taint := range taints {
		if taint.MatchTaint(taintToFind) {
			return true
		}
	}
	return false
}

func TaintSetDiff(t1, t2 []commonv1alpha1.Taint) (taintsToAdd []*commonv1alpha1.Taint, taintsToRemove []*commonv1alpha1.Taint) {
	for _, taint := range t1 {
		if !TaintExists(t2, &taint) {
			t := taint
			taintsToAdd = append(taintsToAdd, &t)
		}
	}

	for _, taint := range t2 {
		if !TaintExists(t1, &taint) {
			t := taint
			taintsToRemove = append(taintsToRemove, &t)
		}
	}

	return
}

func TaintSetFilter(taints []commonv1alpha1.Taint, fn func(*commonv1alpha1.Taint) bool) []commonv1alpha1.Taint {
	res := []commonv1alpha1.Taint{}

	for _, taint := range taints {
		if fn(&taint) {
			res = append(res, taint)
		}
	}

	return res
}

// CheckTaintValidation checks if the given taint is valid.
// Returns error if the given taint is invalid.
func CheckTaintValidation(taint commonv1alpha1.Taint) error {
	if errs := validation.IsQualifiedName(taint.Key); len(errs) > 0 {
		return fmt.Errorf("invalid taint key: %s", strings.Join(errs, "; "))
	}
	if taint.Value != "" {
		if errs := validation.IsValidLabelValue(taint.Value); len(errs) > 0 {
			return fmt.Errorf("invalid taint value: %s", strings.Join(errs, "; "))
		}
	}
	if taint.Effect != "" {
		if err := validateTaintEffect(taint.Effect); err != nil {
			return err
		}
	}

	return nil
}
