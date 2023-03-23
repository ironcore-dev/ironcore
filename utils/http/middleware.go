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

	"github.com/go-logr/logr"
)

// Middleware wraps an http.Handler, returning a new http.Handler.
type Middleware = func(next http.Handler) http.Handler

// InjectLogger is a Middleware to inject the given logr.Logger into the http.Request.Context.
func InjectLogger(log logr.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			ctx = logr.NewContext(ctx, log)
			req = req.WithContext(ctx)
			next.ServeHTTP(w, req)
		})
	}
}
