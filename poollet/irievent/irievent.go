// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package irievent

import (
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
)

type CreateEvent[O irimeta.Object] struct {
	Object O
}

type UpdateEvent[O irimeta.Object] struct {
	ObjectOld O
	ObjectNew O
}

type DeleteEvent[O irimeta.Object] struct {
	Object O
}

type GenericEvent[O irimeta.Object] struct {
	Object O
}

type Handler[O irimeta.Object] interface {
	Create(event CreateEvent[O])
	Update(event UpdateEvent[O])
	Delete(event DeleteEvent[O])
	Generic(event GenericEvent[O])
}

type HandlerFuncs[O irimeta.Object] struct {
	CreateFunc  func(event CreateEvent[O])
	UpdateFunc  func(event UpdateEvent[O])
	DeleteFunc  func(event DeleteEvent[O])
	GenericFunc func(event GenericEvent[O])
}

func (e HandlerFuncs[O]) Create(event CreateEvent[O]) {
	if e.CreateFunc != nil {
		e.CreateFunc(event)
	}
}

func (e HandlerFuncs[O]) Update(event UpdateEvent[O]) {
	if e.UpdateFunc != nil {
		e.UpdateFunc(event)
	}
}

func (e HandlerFuncs[O]) Delete(event DeleteEvent[O]) {
	if e.DeleteFunc != nil {
		e.DeleteFunc(event)
	}
}

func (e HandlerFuncs[O]) Generic(event GenericEvent[O]) {
	if e.GenericFunc != nil {
		e.GenericFunc(event)
	}
}

type HandlerRegistration interface {
	Remove() error
}

type Source[O irimeta.Object] interface {
	AddHandler(handler Handler[O]) (HandlerRegistration, error)
}
