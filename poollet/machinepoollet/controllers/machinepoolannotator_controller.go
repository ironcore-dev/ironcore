// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"github.com/ironcore-dev/ironcore/poollet/irievent"
	"github.com/ironcore-dev/ironcore/poollet/machinepoollet/mcm"
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

type MachinePoolAnnotatorReconciler struct {
	client.Client

	MachinePoolName    string
	MachineClassMapper mcm.MachineClassMapper
}

func (r *MachinePoolAnnotatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	machinePool := &computev1alpha1.MachinePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Name,
		},
	}

	if err := ironcoreclient.PatchAddReconcileAnnotation(ctx, r.Client, machinePool); client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, fmt.Errorf("error patching machine pool: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *MachinePoolAnnotatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("machinepoolannotator", mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}

	src, err := r.iriClassEventSource(mgr)
	if err != nil {
		return err
	}

	if err := c.Watch(src, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
		return []ctrl.Request{{NamespacedName: client.ObjectKey{Name: r.MachinePoolName}}}
	})); err != nil {
		return err
	}

	return nil
}

func (r *MachinePoolAnnotatorReconciler) machinePoolAnnotatorEventHandler(log logr.Logger, c chan<- event.GenericEvent) irievent.EnqueueFunc {
	handleEvent := func() {
		select {
		case c <- event.GenericEvent{Object: &computev1alpha1.MachinePool{ObjectMeta: metav1.ObjectMeta{
			Name: r.MachinePoolName,
		}}}:
			log.V(1).Info("Added item to queue")
		default:
			log.V(5).Info("Channel full, discarding event")
		}
	}

	return irievent.EnqueueFunc{EnqueueFunc: handleEvent}
}

func (r *MachinePoolAnnotatorReconciler) iriClassEventSource(mgr ctrl.Manager) (source.Source, error) {
	ch := make(chan event.GenericEvent, 1024)

	if err := mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		log := ctrl.LoggerFrom(ctx).WithName("machinepool").WithName("irieventhandlers")

		notifierFuncs := []func() (irievent.ListenerRegistration, error){
			func() (irievent.ListenerRegistration, error) {
				return r.MachineClassMapper.AddListener(r.machinePoolAnnotatorEventHandler(log, ch))
			},
		}

		var notifier []irievent.ListenerRegistration
		defer func() {
			log.V(1).Info("Removing notifier")
			for _, n := range notifier {
				if err := r.MachineClassMapper.RemoveListener(n); err != nil {
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
