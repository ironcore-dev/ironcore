// Copyright 2023 OnMetal authors
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

package debug

import (
	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HandlerOptions are options for construction a debug handler.
type HandlerOptions struct {
	// Log is the logger to use. If unspecified, the debug package logger will be used.
	Log logr.Logger

	// ObjectValue controls how an object will be represented as in the log values.
	ObjectValue func(client.Object) any
}

func (o *HandlerOptions) ApplyOptions(opts []HandlerOption) *HandlerOptions {
	for _, opt := range opts {
		opt.ApplyToHandler(o)
	}
	return o
}

func setHandlerOptionsDefaults(o *HandlerOptions) {
	if o.Log.GetSink() == nil {
		o.Log = handlerLog
	}
	if o.ObjectValue == nil {
		o.ObjectValue = DefaultObjectValue
	}
}

// PredicateOptions are options for construction a debug predicate.
type PredicateOptions struct {
	// Log is the logger to use. If unspecified, the debug package logger will be used.
	Log logr.Logger

	// ObjectValue controls how an object will be represented as in the log values.
	ObjectValue func(client.Object) any
}

func (o *PredicateOptions) ApplyToPredicate(o2 *PredicateOptions) {
	if o.Log.GetSink() != nil {
		o2.Log = o.Log
	}
	if o.ObjectValue != nil {
		o2.ObjectValue = DefaultObjectValue
	}
}

func (o *PredicateOptions) ApplyOptions(opts []PredicateOption) *PredicateOptions {
	for _, opt := range opts {
		opt.ApplyToPredicate(o)
	}
	return o
}

func setPredicateOptionsDefaults(o *PredicateOptions) {
	if o.Log.GetSink() == nil {
		o.Log = predicateLog
	}
	if o.ObjectValue == nil {
		o.ObjectValue = DefaultObjectValue
	}
}

type PredicateOption interface {
	ApplyToPredicate(o *PredicateOptions)
}

// DefaultObjectValue provides object logging values by using klog.KObj.
func DefaultObjectValue(obj client.Object) any {
	return klog.KObj(obj)
}

type HandlerOption interface {
	ApplyToHandler(o *HandlerOptions)
}

// WithLog specifies the logger to use.
type WithLog struct {
	Log logr.Logger
}

func (w WithLog) ApplyToHandler(o *HandlerOptions) {
	o.Log = w.Log
}

func (w WithLog) ApplyToPredicate(o *PredicateOptions) {
	o.Log = w.Log
}

// WithObjectValue specifies the function to log an client.Object's value with.
type WithObjectValue func(obj client.Object) any

func (w WithObjectValue) ApplyToHandler(o *HandlerOptions) {
	o.ObjectValue = w
}

func (w WithObjectValue) ApplyToPredicate(o *PredicateOptions) {
	o.ObjectValue = w
}
