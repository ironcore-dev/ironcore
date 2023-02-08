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

package validation

import (
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	"k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func InvalidField(fld string) types.GomegaMatcher {
	return SimpleMatchField(field.ErrorTypeInvalid, fld)
}

func RequiredField(fld string) types.GomegaMatcher {
	return SimpleMatchField(field.ErrorTypeRequired, fld)
}

func NotSupportedField(fld string) types.GomegaMatcher {
	return SimpleMatchField(field.ErrorTypeNotSupported, fld)
}

func DuplicateField(fld string) types.GomegaMatcher {
	return SimpleMatchField(field.ErrorTypeDuplicate, fld)
}

func ForbiddenField(fld string) types.GomegaMatcher {
	return SimpleMatchField(field.ErrorTypeForbidden, fld)
}

func SimpleMatchField(errorType field.ErrorType, fld string) types.GomegaMatcher {
	return gomega.HaveValue(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
		"Type":  gomega.Equal(errorType),
		"Field": gomega.Equal(fld),
	}))
}

func ImmutableField(fld string) types.GomegaMatcher {
	return gomega.HaveValue(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
		"Type":   gomega.Equal(field.ErrorTypeForbidden),
		"Detail": gomega.HavePrefix(validation.FieldImmutableErrorMsg),
		"Field":  gomega.Equal(fld),
	}))
}
