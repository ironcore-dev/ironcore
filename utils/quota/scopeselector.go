// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package quota

import (
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
)

func GetResourceScopeSelectorRequirements(scopeSelector *corev1alpha1.ResourceScopeSelector) []corev1alpha1.ResourceScopeSelectorRequirement {
	if scopeSelector == nil {
		return nil
	}

	return scopeSelector.MatchExpressions
}
