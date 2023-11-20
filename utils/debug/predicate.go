// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package debug

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type debugPredicate struct {
	log         logr.Logger
	predicate   predicate.Predicate
	objectValue func(client.Object) any
}

func (d *debugPredicate) Create(evt event.CreateEvent) bool {
	log := d.log.WithValues("Event", "Create", "Object", d.objectValue(evt.Object))
	log.Info("Handling event")
	res := d.predicate.Create(evt)
	log.Info("Handled event", "Result", res)
	return res
}

func (d *debugPredicate) Delete(evt event.DeleteEvent) bool {
	log := d.log.WithValues("Event", "Delete", "Object", d.objectValue(evt.Object))
	log.Info("Handling event")
	res := d.predicate.Delete(evt)
	log.Info("Handled event", "Result", res)
	return res
}

func (d *debugPredicate) Update(evt event.UpdateEvent) bool {
	log := d.log.WithValues("Event", "Update", "ObjectOld", d.objectValue(evt.ObjectOld), "ObjectNew", d.objectValue(evt.ObjectNew))
	log.Info("Handling event")
	res := d.predicate.Update(evt)
	log.Info("Handled event", "Result", res)
	return res
}

func (d *debugPredicate) Generic(evt event.GenericEvent) bool {
	log := d.log.WithValues("Event", "Generic", "Object", d.objectValue(evt.Object))
	log.Info("Handling event")
	res := d.predicate.Generic(evt)
	log.Info("Handled event", "Result", res)
	return res
}

// Predicate allows debugging a predicate.Predicate by wrapping it and logging each action it does.
//
// Caution: This has a heavy toll on runtime performance and should *not* be used in production code.
// Use only for debugging predicates and remove once done.
func Predicate(name string, prct predicate.Predicate, opts ...PredicateOption) predicate.Predicate {
	o := (&PredicateOptions{}).ApplyOptions(opts)
	setPredicateOptionsDefaults(o)

	return &debugPredicate{
		log:         o.Log.WithName(name),
		predicate:   prct,
		objectValue: o.ObjectValue,
	}
}
