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

package server

import (
	"fmt"

	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	"github.com/onmetal/onmetal-api/ori/apis/compute/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubernetes "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(kubernetes.AddToScheme(scheme))
	utilruntime.Must(computev1alpha1.AddToScheme(scheme))
	utilruntime.Must(networkingv1alpha1.AddToScheme(scheme))
	utilruntime.Must(storagev1alpha1.AddToScheme(scheme))
	utilruntime.Must(ipamv1alpha1.AddToScheme(scheme))
}

type Server struct {
	client client.Client

	namespace           string
	machinePoolName     string
	machinePoolSelector map[string]string
}

var _ v1alpha1.MachineRuntimeServer = (*Server)(nil)

type Options struct {
	Namespace           string
	MachinePoolName     string
	MachinePoolSelector map[string]string
}

func setOptionsDefaults(o *Options) {
	if o.Namespace == "" {
		o.Namespace = corev1.NamespaceDefault
	}
}

func New(cfg *rest.Config, opts Options) (*Server, error) {
	setOptionsDefaults(&opts)

	c, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	return &Server{
		client:              c,
		namespace:           opts.Namespace,
		machinePoolName:     opts.MachinePoolName,
		machinePoolSelector: opts.MachinePoolSelector,
	}, nil
}
