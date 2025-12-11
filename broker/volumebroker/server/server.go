// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/ironcore-dev/controller-utils/metautils"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/common/cleaner"
	"github.com/ironcore-dev/ironcore/broker/common/idgen"
	volumebrokerv1alpha1 "github.com/ironcore-dev/ironcore/broker/volumebroker/api/v1alpha1"
	"github.com/ironcore-dev/ironcore/broker/volumebroker/apiutils"
	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	volumepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/volumepoollet/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubernetes "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
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
	iri.UnimplementedVolumeRuntimeServer

	client client.Client
	idGen  idgen.IDGen

	brokerDownwardAPILabels map[string]string

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
	// BrokerDownwardAPILabels specifies which labels to broker via downward API and what the default
	// label name is to obtain the value in case there is no value for the downward API.
	// Example usage is e.g. to broker the root UID (map "root-volume-uid" to volumepoollet's
	// "volumepoollet.ironcore.dev/volume-uid")
	BrokerDownwardAPILabels map[string]string
	Namespace               string
	VolumePoolName          string
	VolumePoolSelector      map[string]string
	IDGen                   idgen.IDGen
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
//+kubebuilder:rbac:groups=storage.ironcore.dev,resources=volumesnapshots,verbs=get;list;watch;create;update;patch;delete

func New(cfg *rest.Config, opts Options) (*Server, error) {
	setOptionsDefaults(&opts)

	c, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	return &Server{
		brokerDownwardAPILabels: opts.BrokerDownwardAPILabels,
		client:                  c,
		idGen:                   opts.IDGen,
		namespace:               opts.Namespace,
		volumePoolName:          opts.VolumePoolName,
		volumePoolSelector:      opts.VolumePoolSelector,
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

func (s *Server) SetVolumeUIDLabelToAllVolumes(ctx context.Context, log logr.Logger) error {
	volumeList := &storagev1alpha1.VolumeList{}
	log.V(2).Info("Listing brokered volumes")
	if err := s.listManagedAndCreated(ctx, volumeList, nil); err != nil {
		return fmt.Errorf("error listing ironcore volumes: %w", err)
	}

	for i := range volumeList.Items {
		volume := &volumeList.Items[i]
		labels, err := apiutils.GetLabelsAnnotation(volume)
		if err != nil {
			return fmt.Errorf("failed to get labels annotation: %w", err)
		}
		if labels == nil {
			log.V(2).Info("Labels are nil", "name", volume.Name, "namespace", volume.Namespace)
			continue
		}

		volumeUid := labels[volumepoolletv1alpha1.VolumeUIDLabel]
		if volumeUid == "" {
			log.V(2).Info("Volume uid label is empty", "name", volume.Name, "namespace", volume.Namespace)
			continue
		}
		log.V(2).Info("Setting volume uid label for", "name", volume.Name, "namespace", volume.Namespace, "labelKey", volumepoolletv1alpha1.VolumeUIDLabel, "labelValue", volumeUid)
		base := volume.DeepCopy()
		metautils.SetLabel(volume, volumepoolletv1alpha1.VolumeUIDLabel, volumeUid)
		if err := s.client.Patch(ctx, volume, client.MergeFrom(base)); err != nil {
			return fmt.Errorf("error patching volume uid label: %w", err)
		}
	}
	return nil
}
