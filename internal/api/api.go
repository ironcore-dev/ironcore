// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	computeinstall "github.com/onmetal/onmetal-api/internal/apis/compute/install"
	coreinstall "github.com/onmetal/onmetal-api/internal/apis/core/install"
	ipaminstall "github.com/onmetal/onmetal-api/internal/apis/ipam/install"
	networkinginstall "github.com/onmetal/onmetal-api/internal/apis/networking/install"
	storageinstall "github.com/onmetal/onmetal-api/internal/apis/storage/install"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
)

var (
	Scheme = runtime.NewScheme()

	Codecs = serializer.NewCodecFactory(Scheme)

	ParameterCodec = runtime.NewParameterCodec(Scheme)
)

func init() {
	ipaminstall.Install(Scheme)
	computeinstall.Install(Scheme)
	coreinstall.Install(Scheme)
	networkinginstall.Install(Scheme)
	storageinstall.Install(Scheme)

	utilruntime.Must(autoscalingv1.AddToScheme(Scheme))

	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})

	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)
}
