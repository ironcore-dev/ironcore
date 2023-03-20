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

package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/onmetal/onmetal-api/broker/machinebroker/server"
	utilshttp "github.com/onmetal/onmetal-api/utils/http"
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

	r := gin.New()

	r.Match(
		[]string{http.MethodHead, http.MethodGet, http.MethodPost},
		"/exec/:token",
		func(c *gin.Context) {
			srv.ServeExec(c.Writer, c.Request, c.Param("token"))
		},
	)

	return utilshttp.UseMiddleware(r,
		utilshttp.InjectLogger(opts.Log),
	)
}
