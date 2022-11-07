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
	"context"
	"fmt"

	"github.com/go-logr/logr"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/apis/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/apis/storage/v1alpha1"
	"github.com/onmetal/onmetal-api/ori/apis/runtime/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubernetes "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
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
	logger logr.Logger
	client client.Client

	namespace           string
	volumePoolName      string
	machinePoolName     string
	machinePoolSelector map[string]string
}

var _ v1alpha1.MachineRuntimeServer = (*Server)(nil)

type Options struct {
	Logger logr.Logger

	Namespace           string
	VolumePoolName      string
	MachinePoolName     string
	MachinePoolSelector map[string]string
}

func setOptionsDefaults(o *Options) {
	if o.Logger.GetSink() == nil {
		o.Logger = ctrl.Log
	}
	if o.Namespace == "" {
		o.Namespace = corev1.NamespaceDefault
	}
}

func New(cfg *rest.Config, opts Options) (*Server, error) {
	setOptionsDefaults(&opts)

	if opts.MachinePoolName == "" {
		return nil, fmt.Errorf("must specify MachinePoolName")
	}
	if opts.VolumePoolName == "" {
		return nil, fmt.Errorf("must specify VolumePoolName")
	}

	c, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	return &Server{
		logger:              opts.Logger,
		client:              c,
		namespace:           opts.Namespace,
		volumePoolName:      opts.VolumePoolName,
		machinePoolName:     opts.MachinePoolName,
		machinePoolSelector: opts.MachinePoolSelector,
	}, nil
}

func (s *Server) loggerFrom(ctx context.Context, keysWithValues ...interface{}) logr.Logger {
	return s.logger.WithValues(keysWithValues...)
}
