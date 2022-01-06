// Copyright 2021 OnMetal authors
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

package equality

import (
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/conversion"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/third_party/forked/golang/reflect"
)

// Semantic checks whether onmetal types are semantically equal.
// It uses equality.Semantic as baseline and adds custom functions on top.
var Semantic conversion.Equalities

func init() {
	base := make(reflect.Equalities)
	for k, v := range equality.Semantic.Equalities {
		base[k] = v
	}
	Semantic = conversion.Equalities{Equalities: base}
	utilruntime.Must(Semantic.AddFuncs(
		func(a, b commonv1alpha1.IPPrefix) bool {
			return a.String() == b.String()
		},
		func(a, b commonv1alpha1.IP) bool {
			return a.String() == b.String()
		},
	))
}
