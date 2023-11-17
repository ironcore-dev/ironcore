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

package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-logr/logr"
	"github.com/ironcore-dev/ironcore/broker/machinebroker/server"
	utilshttp "github.com/ironcore-dev/ironcore/utils/http"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("http")

type HandlerOptions struct {
	Log logr.Logger
}

func setHandlerOptionsDefaults(opts *HandlerOptions) {
	if opts.Log.GetSink() == nil {
		opts.Log = log.WithName("server")
	}
}

func NewHandler(srv *server.Server, opts HandlerOptions) http.Handler {
	setHandlerOptionsDefaults(&opts)

	r := chi.NewRouter()

	r.Use(utilshttp.InjectLogger(opts.Log))
	r.Use(utilshttp.LogRequest)

	for _, method := range []string{http.MethodHead, http.MethodGet, http.MethodPost} {
		r.MethodFunc(method, "/exec/{token}", func(w http.ResponseWriter, req *http.Request) {
			token := chi.URLParam(req, "token")
			srv.ServeExec(w, req, token)
		})
	}

	return r
}
