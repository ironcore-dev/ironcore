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

package usagecache

import (
	"github.com/onmetal/onmetal-api/pkg/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("UsageCache", func() {

	var cache *usageCache
	objectk10 := utils.MustParseObjectId("k1.g1/default/object0")
	objectk11 := utils.MustParseObjectId("k1.g1/default/object1")
	objectk12 := utils.MustParseObjectId("k1.g1/default/object2")
	objectk13 := utils.MustParseObjectId("k1.g1/default/object3")
	//objectk20 := utils.MustParseObjectId("k2.g1/default/object0")
	objectk21 := utils.MustParseObjectId("k2.g1/default/object1")

	gk1 := schema.GroupKind{
		Group: "g1",
		Kind:  "k1",
	}
	gk2 := schema.GroupKind{
		Group: "g1",
		Kind:  "k2",
	}
	_ = gk1
	_ = gk2

	BeforeEach(func() {
		cache = NewUsageCache(nil, nil)
	})

	Context("adding", func() {
		It("simple usage", func() {
			info := NewObjectUsageInfo("uses", objectk11)
			cache.ReplaceObjectUsageInfo(objectk10, info)
			Expect(cache.GetUsedObjectsFor(objectk10)).To(Equal(utils.NewObjectIds(objectk11)))
			Expect(cache.GetUsersFor(objectk11)).To(Equal(utils.NewObjectIds(objectk10)))
		})
		It("two relations", func() {
			info := NewObjectUsageInfo("uses", objectk11, "owner", objectk12)
			cache.ReplaceObjectUsageInfo(objectk10, info)
			Expect(cache.GetUsedObjectsFor(objectk10)).To(Equal(utils.NewObjectIds(objectk11, objectk12)))
			Expect(cache.GetUsersFor(objectk11)).To(Equal(utils.NewObjectIds(objectk10)))
			Expect(cache.GetUsersFor(objectk12)).To(Equal(utils.NewObjectIds(objectk10)))
		})
		It("query two relations", func() {
			info := NewObjectUsageInfo("uses", objectk11, "owner", objectk12)
			cache.ReplaceObjectUsageInfo(objectk10, info)
			Expect(cache.GetUsedObjectsForRelation(objectk10, "uses")).To(Equal(utils.NewObjectIds(objectk11)))
			Expect(cache.GetUsedObjectsForRelation(objectk10, "owner")).To(Equal(utils.NewObjectIds(objectk12)))
			Expect(cache.GetUsersForRelation(objectk11, "uses")).To(Equal(utils.NewObjectIds(objectk10)))
			Expect(cache.GetUsersForRelation(objectk11, "owner")).To(BeNil())
			Expect(cache.GetUsersForRelation(objectk12, "uses")).To(BeNil())
			Expect(cache.GetUsersForRelation(objectk12, "owner")).To(Equal(utils.NewObjectIds(objectk10)))
		})
	})

	Context("deleting", func() {
		It("simple usage", func() {
			info := NewObjectUsageInfo("uses", objectk11)
			cache.ReplaceObjectUsageInfo(objectk10, info)
			cache.ReplaceObjectUsageInfo(objectk10, NewObjectUsageInfo())
			Expect(cache.GetUsedObjectsFor(objectk10)).To(BeNil())
			Expect(cache.GetUsersFor(objectk11)).To(BeNil())
		})
		It("two relations", func() {
			info := NewObjectUsageInfo("uses", objectk11, "owner", objectk12)
			cache.ReplaceObjectUsageInfo(objectk10, info)
			cache.ReplaceObjectUsageInfo(objectk10, NewObjectUsageInfo())
			Expect(cache.GetUsedObjectsFor(objectk10)).To(BeNil())
			Expect(cache.GetUsersFor(objectk11)).To(BeNil())
			Expect(cache.GetUsersFor(objectk12)).To(BeNil())
		})
		It("query two relations", func() {
			info := NewObjectUsageInfo("uses", objectk11, "owner", objectk12)
			cache.ReplaceObjectUsageInfo(objectk10, info)
			cache.ReplaceObjectUsageInfo(objectk10, NewObjectUsageInfo())
			Expect(cache.GetUsedObjectsForRelation(objectk10, "uses")).To(BeNil())
			Expect(cache.GetUsedObjectsForRelation(objectk10, "owner")).To(BeNil())
			Expect(cache.GetUsersForRelation(objectk11, "uses")).To(BeNil())
			Expect(cache.GetUsersForRelation(objectk11, "owner")).To(BeNil())
			Expect(cache.GetUsersForRelation(objectk12, "uses")).To(BeNil())
			Expect(cache.GetUsersForRelation(objectk12, "owner")).To(BeNil())
		})
		It("simple delete", func() {
			info := NewObjectUsageInfo("uses", objectk11, "owner", objectk12)
			cache.ReplaceObjectUsageInfo(objectk10, info)
			cache.DeleteObject(objectk10)
			Expect(cache.GetUsedObjectsFor(objectk10)).To(BeNil())
			Expect(cache.GetUsersFor(objectk11)).To(BeNil())
			Expect(cache.GetUsersFor(objectk12)).To(BeNil())
		})
		It("delete used", func() {
			info := NewObjectUsageInfo("uses", objectk11, "owner", objectk12)
			cache.ReplaceObjectUsageInfo(objectk10, info)
			cache.DeleteObject(objectk11)
			Expect(cache.GetUsedObjectsFor(objectk10)).To(Equal(utils.NewObjectIds(objectk11, objectk12)))
			Expect(cache.GetUsersFor(objectk11)).To(Equal(utils.NewObjectIds(objectk10)))
		})
	})

	Context("replacing", func() {
		It("simple usage", func() {
			info := NewObjectUsageInfo("uses", objectk11)
			cache.ReplaceObjectUsageInfo(objectk10, info)
			cache.ReplaceObjectUsageInfo(objectk10, NewObjectUsageInfo("owner", objectk21))
			Expect(cache.GetUsedObjectsFor(objectk10)).To(Equal(utils.NewObjectIds(objectk21)))
			Expect(cache.GetUsersFor(objectk11)).To(BeNil())
			Expect(cache.GetUsersFor(objectk21)).To(Equal(utils.NewObjectIds(objectk10)))
		})
		It("two relations", func() {
			info := NewObjectUsageInfo("uses", objectk11, "owner", objectk12)
			cache.ReplaceObjectUsageInfo(objectk10, info)
			cache.ReplaceObjectUsageInfo(objectk10, NewObjectUsageInfo("uses", objectk11, "owner", objectk21))
			Expect(cache.GetUsedObjectsFor(objectk10)).To(Equal(utils.NewObjectIds(objectk11, objectk21)))
			Expect(cache.GetUsersFor(objectk11)).To(Equal(utils.NewObjectIds(objectk10)))
			Expect(cache.GetUsersFor(objectk12)).To(BeNil())
			Expect(cache.GetUsersFor(objectk21)).To(Equal(utils.NewObjectIds(objectk10)))
		})
	})

	Context("GroupKind update", func() {
		It("simple usage", func() {
			info := NewObjectUsageInfo("uses", objectk11)
			cache.ReplaceObjectUsageInfo(objectk10, info)
			cache.ReplaceObjectUsageInfoForGKs(objectk10, utils.NewGroupKinds(gk2), NewObjectUsageInfo("uses", objectk21))
			Expect(cache.GetUsedObjectsFor(objectk10)).To(Equal(utils.NewObjectIds(objectk11, objectk21)))
			Expect(cache.GetUsersFor(objectk11)).To(Equal(utils.NewObjectIds(objectk10)))
			Expect(cache.GetUsersFor(objectk21)).To(Equal(utils.NewObjectIds(objectk10)))
		})
		It("two relations", func() {
			info := NewObjectUsageInfo("uses", objectk11, "owner", objectk12)
			cache.ReplaceObjectUsageInfo(objectk10, info)
			cache.ReplaceObjectUsageInfoForGKs(objectk10, utils.NewGroupKinds(gk2), NewObjectUsageInfo("owner", objectk21))
			Expect(cache.GetUsedObjectsFor(objectk10)).To(Equal(utils.NewObjectIds(objectk11, objectk21, objectk12)))
			Expect(cache.GetUsersFor(objectk11)).To(Equal(utils.NewObjectIds(objectk10)))
			Expect(cache.GetUsersFor(objectk12)).To(Equal(utils.NewObjectIds(objectk10)))
			Expect(cache.GetUsersFor(objectk21)).To(Equal(utils.NewObjectIds(objectk10)))
		})
	})

	Context("Detecting cycles", func() {
		It("Should pass when no cycle", func() {
			cache.ReplaceObjectUsageInfo(objectk10, NewObjectUsageInfo("uses", objectk11))
			cache.ReplaceObjectUsageInfo(objectk11, NewObjectUsageInfo("uses", objectk12))
			Expect(cache.IsCyclicForRelationForGK(objectk11, "uses", gk1)).Should(BeNil())
		})
		It("Should detect simple cycle", func() {
			cache.ReplaceObjectUsageInfo(objectk10, NewObjectUsageInfo("uses", objectk11))
			cache.ReplaceObjectUsageInfo(objectk11, NewObjectUsageInfo("uses", objectk12))
			cache.ReplaceObjectUsageInfo(objectk12, NewObjectUsageInfo("uses", objectk10))
			Expect(cache.IsCyclicForRelationForGK(objectk10, "uses", gk1)).Should(Equal([]utils.ObjectId{
				objectk10, objectk11, objectk12, objectk10,
			}))
		})
		It("Should not detect cycle", func() {
			cache.ReplaceObjectUsageInfo(objectk10, NewObjectUsageInfo("uses", objectk11))
			cache.ReplaceObjectUsageInfo(objectk11, NewObjectUsageInfo("uses", objectk12))
			cache.ReplaceObjectUsageInfo(objectk11, NewObjectUsageInfo("uses", objectk13))
			Expect(cache.IsCyclicForRelationForGK(objectk10, "uses", gk1)).Should(BeNil())
		})
	})
})
