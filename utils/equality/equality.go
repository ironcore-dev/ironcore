// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package equality

import (
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/conversion"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/third_party/forked/golang/reflect"
)

// Semantic checks whether ironcore types are semantically equal.
// It uses equality.Semantic as baseline and adds custom functions on top.
var Semantic conversion.Equalities

func init() {
	base := make(reflect.Equalities)
	for k, v := range equality.Semantic.Equalities {
		base[k] = v
	}
	Semantic = conversion.Equalities{Equalities: base}
	utilruntime.Must(AddFuncs(Semantic))
}

func AddFuncs(equality conversion.Equalities) error {
	return equality.AddFuncs(
		commonv1alpha1.EqualIPs,
		commonv1alpha1.EqualIPPrefixes,
		commonv1alpha1.EqualIPRanges,
	)
}
