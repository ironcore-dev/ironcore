// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/common/cleaner"
	"github.com/ironcore-dev/ironcore/broker/common/idgen"
	volumebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/volumebroker/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/volumebroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubernetes "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(kubernetes.AddToScheme(scheme))
	utilruntime.Must(computev1alpha1.AddToScheme(scheme))
	utilruntime.Must(networkingv1alpha1.AddToScheme(scheme))
	utilruntime.Must(storagev1alpha1.AddToScheme(scheme))
	utilruntime.Must(ipamv1alpha1.AddToScheme(scheme))
}

type Server struct {
	client client.Client
	idGen  idgen.IDGen

	namespace          string
	volumePoolName     string
	volumePoolSelector map[string]string
}

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

type Options struct {
	Namespace          string
	VolumePoolName     string
	VolumePoolSelector map[string]string
	IDGen              idgen.IDGen
}

func setOptionsDefaults(o *Options) {
	if o.Namespace == "" {
		o.Namespace = corev1.NamespaceDefault
	}
	if o.IDGen == nil {
		o.IDGen = idgen.Default
	}
}

var _ iri.VolumeRuntimeServer = (*Server)(nil)

//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumes,verbs=get;list;watch;create;update;patch;delete

func New(ctx context.Context, cfg *rest.Config, opts Options) (*Server, error) {
	setOptionsDefaults(&opts)

	readCache, err := cache.New(cfg, cache.Options{
		Scheme: scheme,
		DefaultNamespaces: map[string]cache.Config{
			opts.Namespace: {},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating cache: %w", err)
	}

	go func() {
		if err := readCache.Start(ctx); err != nil {
			fmt.Printf("Error starting cache: %v\n", err)
		}
	}()
	if !readCache.WaitForCacheSync(ctx) {
		return nil, fmt.Errorf("failed to sync cache")
	}

	c, err := client.New(cfg, client.Options{
		Scheme: scheme,
		Cache: &client.CacheOptions{
			Reader: readCache,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	return &Server{
		client:             c,
		idGen:              opts.IDGen,
		namespace:          opts.Namespace,
		volumePoolName:     opts.VolumePoolName,
		volumePoolSelector: opts.VolumePoolSelector,
	}, nil
}

func (s *Server) getManagedAndCreated(ctx context.Context, name string, obj client.Object) error {
	key := client.ObjectKey{Namespace: s.namespace, Name: name}
	if err := s.client.Get(ctx, key, obj); err != nil {
		return err
	}
	if !apiutils.IsManagedBy(obj, volumebrokerv1alpha1.VolumeBrokerManager) || !apiutils.IsCreated(obj) {
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
