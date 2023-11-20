// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

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

// LogRequest logs incoming requests (method and URL).
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log := logr.FromContextOrDiscard(req.Context())
		log.V(1).Info("Request", "Method", req.Method, "URL", req.URL)
		next.ServeHTTP(w, req)
	})
}
