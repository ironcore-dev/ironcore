// Copyright 2023 OnMetal authors
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
	"net/http"
	"net/url"

	"github.com/go-logr/logr"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	onmetalapiclientgoscheme "github.com/onmetal/onmetal-api/client-go/onmetalapi/scheme"
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/client-go/rest"
)

func (s *Server) Exec(ctx context.Context, req *ori.ExecRequest) (*ori.ExecResponse, error) {
	machineID := req.MachineId
	log := s.loggerFrom(ctx, "MachineID", machineID)

	log.V(1).Info("Inserting request into cache")
	token, err := s.execRequestCache.Insert(req)
	if err != nil {
		return nil, err
	}

	log.V(1).Info("Returning url with token")
	return &ori.ExecResponse{
		Url: s.buildURL("exec", token),
	}, nil
}

func (s *Server) ServeExec(w http.ResponseWriter, req *http.Request, token string) {
	ctx := req.Context()
	log := logr.FromContextOrDiscard(ctx)

	request, ok := s.execRequestCache.Consume(token)
	if !ok {
		log.V(1).Info("Rejecting unknown / expired token")
		http.NotFound(w, req)
		return
	}

	cfgShallowCopy := s.cluster.Config()
	cfgShallowCopy.GroupVersion = &computev1alpha1.SchemeGroupVersion
	cfgShallowCopy.APIPath = "/apis"
	cfgShallowCopy.NegotiatedSerializer = onmetalapiclientgoscheme.Codecs.WithoutConversion()
	if cfgShallowCopy.UserAgent == "" {
		cfgShallowCopy.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	restClient, err := rest.RESTClientFor(cfgShallowCopy)
	if err != nil {
		log.Error(err, "Error getting rest client for cluster")
		code := http.StatusInternalServerError
		http.Error(w, http.StatusText(code), code)
		return
	}

	targetURL := restClient.
		Verb(req.Method).
		Namespace(s.cluster.Namespace()).
		Resource("machines").
		Name(request.MachineId).
		SubResource("exec").
		VersionedParams(&computev1alpha1.MachineExecOptions{}, onmetalapiclientgoscheme.ParameterCodec).
		URL()

	proxyStream(w, req, targetURL, restClient.Client.Transport)
}

func proxyStream(w http.ResponseWriter, req *http.Request, url *url.URL, transport http.RoundTripper) {
	handler := proxy.NewUpgradeAwareHandler(url, transport, false, false, &responder{})
	handler.ServeHTTP(w, req)
}

type responder struct{}

func (r *responder) Error(w http.ResponseWriter, req *http.Request, err error) {
	log := logr.FromContextOrDiscard(req.Context())
	log.Error(err, "Error while proxying request")
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
