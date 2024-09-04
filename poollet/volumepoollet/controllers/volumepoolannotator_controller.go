// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/irievent"
	"github.com/ironcore-dev/ironcore/poollet/volumepoollet/vcm"
	ironcoreclient "github.com/ironcore-dev/ironcore/utils/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type VolumePoolAnnotatorReconciler struct {
	client.Client

	VolumePoolName    string
	VolumeClassMapper vcm.VolumeClassMapper
}

func (r *VolumePoolAnnotatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	volumePool := &storagev1alpha1.VolumePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Name,
		},
	}

	if err := ironcoreclient.PatchAddReconcileAnnotation(ctx, r.Client, volumePool); client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, fmt.Errorf("error patching volume pool: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *VolumePoolAnnotatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("volumepoolannotator", mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}

	src, err := r.iriClassEventSource(mgr)
	if err != nil {
		return err
	}

	if err := c.Watch(src); err != nil {
		return err
	}

	return nil
}

func (r *VolumePoolAnnotatorReconciler) volumePoolAnnotatorEventHandler(log logr.Logger, c chan<- event.GenericEvent) irievent.EnqueueFunc {
	handleEvent := func() {
		select {
		case c <- event.GenericEvent{Object: &storagev1alpha1.VolumePool{ObjectMeta: metav1.ObjectMeta{
			Name: r.VolumePoolName,
		}}}:
			log.V(1).Info("Added item to queue")
		default:
			log.V(5).Info("Channel full, discarding event")
		}
	}

	return irievent.EnqueueFunc{EnqueueFunc: handleEvent}
}

func (r *VolumePoolAnnotatorReconciler) iriClassEventSource(mgr ctrl.Manager) (source.Source, error) {
	ch := make(chan event.GenericEvent, 1024)

	if err := mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		log := ctrl.LoggerFrom(ctx).WithName("volumepool").WithName("irieventhandlers")

		notifierFuncs := []func() (irievent.ListenerRegistration, error){
			func() (irievent.ListenerRegistration, error) {
				return r.VolumeClassMapper.AddListener(r.volumePoolAnnotatorEventHandler(log, ch))
			},
		}

		var notifier []irievent.ListenerRegistration
		defer func() {
			log.V(1).Info("Removing notifier")
			for _, n := range notifier {
				if err := r.VolumeClassMapper.RemoveListener(n); err != nil {
					log.Error(err, "Error removing handle")
				}
			}
		}()

		for _, notifierFunc := range notifierFuncs {
			ntf, err := notifierFunc()
			if err != nil {
				return err
			}

			notifier = append(notifier, ntf)
		}

		<-ctx.Done()
		return nil
	})); err != nil {
		return nil, err
	}

	return source.Channel(ch, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		return []ctrl.Request{{NamespacedName: client.ObjectKey{Name: r.VolumePoolName}}}
	})), nil
}
