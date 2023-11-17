// Copyright 2023 IronCore authors
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
	"net/http"
	"net/url"
	"strconv"

	ori "github.com/ironcore-dev/ironcore/ori/apis/machine/v1alpha1"
	machinepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/proxy"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (s *Server) serveExec(w http.ResponseWriter, req *http.Request, namespace, name string) {
	ctx := req.Context()
	log := ctrl.LoggerFrom(ctx)

	listMachinesRes, err := s.machineRuntime.ListMachines(ctx, &ori.ListMachinesRequest{
		Filter: &ori.MachineFilter{
			LabelSelector: map[string]string{
				machinepoolletv1alpha1.MachineNamespaceLabel: namespace,
				machinepoolletv1alpha1.MachineNameLabel:      name,
			},
		},
	})
	if err != nil {
		log.Error(err, "Error listing machines")
		s.writeError(w, err)
		return
	}
	if len(listMachinesRes.Machines) == 0 {
		http.Error(w, "machine not found", http.StatusNotFound)
		return
	}

	machine := listMachinesRes.Machines[0]
	execRes, err := s.machineRuntime.Exec(ctx, &ori.ExecRequest{
		MachineId: machine.Metadata.Id,
	})
	if err != nil {
		log.Error(err, "Error getting exec url")
		s.writeError(w, err)
		return
	}

	execURL, err := url.Parse(execRes.Url)
	if err != nil {
		log.Error(err, "Error parsing exec url")
		s.writeError(w, err)
		return
	}

	proxyStream(w, req, execURL)
}

func (s *Server) writeError(w http.ResponseWriter, err error) {
	status, _ := grpcstatus.FromError(err)
	var code int
	switch status.Code() {
	case codes.NotFound:
		code = http.StatusNotFound
	case codes.ResourceExhausted:
		w.Header().Set("Retry-After", strconv.Itoa(int(s.cacheTTL.Seconds())))
	default:
		code = http.StatusInternalServerError
	}
	w.WriteHeader(code)
	_, _ = w.Write([]byte(err.Error()))
}

func proxyStream(w http.ResponseWriter, req *http.Request, url *url.URL) {
	handler := proxy.NewUpgradeAwareHandler(url, nil, false, true, &responder{})
	handler.ServeHTTP(w, req)
}

type responder struct{}

func (r *responder) Error(w http.ResponseWriter, req *http.Request, err error) {
	ctx := req.Context()
	log := ctrl.LoggerFrom(ctx)
	log.Error(err, "Error while proxying request")
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
