// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package core

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// ObjectSelector specifies how to select objects of a certain kind.
type ObjectSelector struct {
	// Kind is the kind of object to select.
	Kind string
	// LabelSelector is the label selector to select objects of the specified Kind by.
	metav1.LabelSelector
}
