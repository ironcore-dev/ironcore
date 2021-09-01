/*
 * Copyright (c) 2021 by the OnMetal authors.
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

// Package v1alpha1 is the v1alpha1 version of the API.
// +k8s:deepcopy-gen=package,register
// +k8s:openapi-gen=true
// +k8s:defaulter-gen=TypeMeta
// +k8s:protobuf-gen=package

//go:generate go run github.com/ahmetb/gen-crd-api-reference-docs -api-dir . -config ../../../hack/api-reference/compute-config.json -template-dir ../../../hack/api-reference/template -out-file ../../../docs/api-reference/compute.md

// Package v1beta1 is a version of the API.
// +groupName=compute.onmetal.de

package v1alpha1
