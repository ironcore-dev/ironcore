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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	"github.com/onmetal/onmetal-api/poollet/orievent"
	"github.com/onmetal/onmetal-api/poollet/volumepoollet/vcm"
	onmetalapiclient "github.com/onmetal/onmetal-api/utils/client"
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

	if err := onmetalapiclient.PatchAddReconcileAnnotation(ctx, r.Client, volumePool); client.IgnoreNotFound(err) != nil {
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

	src, err := r.oriClassEventSource(mgr)
	if err != nil {
		return err
	}

	if err := c.Watch(src, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		return []ctrl.Request{{NamespacedName: client.ObjectKey{Name: r.VolumePoolName}}}
	})); err != nil {
		return err
	}

	return nil
}

func (r *VolumePoolAnnotatorReconciler) volumePoolAnnotatorEventHandler(log logr.Logger, c chan<- event.GenericEvent) orievent.EnqueueFunc {
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

	return orievent.EnqueueFunc{EnqueueFunc: handleEvent}
}

func (r *VolumePoolAnnotatorReconciler) oriClassEventSource(mgr ctrl.Manager) (source.Source, error) {
	ch := make(chan event.GenericEvent, 1024)

	if err := mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		log := ctrl.LoggerFrom(ctx).WithName("volumepool").WithName("orieventhandlers")

		notifierFuncs := []func() (orievent.ListenerRegistration, error){
			func() (orievent.ListenerRegistration, error) {
				return r.VolumeClassMapper.AddListener(r.volumePoolAnnotatorEventHandler(log, ch))
			},
		}

		var notifier []orievent.ListenerRegistration
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

	return &source.Channel{Source: ch}, nil
}
