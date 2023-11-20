// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package runtime

import "github.com/ironcore-dev/ironcore/utils/slices"

type DeepCopier[E any] interface {
	DeepCopy() E
}

type RefDeepCopier[E any] interface {
	*E
	DeepCopier[*E]
}

func DeepCopySlice[E DeepCopier[E], S ~[]E](slice S) S {
	return slices.Map(slice, func(e E) E {
		return e.DeepCopy()
	})
}

// DeepCopySliceRefs runs DeepCopy on the references of the elements of the slice and returns the created structs.
func DeepCopySliceRefs[E any, D RefDeepCopier[E], S ~[]E](slice S) []E {
	return slices.MapRef(slice, func(e *E) E {
		return *(D(e)).DeepCopy()
	})
}
