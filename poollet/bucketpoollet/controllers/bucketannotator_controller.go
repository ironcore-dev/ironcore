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
	ori "github.com/onmetal/onmetal-api/ori/apis/bucket/v1alpha1"
	orimeta "github.com/onmetal/onmetal-api/ori/apis/meta/v1alpha1"
	bucketpoolletv1alpha1 "github.com/onmetal/onmetal-api/poollet/bucketpoollet/api/v1alpha1"
	"github.com/onmetal/onmetal-api/poollet/orievent"
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

type BucketAnnotatorReconciler struct {
	client.Client

	BucketEvents orievent.Source[*ori.Bucket]
}

func (r *BucketAnnotatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	bucket := &storagev1alpha1.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: req.Namespace,
			Name:      req.Name,
		},
	}

	if err := onmetalapiclient.PatchAddReconcileAnnotation(ctx, r.Client, bucket); client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, fmt.Errorf("error patching bucket: %w", err)
	}
	return ctrl.Result{}, nil
}

func bucketAnnotatorEventHandler[O orimeta.Object](log logr.Logger, c chan<- event.GenericEvent) orievent.HandlerFuncs[O] {
	handleEvent := func(obj orimeta.Object) {
		namespace, ok := obj.GetMetadata().Labels[bucketpoolletv1alpha1.BucketNamespaceLabel]
		if !ok {
			return
		}

		name, ok := obj.GetMetadata().Labels[bucketpoolletv1alpha1.BucketNameLabel]
		if !ok {
			return
		}

		bucket := &storagev1alpha1.Bucket{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      name,
			},
		}

		select {
		case c <- event.GenericEvent{Object: bucket}:
		default:
			log.V(5).Info("Channel full, discarding event")
		}
	}

	return orievent.HandlerFuncs[O]{
		CreateFunc: func(event orievent.CreateEvent[O]) {
			handleEvent(event.Object)
		},
		UpdateFunc: func(event orievent.UpdateEvent[O]) {
			handleEvent(event.ObjectNew)
		},
		DeleteFunc: func(event orievent.DeleteEvent[O]) {
			handleEvent(event.Object)
		},
		GenericFunc: func(event orievent.GenericEvent[O]) {
			handleEvent(event.Object)
		},
	}
}

func (r *BucketAnnotatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("bucketannotator", mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}

	src, err := r.oriBucketEventSource(mgr)
	if err != nil {
		return err
	}

	if err := c.Watch(src, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}

	return nil
}

func (r *BucketAnnotatorReconciler) oriBucketEventSource(mgr ctrl.Manager) (source.Source, error) {
	ch := make(chan event.GenericEvent, 1024)

	if err := mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		log := ctrl.LoggerFrom(ctx).WithName("bucketannotator").WithName("orieventhandlers")

		registrationFuncs := []func() (orievent.HandlerRegistration, error){
			func() (orievent.HandlerRegistration, error) {
				return r.BucketEvents.AddHandler(bucketAnnotatorEventHandler[*ori.Bucket](log, ch))
			},
		}

		var handles []orievent.HandlerRegistration
		defer func() {
			log.V(1).Info("Removing handles")
			for _, handle := range handles {
				if err := handle.Remove(); err != nil {
					log.Error(err, "Error removing handle")
				}
			}
		}()

		for _, registrationFunc := range registrationFuncs {
			handle, err := registrationFunc()
			if err != nil {
				return err
			}

			handles = append(handles, handle)
		}

		<-ctx.Done()
		return nil
	})); err != nil {
		return nil, err
	}

	return &source.Channel{Source: ch}, nil
}
