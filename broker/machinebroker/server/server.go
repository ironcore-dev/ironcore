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
	"net/url"
	"path"

	"github.com/onmetal/onmetal-api/broker/common/request"
	"github.com/onmetal/onmetal-api/broker/machinebroker/cluster"
	"github.com/onmetal/onmetal-api/broker/machinebroker/networks"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"k8s.io/client-go/rest"
)

var _ ori.MachineRuntimeServer = (*Server)(nil)

//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=compute.api.onmetal.de,resources=machines/exec,verbs=get;create
//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=storage.api.onmetal.de,resources=volumes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networkinterfaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=networks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualips,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=virtualips/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=loadbalancers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=loadbalancers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=loadbalancerroutings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=natgateways,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.api.onmetal.de,resources=natgateways/status,verbs=get;update;patch

type BrokerLabel struct {
	DefaultLabel     string
	DownwardAPILabel string
}

type Server struct {
	baseURL *url.URL

	brokerDownwardAPILabels map[string]string

	cluster cluster.Cluster

	networks *networks.Manager

	execRequestCache request.Cache[*ori.ExecRequest]
}

type Options struct {
	// BaseURL is the base URL in form http(s)://host:port/path?query to produce request URLs relative to.
	BaseURL string
	// BrokerDownwardAPILabels specifies which labels to broker via downward API and what the default
	// label name is to obtain the value in case there is no value for the downward API.
	// Example usage is e.g. to broker the root UID (map "root-machine-uid" to machinepoollet's
	// "machinepoollet.api.onmetal.de/machine-uid")
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
		execRequestCache:        request.NewCache[*ori.ExecRequest](),
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
