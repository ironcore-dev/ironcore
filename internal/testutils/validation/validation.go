// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

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
