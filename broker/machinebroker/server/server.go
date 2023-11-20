// Copyright 2022 IronCore authors
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
	"net/url"
	"path"

	"github.com/ironcore-dev/ironcore/broker/common/request"
	"github.com/ironcore-dev/ironcore/broker/machinebroker/cluster"
	"github.com/ironcore-dev/ironcore/broker/machinebroker/networks"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"k8s.io/client-go/rest"
)

var _ iri.MachineRuntimeServer = (*Server)(nil)

//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=compute.ironcore.dev,resources=machines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=compute.ironcore.dev,resources=machines/exec,verbs=get;create
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networkinterfaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=networks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=virtualips,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=virtualips/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=loadbalancers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=loadbalancers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=loadbalancerroutings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=natgateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.ironcore.dev,resources=natgateways/status,verbs=get;update;patch

type BrokerLabel struct {
	DefaultLabel     string
	DownwardAPILabel string
}

type Server struct {
	baseURL *url.URL

	brokerDownwardAPILabels map[string]string

	cluster cluster.Cluster

	networks *networks.Manager

	execRequestCache request.Cache[*iri.ExecRequest]
}

type Options struct {
	// BaseURL is the base URL in form http(s)://host:port/path?query to produce request URLs relative to.
	BaseURL string
	// BrokerDownwardAPILabels specifies which labels to broker via downward API and what the default
	// label name is to obtain the value in case there is no value for the downward API.
	// Example usage is e.g. to broker the root UID (map "root-machine-uid" to machinepoollet's
	// "machinepoollet.ironcore.dev/machine-uid")
	BrokerDownwardAPILabels map[string]string
	MachinePoolName         string
	MachinePoolSelector     map[string]string
}

func New(cfg *rest.Config, namespace string, opts Options) (*Server, error) {
	baseURL, err := url.ParseRequestURI(opts.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base url %q: %w", opts.BaseURL, err)
	}

	c, err := cluster.New(cfg, namespace, cluster.Options{
		MachinePoolName:     opts.MachinePoolName,
		MachinePoolSelector: opts.MachinePoolSelector,
	})
	if err != nil {
		return nil, err
	}

	return &Server{
		baseURL:                 baseURL,
		brokerDownwardAPILabels: opts.BrokerDownwardAPILabels,
		cluster:                 c,
		networks:                networks.NewManager(c),
		execRequestCache:        request.NewCache[*iri.ExecRequest](),
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	return s.networks.Start(ctx)
}

func (s *Server) buildURL(method string, token string) string {
	return s.baseURL.ResolveReference(&url.URL{
		Path: path.Join(method, token),
	}).String()
}
