// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"

	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	ctrl "sigs.k8s.io/controller-runtime"
)

// InjectLogger injects the given logr.Logger into the context using ctrl.LoggerInto.
func InjectLogger(log logr.Logger) grpc.UnaryServerInterceptor {
	return grpc.UnaryServerInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		log := log.WithName(info.FullMethod)
		ctx = ctrl.LoggerInto(ctx, log)
		return handler(ctx, req)
	})
}

// LogRequest logs grpc requests. In case any request returns with status.Code == codes.Unknown, the error is logged.
var LogRequest = grpc.UnaryServerInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	log := ctrl.LoggerFrom(ctx)
	log.V(1).Info("Request")
	resp, err = handler(ctx, req)
	if err != nil {
		if code := status.Code(err); code == codes.Unknown {
			log.Error(err, "Unknown error handling request")
		}
	}
	return resp, err
})
