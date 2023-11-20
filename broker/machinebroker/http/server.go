// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

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
