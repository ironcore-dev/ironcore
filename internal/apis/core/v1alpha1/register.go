// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Package v1alpha1 contains API Schema definitions for the core v1alpha1 API group
// +groupName=core.ironcore.dev
package v1alpha1

import (
	"github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "core.ironcore.dev", Version: "v1alpha1"}

	localSchemeBuilder = &v1alpha1.SchemeBuilder

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = localSchemeBuilder.AddToScheme
)

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}
