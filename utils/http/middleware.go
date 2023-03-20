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

// Middleware allows producing an http.Handler by applying arbitrary surrounding functionality.
type Middleware interface {
	Handler(next http.Handler) http.Handler
}

// MiddlewareFunc allows implementation of Middleware using a regular function.
type MiddlewareFunc func(next http.Handler) http.Handler

// Handler implements Middleware.
func (f MiddlewareFunc) Handler(next http.Handler) http.Handler {
	return f(next)
}

// MiddlewareHandlerFunc is a function that combines MiddlewareFunc with http.HandlerFunc in one function for
// easier Middleware implementation.
type MiddlewareHandlerFunc func(w http.ResponseWriter, req *http.Request, next http.Handler)

// Handler implements Middleware.
func (f MiddlewareHandlerFunc) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		f(w, req, next)
	})
}

// InjectLogger is a Middleware to inject the given logr.Logger into the http.Request.Context.
func InjectLogger(log logr.Logger) Middleware {
	return MiddlewareHandlerFunc(func(w http.ResponseWriter, req *http.Request, next http.Handler) {
		ctx := req.Context()
		ctx = logr.NewContext(ctx, log)
		req = req.WithContext(ctx)
		next.ServeHTTP(w, req)
	})
}

// UseMiddleware returns a http.Handler that has been transformed using all given Middleware.
func UseMiddleware(handler http.Handler, middleware ...Middleware) http.Handler {
	for _, middleware := range middleware {
		handler = middleware.Handler(handler)
	}
	return handler
}
