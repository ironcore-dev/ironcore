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
		func(a, b commonv1alpha1.CIDR) bool {
			return a.String() == b.String()
		},
		func(a, b commonv1alpha1.IPAddr) bool {
			return a.String() == b.String()
		},
	))
}
