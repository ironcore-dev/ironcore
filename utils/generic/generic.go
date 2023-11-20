// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package generic

import (
	"fmt"
	"reflect"
	"strings"
)

// Identity is a function that returns its given parameters.
func Identity[E any](e E) E {
	return e
}

// Const produces a function that takes an argument and returns the original argument, ignoring the passed-in value.
func Const[E, F any](e E) func(F) E {
	return func(F) E {
		return e
	}
}

// Zero returns the zero value for the given type.
func Zero[E any]() E {
	var zero E
	return zero
}

func Cast[E any](v any) (E, error) {
	e, ok := v.(E)
	if !ok {
		return Zero[E](), fmt.Errorf("expected %T but got %T", e, v)
	}
	return e, nil
}

func ReflectType[E any]() reflect.Type {
	var ePtr *E // use a pointer to avoid initializing the entire type
	return reflect.TypeOf(ePtr).Elem()
}

// Pointer returns a pointer for the given value.
func Pointer[E any](e E) *E {
	return &e
}

// ZeroPointer returns a pointer to a zero value of type E.
func ZeroPointer[E any]() *E {
	var zero E
	return &zero
}

// DerefFunc returns the value e points to if it's non-nil. Otherwise, it returns the result of calling defaultFunc.
func DerefFunc[E any](e *E, defaultFunc func() E) E {
	if e != nil {
		return *e
	}
	return defaultFunc()
}

// Deref returns the value e points to if it's non-nil. Otherwise, it returns the defaultValue.
func Deref[E any](e *E, defaultValue E) E {
	return DerefFunc(e, func() E {
		return defaultValue
	})
}

// DerefZero returns the value e points to if it's non-nil. Otherwise, it returns the zero value for type E.
func DerefZero[E any](e *E) E {
	if e != nil {
		return *e
	}
	var zero E
	return zero
}

func PipeMap[E any](e E, fs ...func(E) E) E {
	for _, f := range fs {
		e = f(e)
	}
	return e
}

// TODO is a function to create holes when stubbing out more complex mechanisms.
//
// By default, it will panic with 'TODO: provide a value of type <type>' where <type> is the type of V.
// The panic message can be altered by passing in additional args that will be printed as
// 'TODO: <args separated by space>'
func TODO[V any](args ...any) V {
	var sb strings.Builder
	sb.WriteString("TODO: ")
	if len(args) > 0 {
		_, _ = fmt.Fprintln(&sb, args...)
	} else {
		_, _ = fmt.Fprintf(&sb, "provide a value of type %T", Zero[V]())
	}
	panic(sb.String())
}
