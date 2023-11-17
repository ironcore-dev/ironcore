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
