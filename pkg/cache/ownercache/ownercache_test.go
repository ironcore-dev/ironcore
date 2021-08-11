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

package ownercache

import (
	api "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"github.com/onmetal/onmetal-api/pkg/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("OwnerCache", func() {

	var cache *ownerCache
	ipamrange1s1 := createTestObject("o1", "s1")
	ipamrange1Empty := createTestObject("o1")
	ipamrange1s2 := createTestObject("o1", "s2")

	o1 := utils.NewObjectId(ipamrange1s1)
	s1 := singleId(utils.GetOwnerIdsFor(ipamrange1s1))
	s2 := singleId(utils.GetOwnerIdsFor(ipamrange1s2))

	BeforeEach(func() {
		cache = NewOwnerCache(nil, nil)
	})

	Context("Adding", func() {
		It("adding an object", func() {
			cache.ReplaceObject(ipamrange1s1)
			Expect(cache.GetOwnersFor(o1)).To(Equal(utils.NewObjectIds(s1)))
			Expect(cache.GetOwnersByTypeFor(o1, api.IPAMRangeGK)).To(BeNil())
			Expect(cache.GetOwnersByTypeFor(o1, api.SubnetGK)).To(Equal(utils.NewObjectIds(s1)))
			Expect(cache.GetSerfsFor(s1)).To(Equal(utils.NewObjectIds(o1)))
		})
	})

	Context("Deleting", func() {
		It("deleting an object", func() {
			cache.ReplaceObject(ipamrange1s1)
			cache.ReplaceObject(ipamrange1Empty)
			Expect(cache.GetOwnersFor(o1)).To(BeNil())
			Expect(cache.GetSerfsFor(s1)).To(BeNil())
		})
	})

	Context("Replacing", func() {
		It("replacing an object", func() {
			cache.ReplaceObject(ipamrange1s1)
			cache.ReplaceObject(ipamrange1s2)
			Expect(cache.GetOwnersFor(o1)).To(Equal(utils.NewObjectIds(s2)))
			Expect(cache.GetSerfsFor(s1)).To(BeNil())
			Expect(cache.GetSerfsFor(s2)).To(Equal(utils.NewObjectIds(o1)))
		})
	})

})

func singleId(oids utils.ObjectIds) utils.ObjectId {
	for id := range oids {
		return id
	}
	panic("no objectids found")
}

func createTestObject(name string, owners ...string) client.Object {
	obj := &api.IPAMRange{
		TypeMeta: metav1.TypeMeta{
			Kind:       api.IPAMRangeGK.Kind,
			APIVersion: api.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
	}
	for _, o := range owners {
		obj.OwnerReferences = append(obj.OwnerReferences, metav1.OwnerReference{
			APIVersion: api.GroupVersion.String(),
			Kind:       api.SubnetGK.Kind,
			Name:       o,
		})
	}
	return obj
}
