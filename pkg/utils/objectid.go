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
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectId struct {
	client.ObjectKey
	schema.GroupKind
}

func NewObjectIdForRequest(req ctrl.Request, gk schema.GroupKind) ObjectId {
	return ObjectId{
		ObjectKey: client.ObjectKey{
			Namespace: req.Namespace,
			Name:      req.Name,
		},
		GroupKind: gk,
	}
}

func MustParseObjectId(s string) ObjectId {
	id, err := ParseObjectId(s)
	if err != nil {
		panic(err)
	}
	return id
}

func ParseObjectId(s string) (ObjectId, error) {
	elems := strings.Split(s, "/")
	if len(elems) != 3 {
		return ObjectId{}, fmt.Errorf("invalid object id, required 3 elements got %d", len(elems))
	}
	key := client.ObjectKey{
		Namespace: elems[1],
		Name:      elems[2],
	}
	i := strings.Index(elems[0], ".")
	gk := schema.GroupKind{}
	if i < 0 {
		gk.Kind = elems[0]
	} else {
		gk.Kind = elems[0][0:i]
		gk.Group = elems[0][i+1:]
	}
	return ObjectId{
		ObjectKey: key,
		GroupKind: gk,
	}, nil
}

func NewObjectId(object client.Object) ObjectId {
	gvk := object.GetObjectKind().GroupVersionKind()
	return ObjectId{
		ObjectKey: client.ObjectKey{
			Namespace: object.GetNamespace(),
			Name:      object.GetName()},
		GroupKind: schema.GroupKind{
			Group: gvk.Group,
			Kind:  gvk.Kind,
		},
	}
}

func GetOwnerIdsFor(object client.Object) ObjectIds {
	ids := ObjectIds{}
	for _, o := range object.GetOwnerReferences() {
		gv, err := schema.ParseGroupVersion(o.APIVersion)
		if err == nil {
			id := ObjectId{
				ObjectKey: client.ObjectKey{
					Namespace: object.GetNamespace(),
					Name:      o.Name,
				},
				GroupKind: schema.GroupKind{
					Group: gv.Group,
					Kind:  o.Kind,
				},
			}
			ids.Add(id)
		}
	}
	return ids
}

func (o ObjectId) String() string {
	return fmt.Sprintf("%s/%s", o.GroupKind, o.ObjectKey)
}

type ObjectIds map[ObjectId]struct{}

func NewObjectIds(ids ...ObjectId) ObjectIds {
	set := ObjectIds{}
	for _, id := range ids {
		set.Add(id)
	}
	return set
}

func (o ObjectIds) Add(id ObjectId) {
	o[id] = struct{}{}
}

func (o ObjectIds) AddAll(ids ObjectIds) {
	for id := range ids {
		o.Add(id)
	}
}

func (o ObjectIds) Remove(id ObjectId) {
	delete(o, id)
}

func (o ObjectIds) Copy() ObjectIds {
	if o == nil {
		return nil
	}
	new := ObjectIds{}
	for id := range o {
		new.Add(id)
	}
	return new
}

func (o ObjectIds) String() string {
	s := "["
	sep := ""
	for id := range o {
		s = fmt.Sprintf("%s%s%s", s, sep, id)
		sep = ","
	}
	return s + "]"
}

func (o ObjectIds) Equal(ids ObjectIds) bool {
	if o == nil && ids == nil {
		return true
	}

	if len(o) != len(ids) {
		return false
	}

	for id := range ids {
		if _, ok := o[id]; !ok {
			return false
		}
	}
	return true
}

func (o ObjectIds) Join(ids ObjectIds) ObjectIds {
	new := o.Copy()
	for id := range ids {
		new.Add(id)
	}
	return new
}

func (o ObjectIds) Contains(id ObjectId) bool {
	if o == nil {
		return false
	}
	_, ok := o[id]
	return ok
}
