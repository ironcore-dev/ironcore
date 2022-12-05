// Copyright 2022 OnMetal authors
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
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	machinebrokerv1alpha1 "github.com/onmetal/onmetal-api/machinebroker/api/v1alpha1"
	"github.com/onmetal/onmetal-api/machinebroker/apiutils"
	"github.com/onmetal/onmetal-api/machinebroker/cleaner"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func objectStructsToObjectPtrByNameMap[Obj any, ObjPtr ObjPtrable[Obj], S ~[]Obj](items S) map[string]*Obj {
	res := make(map[string]*Obj)
	for i := range items {
		itemPtr := &items[i]
		name := (ObjPtr(itemPtr)).GetName()
		res[name] = itemPtr
	}
	return res
}

func objectByNameMapGetter[Obj client.Object](gr schema.GroupResource, m map[string]Obj) func(name string) (Obj, error) {
	return func(name string) (Obj, error) {
		obj, ok := m[name]
		if ok {
			return obj, nil
		}
		return obj, apierrors.NewNotFound(gr, name)
	}
}

type ObjPtrable[E any] interface {
	*E
	client.Object
}

func clientGetter[Obj any, ObjPtr ObjPtrable[Obj]](ctx context.Context, c client.Client, namespace string) func(name string) (ObjPtr, error) {
	return func(name string) (ObjPtr, error) {
		obj := ObjPtr(new(Obj))
		if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, obj); err != nil {
			return nil, err
		}
		return obj, nil
	}
}

func (s *Server) loggerFrom(ctx context.Context, keysWithValues ...interface{}) logr.Logger {
	return ctrl.LoggerFrom(ctx, keysWithValues...)
}

func (s *Server) listManagedAndCreated(ctx context.Context, list client.ObjectList) error {
	return s.client.List(ctx, list,
		client.InNamespace(s.namespace),
		client.MatchingLabels{
			machinebrokerv1alpha1.ManagerLabel: machinebrokerv1alpha1.MachineBrokerManager,
			machinebrokerv1alpha1.CreatedLabel: "true",
		},
	)
}

func (s *Server) listWithPurpose(ctx context.Context, list client.ObjectList, purpose string) error {
	return s.client.List(ctx, list,
		client.InNamespace(s.namespace),
		client.MatchingLabels{
			machinebrokerv1alpha1.PurposeLabel: purpose,
		},
	)
}

func (s *Server) getManagedAndCreated(ctx context.Context, name string, obj client.Object) error {
	key := client.ObjectKey{Namespace: s.namespace, Name: name}
	if err := s.client.Get(ctx, key, obj); err != nil {
		return err
	}
	if !apiutils.IsManagedBy(obj, machinebrokerv1alpha1.MachineBrokerManager) || !apiutils.IsCreated(obj) {
		gvk, err := apiutil.GVKForObject(obj, s.client.Scheme())
		if err != nil {
			return err
		}

		return apierrors.NewNotFound(schema.GroupResource{
			Group:    gvk.Group,
			Resource: gvk.Kind, // Yes, kind is good enough here
		}, key.Name)
	}
	return nil
}

// idLength is 63 as this is the highest common denominator among resource name length limitations (e.g. k8s secret).
const idLength = 63

func (s *Server) generateID() string {
	data := make([]byte, 32)
	for {
		_, _ = rand.Read(data)
		id := hex.EncodeToString(data)

		// Truncated versions of the id should not be numerical.
		if _, err := strconv.ParseInt(id[:12], 10, 64); err != nil {
			continue
		}

		return id[:idLength]
	}
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

func (s *Server) convertOnmetalIPSourcesToIPs(ipSources []networkingv1alpha1.IPSource) ([]string, error) {
	res := make([]string, len(ipSources))
	for i, ipSource := range ipSources {
		if ipSource.Value == nil {
			return nil, fmt.Errorf("ip source %d does not specify an ip literal", i)
		}
		res[i] = ipSource.Value.String()
	}
	return res, nil
}

func (s *Server) getOnmetalIPsIPFamilies(ips []commonv1alpha1.IP) []corev1.IPFamily {
	res := make([]corev1.IPFamily, len(ips))
	for i, ip := range ips {
		res[i] = ip.Family()
	}
	return res
}

func (s *Server) onmetalIPsToOnmetalIPSources(ips []commonv1alpha1.IP) []networkingv1alpha1.IPSource {
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
