// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/common/cleaner"
	metautils "github.com/ironcore-dev/ironcore/utils/meta"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	ironcoreMachineGVK = computev1alpha1.SchemeGroupVersion.WithKind("Machine")

	ErrMachineNotFound    = errors.New("machine not found")
	ErrMachineIsntManaged = errors.New("machine isn't managed")
)

func (s *Server) loggerFrom(ctx context.Context, keysWithValues ...interface{}) logr.Logger {
	return ctrl.LoggerFrom(ctx, keysWithValues...)
}

func (s *Server) setupCleaner(ctx context.Context, log logr.Logger, retErr *error) (c *cleaner.Cleaner, cleanup func()) {
	c = cleaner.New()
	cleanup = func() {
		if *retErr != nil {
			select {
			case <-ctx.Done():
				log.Info("Cannot do cleanup since context expired")
				return
			default:
				if err := c.Cleanup(ctx); err != nil {
					log.Error(err, "Error cleaning up")
				}
			}
		}
	}
	return c, cleanup
}

func (s *Server) convertIronCoreIPSourcesToIPs(ipSources []networkingv1alpha1.IPSource) ([]string, error) {
	res := make([]string, len(ipSources))
	for i, ipSource := range ipSources {
		if ipSource.Value == nil {
			return nil, fmt.Errorf("ip source %d does not specify an ip literal", i)
		}
		res[i] = ipSource.Value.String()
	}
	return res, nil
}

func (s *Server) getIronCoreIPsIPFamilies(ips []commonv1alpha1.IP) []corev1.IPFamily {
	res := make([]corev1.IPFamily, len(ips))
	for i, ip := range ips {
		res[i] = ip.Family()
	}
	return res
}

func (s *Server) ironcoreIPsToIronCoreIPSources(ips []commonv1alpha1.IP) []networkingv1alpha1.IPSource {
	res := make([]networkingv1alpha1.IPSource, len(ips))
	for i := range ips {
		res[i] = networkingv1alpha1.IPSource{
			Value: &ips[i],
		}
	}
	return res
}

func (s *Server) parseIPs(ipStrings []string) ([]commonv1alpha1.IP, error) {
	var ips []commonv1alpha1.IP
	for _, ipString := range ipStrings {
		ip, err := commonv1alpha1.ParseIP(ipString)
		if err != nil {
			return nil, fmt.Errorf("error parsing ip %q: %w", ipString, err)
		}

		ips = append(ips, ip)
	}
	return ips, nil
}

func (s *Server) optionalOwnerReferences(gvk schema.GroupVersionKind, optionalOwner metav1.Object) []metav1.OwnerReference {
	if optionalOwner == nil {
		return nil
	}

	return []metav1.OwnerReference{
		metautils.MakeControllerRef(
			gvk,
			optionalOwner,
		),
	}
}

func (s *Server) optionalLocalUIDReference(optionalObj metav1.Object) *commonv1alpha1.LocalUIDReference {
	if optionalObj == nil {
		return nil
	}
	return &commonv1alpha1.LocalUIDReference{
		Name: optionalObj.GetName(),
		UID:  optionalObj.GetUID(),
	}
}

func (s *Server) localObjectReferenceTo(obj metav1.Object) commonv1alpha1.LocalUIDReference {
	return commonv1alpha1.LocalUIDReference{
		Name: obj.GetName(),
		UID:  obj.GetUID(),
	}
}

func convertInternalErrorToGRPC(err error) error {
	_, ok := status.FromError(err)
	if ok {
		return err
	}

	code := codes.Internal

	switch {
	case errors.Is(err, ErrMachineNotFound):
		code = codes.NotFound
	case errors.Is(err, ErrMachineIsntManaged):
		code = codes.InvalidArgument
	}

	return status.Error(code, err.Error())
}
