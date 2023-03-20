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

package generic

// Identity is a function that returns its given parameters.
func Identity[E any](e E) E {
	return e
}

// Zero returns the zero value for the given type.
func Zero[E any]() E {
	var zero E
	return zero
}

// Pointer returns a pointer for the given value.
func Pointer[E any](e E) *E {
	return &e
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
