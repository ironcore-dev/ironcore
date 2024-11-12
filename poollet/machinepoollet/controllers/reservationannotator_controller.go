// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/irievent"
	machinepoolletv1alpha1 "github.com/ironcore-dev/ironcore/poollet/machinepoollet/api/v1alpha1"
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

type ReservationAnnotatorReconciler struct {
	client.Client

	ReservationEvents irievent.Source[*iri.Reservation]
}

func (r *ReservationAnnotatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reservation := &computev1alpha1.Reservation{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: req.Namespace,
			Name:      req.Name,
		},
	}

	if err := ironcoreclient.PatchAddReconcileAnnotation(ctx, r.Client, reservation); client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, fmt.Errorf("error patching reservation: %w", err)
	}
	return ctrl.Result{}, nil
}

func reservationAnnotatorEventHandler[O irimeta.Object](log logr.Logger, c chan<- event.GenericEvent) irievent.HandlerFuncs[O] {
	handleEvent := func(obj irimeta.Object) {
		namespace, ok := obj.GetMetadata().Labels[machinepoolletv1alpha1.ReservationNamespaceLabel]
		if !ok {
			return
		}

		name, ok := obj.GetMetadata().Labels[machinepoolletv1alpha1.ReservationNameLabel]
		if !ok {
			return
		}

		reservation := &computev1alpha1.Reservation{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      name,
			},
		}

		select {
		case c <- event.GenericEvent{Object: reservation}:
		default:
			log.V(5).Info("Channel full, discarding event")
		}
	}

	return irievent.HandlerFuncs[O]{
		CreateFunc: func(event irievent.CreateEvent[O]) {
			handleEvent(event.Object)
		},
		UpdateFunc: func(event irievent.UpdateEvent[O]) {
			handleEvent(event.ObjectNew)
		},
		DeleteFunc: func(event irievent.DeleteEvent[O]) {
			handleEvent(event.Object)
		},
		GenericFunc: func(event irievent.GenericEvent[O]) {
			handleEvent(event.Object)
		},
	}
}

func (r *ReservationAnnotatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("reservationannotator", mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}

	src, err := r.iriMachineEventSource(mgr)
	if err != nil {
		return err
	}

	if err := c.Watch(src, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}

	return nil
}

func (r *ReservationAnnotatorReconciler) iriMachineEventSource(mgr ctrl.Manager) (source.Source, error) {
	ch := make(chan event.GenericEvent, 1024)

	if err := mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		log := ctrl.LoggerFrom(ctx).WithName("reservationannotator").WithName("irieventhandlers")

		registrationFuncs := []func() (irievent.HandlerRegistration, error){
			func() (irievent.HandlerRegistration, error) {
				return r.ReservationEvents.AddHandler(reservationAnnotatorEventHandler[*iri.Reservation](log, ch))
			},
		}

		var handles []irievent.HandlerRegistration
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
