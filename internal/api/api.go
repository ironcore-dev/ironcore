// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	computeinstall "github.com/ironcore-dev/ironcore/internal/apis/compute/install"
	coreinstall "github.com/ironcore-dev/ironcore/internal/apis/core/install"
	ipaminstall "github.com/ironcore-dev/ironcore/internal/apis/ipam/install"
	networkinginstall "github.com/ironcore-dev/ironcore/internal/apis/networking/install"
	storageinstall "github.com/ironcore-dev/ironcore/internal/apis/storage/install"
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
