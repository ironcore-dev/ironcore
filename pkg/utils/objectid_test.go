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

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ObjectId", func() {
	gk := schema.GroupKind{
		Group: "g1",
		Kind:  "k1",
	}
	oid1g1k1 := ObjectId{
		ObjectKey: client.ObjectKey{
			Namespace: "default",
			Name:      "o1",
		},
		GroupKind: schema.GroupKind{
			Group: "g1",
			Kind:  "k1",
		},
	}
	oid2k1 := ObjectId{
		ObjectKey: client.ObjectKey{
			Namespace: "default",
			Name:      "o1",
		},
		GroupKind: schema.GroupKind{
			Kind: "k1",
		},
	}

	Context("Creation", func() {
		It("Should create an ObjectId from string", func() {
			s := "k1.g1/default/o1"
			Expect(MustParseObjectId(s)).Should(Equal(oid1g1k1))
		})
		It("Should create an ObjectId from string", func() {
			s := "k1/default/o1"
			Expect(MustParseObjectId(s)).Should(Equal(oid2k1))
		})
		It("Should create an ObjectId from string and request", func() {
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "o1",
				},
			}
			newOid := NewObjectIdForRequest(req, gk)
			Expect(newOid).Should(Equal(oid1g1k1))
		})
	})
})

var _ = Describe("ObjectIds", func() {
	var objectIds ObjectIds
	oid1g1k1 := ObjectId{
		ObjectKey: client.ObjectKey{
			Namespace: "default",
			Name:      "o1",
		},
		GroupKind: schema.GroupKind{
			Group: "g1",
			Kind:  "k1",
		},
	}
	oid2g1k1 := ObjectId{
		ObjectKey: client.ObjectKey{
			Namespace: "default",
			Name:      "o2",
		},
		GroupKind: schema.GroupKind{
			Group: "g1",
			Kind:  "k1",
		},
	}
	oid1g2k1 := ObjectId{
		ObjectKey: client.ObjectKey{
			Namespace: "default",
			Name:      "o1",
		},
		GroupKind: schema.GroupKind{
			Group: "g2",
			Kind:  "k1",
		},
	}

	BeforeEach(func() {
		objectIds = NewObjectIds()
	})

	Context("Adding", func() {
		It("It should add one element", func() {
			objectIds.Add(oid1g1k1)
			Expect(objectIds.Contains(oid1g1k1)).Should(BeTrue())
		})
		It("It should add multiple elements", func() {
			objectIds.Add(oid1g1k1)
			objectIds.Add(oid2g1k1)
			Expect(objectIds.Contains(oid1g1k1)).Should(BeTrue())
			Expect(objectIds.Contains(oid2g1k1)).Should(BeTrue())
		})
		It("It should add multiple elements", func() {
			objectIdsTemp := NewObjectIds(oid1g1k1, oid2g1k1)
			objectIds.AddAll(objectIdsTemp)
			Expect(objectIds.Contains(oid1g1k1)).Should(BeTrue())
			Expect(objectIds.Contains(oid2g1k1)).Should(BeTrue())
		})
		It("It should join two ObjectIds", func() {
			objectIdsTemp := NewObjectIds(oid1g1k1, oid2g1k1)
			objectIds.Add(oid1g2k1)
			joined := objectIds.Join(objectIdsTemp)
			Expect(joined.Contains(oid1g1k1)).Should(BeTrue())
			Expect(joined.Contains(oid2g1k1)).Should(BeTrue())
			Expect(joined.Contains(oid1g2k1)).Should(BeTrue())
		})
	})

	Context("Compare", func() {
		It("It should compare two equal ObjectIds", func() {
			objectIds.Add(oid1g1k1)
			Expect(objectIds.Equal(NewObjectIds(oid1g1k1))).Should(BeTrue())
		})
		It("It should compare two unequal ObjectIds", func() {
			objectIds.Add(oid1g1k1)
			Expect(objectIds.Equal(NewObjectIds(oid2g1k1))).Should(BeFalse())
		})
		It("It should compare two unequal ObjectIds", func() {
			objectIds.Add(oid1g1k1)
			Expect(objectIds.Equal(NewObjectIds())).Should(BeFalse())
		})
		It("It should compare two empty ObjectIds", func() {
			Expect(objectIds.Equal(NewObjectIds())).Should(BeTrue())
		})
	})

	Context("Output", func() {
		It("It should render a correct string", func() {
			objectIds.Add(oid1g1k1)
			output := "[k1.g1/default/o1]"
			Expect(objectIds.String()).Should(Equal(output))
		})
		It("It should render a correct string for multiple ObjectIds", func() {
			objectIds.Add(oid1g1k1)
			objectIds.Add(oid1g2k1)
			output1 := "[k1.g1/default/o1,k1.g2/default/o1]"
			output2 := "[k1.g2/default/o1,k1.g1/default/o1]"
			Expect(objectIds.String()).Should(Or(Equal(output1), Equal(output2)))
		})
	})

	Context("Remove", func() {
		It("It should remove a single element", func() {
			objectIds.Add(oid1g1k1)
			objectIds.Add(oid2g1k1)
			Expect(objectIds.Contains(oid1g1k1)).Should(BeTrue())
			Expect(objectIds.Contains(oid2g1k1)).Should(BeTrue())
			objectIds.Remove(oid2g1k1)
			Expect(objectIds.Contains(oid2g1k1)).Should(BeFalse())
		})
	})
})
