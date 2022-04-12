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

package validation_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Validation Suite")
}

func UnsupportedField(fld string) types.GomegaMatcher {
	return FieldError(field.ErrorTypeNotSupported, fld)
}

func ForbiddenField(fld string) types.GomegaMatcher {
	return FieldError(field.ErrorTypeForbidden, fld)
}

func InvalidField(fld string) types.GomegaMatcher {
	return FieldError(field.ErrorTypeInvalid, fld)
}

func RequiredField(fld string) types.GomegaMatcher {
	return FieldError(field.ErrorTypeRequired, fld)
}

func ImmutableField(fld string) types.GomegaMatcher {
	return And(
		MaskedFieldError(func(error *field.Error) {
			error.BadValue = nil
			error.Detail = ""
		}, &field.Error{
			Type:  field.ErrorTypeInvalid,
			Field: fld,
		}),
		WithTransform(func(error *field.Error) string {
			return error.Detail
		}, HavePrefix(validation.FieldImmutableErrorMsg)),
	)
}

func MaskedFieldError(maskFn func(error *field.Error), fieldErr *field.Error) types.GomegaMatcher {
	return WithTransform(func(fieldErr *field.Error) *field.Error {
		res := *fieldErr
		maskFn(&res)
		return &res
	}, Equal(fieldErr))
}

func FieldError(errorType field.ErrorType, fld string) types.GomegaMatcher {
	return MaskedFieldError(func(error *field.Error) {
		error.Detail = ""
		error.BadValue = nil
	}, &field.Error{
		Type:  errorType,
		Field: fld,
	})
}
