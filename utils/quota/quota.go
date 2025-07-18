// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package quota

import (
	"strings"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Equals returns true if the two lists are equivalent
func Equals(a corev1alpha1.ResourceList, b corev1alpha1.ResourceList) bool {
	if len(a) != len(b) {
		return false
	}

	for key, value1 := range a {
		value2, found := b[key]
		if !found {
			return false
		}
		if value1.Cmp(value2) != 0 {
			return false
		}
	}

	return true
}

// Contains returns true if first list contains all of second list
func Contains(a corev1alpha1.ResourceList, b corev1alpha1.ResourceList) bool {
	for key, value1 := range b {
		value2, found := a[key]
		if !found {
			return false
		}
		if value1.Cmp(value2) != 0 {
			return false
		}
	}

	return true
}

// LessThanOrEqual returns true if a < b for each key in b
// If false, it returns the keys in a that exceeded b
func LessThanOrEqual(a corev1alpha1.ResourceList, b corev1alpha1.ResourceList) (bool, sets.Set[corev1alpha1.ResourceName]) {
	result := true
	resourceNames := sets.New[corev1alpha1.ResourceName]()
	for key, value := range b {
		if other, found := a[key]; found {
			if other.Cmp(value) > 0 {
				result = false
				resourceNames.Insert(key)
			}
		}
	}
	return result, resourceNames
}

// Max returns the result of Max(a, b) for each named resource
func Max(a corev1alpha1.ResourceList, b corev1alpha1.ResourceList) corev1alpha1.ResourceList {
	result := corev1alpha1.ResourceList{}
	for key, value := range a {
		if other, found := b[key]; found {
			if value.Cmp(other) <= 0 {
				result[key] = other.DeepCopy()
				continue
			}
		}
		result[key] = value.DeepCopy()
	}
	for key, value := range b {
		if _, found := result[key]; !found {
			result[key] = value.DeepCopy()
		}
	}
	return result
}

// Add returns the result of a + b for each named resource
func Add(a corev1alpha1.ResourceList, b corev1alpha1.ResourceList) corev1alpha1.ResourceList {
	result := corev1alpha1.ResourceList{}
	for key, value := range a {
		quantity := value.DeepCopy()
		if other, found := b[key]; found {
			quantity.Add(other)
		}
		result[key] = quantity
	}
	for key, value := range b {
		if _, found := result[key]; !found {
			result[key] = value.DeepCopy()
		}
	}
	return result
}

// SubtractWithNonNegativeResult - subtracts and returns result of a - b but
// makes sure we don't return negative values to prevent negative resource usage.
func SubtractWithNonNegativeResult(a corev1alpha1.ResourceList, b corev1alpha1.ResourceList) corev1alpha1.ResourceList {
	zero := resource.MustParse("0")

	result := corev1alpha1.ResourceList{}
	for key, value := range a {
		quantity := value.DeepCopy()
		if other, found := b[key]; found {
			quantity.Sub(other)
		}
		if quantity.Cmp(zero) > 0 {
			result[key] = quantity
		} else {
			result[key] = zero
		}
	}

	for key := range b {
		if _, found := result[key]; !found {
			result[key] = zero
		}
	}
	return result
}

// Subtract returns the result of a - b for each named resource
func Subtract(a corev1alpha1.ResourceList, b corev1alpha1.ResourceList) corev1alpha1.ResourceList {
	result := corev1alpha1.ResourceList{}
	for key, value := range a {
		quantity := value.DeepCopy()
		if other, found := b[key]; found {
			quantity.Sub(other)
		}
		result[key] = quantity
	}
	for key, value := range b {
		if _, found := result[key]; !found {
			quantity := value.DeepCopy()
			quantity.Neg()
			result[key] = quantity
		}
	}
	return result
}

// Mask returns a new resource list that only has the values with the specified names
func Mask(resources corev1alpha1.ResourceList, names sets.Set[corev1alpha1.ResourceName]) corev1alpha1.ResourceList {
	result := corev1alpha1.ResourceList{}
	for key, value := range resources {
		if names.Has(key) {
			result[key] = value.DeepCopy()
		}
	}
	return result
}

// ResourceNames returns a list of all resource names in the ResourceList
func ResourceNames(resources corev1alpha1.ResourceList) sets.Set[corev1alpha1.ResourceName] {
	return sets.KeySet(resources)
}

// ContainsPrefix returns true if the specified item has a prefix that contained in given prefix Set
func ContainsPrefix(prefixSet []string, item corev1alpha1.ResourceName) bool {
	for _, prefix := range prefixSet {
		if strings.HasPrefix(string(item), prefix) {
			return true
		}
	}
	return false
}

// IsZero returns true if each key maps to the quantity value 0
func IsZero(a corev1alpha1.ResourceList) bool {
	zero := resource.MustParse("0")
	for _, v := range a {
		if v.Cmp(zero) != 0 {
			return false
		}
	}
	return true
}

// RemoveZeros returns a new resource list that only has no zero values
func RemoveZeros(a corev1alpha1.ResourceList) corev1alpha1.ResourceList {
	result := corev1alpha1.ResourceList{}
	for key, value := range a {
		if !value.IsZero() {
			result[key] = value
		}
	}
	return result
}

// IsNegative returns the set of resource names that have a negative value.
func IsNegative(a corev1alpha1.ResourceList) sets.Set[corev1alpha1.ResourceName] {
	results := sets.New[corev1alpha1.ResourceName]()
	zero := resource.MustParse("0")
	for k, v := range a {
		if v.Cmp(zero) < 0 {
			results.Insert(k)
		}
	}
	return results
}

// ToSet takes a list of resource names and converts to a string set
func ToSet(resourceNames []corev1alpha1.ResourceName) sets.Set[corev1alpha1.ResourceName] {
	return sets.New(resourceNames...)
}

func EvaluatorMatchingResourceNames(evaluator Evaluator, names sets.Set[corev1alpha1.ResourceName]) sets.Set[corev1alpha1.ResourceName] {
	res := sets.New[corev1alpha1.ResourceName]()
	for name := range names {
		if evaluator.MatchesResourceName(name) {
			res.Insert(name)
		}
	}
	return res
}

func EvaluatorMatchesResourceNames(evaluator Evaluator, names sets.Set[corev1alpha1.ResourceName]) bool {
	for name := range names {
		if evaluator.MatchesResourceName(name) {
			return true
		}
	}
	return false
}

func EvaluatorMatchesResourceList(evaluator Evaluator, list corev1alpha1.ResourceList) bool {
	for resourceName := range list {
		if evaluator.MatchesResourceName(resourceName) {
			return true
		}
	}
	return false
}

func EvaluatorMatchesResourceScopeSelector(
	evaluator Evaluator,
	item client.Object,
	resourceScopeSelector *corev1alpha1.ResourceScopeSelector,
) (bool, error) {
	if resourceScopeSelector == nil {
		return true, nil
	}
	for _, req := range resourceScopeSelector.MatchExpressions {
		ok, err := evaluator.MatchesResourceScopeSelectorRequirement(item, req)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func KubernetesResourceListToResourceList(k8sResourceList corev1.ResourceList) corev1alpha1.ResourceList {
	res := make(corev1alpha1.ResourceList, len(k8sResourceList))
	for name, quantity := range k8sResourceList {
		res[corev1alpha1.ResourceName(name)] = quantity
	}
	return res
}
