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

package ipamrange

import (
	"context"
	common "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	api "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type IPAMStatus struct {
	State            string
	CIDRs            []api.CIDRAllocationStatus
	AllocationState  []string
	PendingRequest   *api.IPAMPendingRequest
	PendingDeletions []api.CIDRAllocationStatus
}

const (
	timeout  = time.Second * 60
	interval = time.Second * 1
)

var ctx = context.Background()

var activeTest = sets.NewString(
	"IPAMRange controller",
	"IPAMRange extension",
	"IPAMRange three level extension",
	"IPAMRange three level deletion",
)

func OptionalDescribe(n string, f func()) bool {
	if activeTest.Has(n) {
		return Describe(n, f)
	}
	return false
}

var cleanUp = func(keys ...client.ObjectKey) {
	for _, key := range keys {
		ipamRange := &api.IPAMRange{}
		err := k8sClient.Get(ctx, types.NamespacedName{
			Namespace: key.Namespace,
			Name:      key.Name,
		}, ipamRange)
		if errors.IsNotFound(err) {
			return
		}
		Expect(err).Should(Succeed())
		Expect(k8sClient.Delete(ctx, ipamRange)).Should(Succeed())
		if len(ipamRange.GetFinalizers()) > 0 {
			newRange := ipamRange.DeepCopy()
			newRange.Finalizers = nil
			Expect(k8sClient.Patch(ctx, newRange, client.MergeFrom(ipamRange))).Should(Succeed())
		}
	}
}

var createObject = func(key client.ObjectKey, parent *common.ScopedReference, cidrs ...string) {
	Expect(k8sClient.Create(ctx, &api.IPAMRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: api.IPAMRangeSpec{
			Parent: parent,
			CIDRs:  cidrs,
		},
	})).Should(Succeed())
}

var updateObject = func(key client.ObjectKey, parent *common.ScopedReference, cidrs ...string) {
	obj := &api.IPAMRange{}
	Expect(k8sClient.Get(ctx, key, obj)).Should(Succeed())
	newObj := obj.DeepCopy()
	newObj.Spec = api.IPAMRangeSpec{
		Parent: parent,
		CIDRs:  cidrs,
	}
	Expect(k8sClient.Patch(ctx, newObj, client.MergeFrom(obj))).Should(Succeed())
}

func projectStatus(ctx context.Context, lookupKey types.NamespacedName) *IPAMStatus {
	obj := &api.IPAMRange{}
	Expect(k8sClient.Get(ctx, lookupKey, obj)).Should(Succeed())
	return &IPAMStatus{
		State:            obj.Status.State,
		CIDRs:            obj.Status.CIDRs,
		PendingDeletions: obj.Status.PendingDeletions,
		AllocationState:  obj.Status.AllocationState,
		PendingRequest:   obj.Status.PendingRequest,
	}
}
