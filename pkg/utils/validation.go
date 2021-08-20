/*
 * Copyright (c) 2021 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import "k8s.io/apimachinery/pkg/util/validation/field"

const FieldIsImmutable = "field is immutable"

// ValidateFieldIsImmutatable validates if a value should be changed and returns a field.Error if that is the case.
func ValidateFieldIsImmutatable(new, old string, fldPath *field.Path) *field.Error {
	if new != old {
		return field.Invalid(fldPath, new, FieldIsImmutable)
	}
	return nil
}
