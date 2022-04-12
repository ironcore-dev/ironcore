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

package ipam

import (
	"context"

	ipamv1alpha1 "github.com/onmetal/onmetal-api/apis/ipam/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	prefixSpecIPFamilyField = "spec.ipFamily"
)

func SetupPrefixSpecIPFamilyFieldIndexer(mgr ctrl.Manager) error {
	ctx := context.Background()
	return mgr.GetFieldIndexer().IndexField(ctx, &ipamv1alpha1.Prefix{}, prefixSpecIPFamilyField, func(obj client.Object) []string {
		prefix := obj.(*ipamv1alpha1.Prefix)
		return []string{string(prefix.Spec.IPFamily)}
	})
}

type readerClient struct {
	client.Reader
	noReaderClient
}

type noReaderClient interface {
	client.Writer
	client.StatusClient
	Scheme() *runtime.Scheme
	RESTMapper() meta.RESTMapper
}

func ReaderClient(reader client.Reader, c client.Client) client.Client {
	return readerClient{reader, c}
}
