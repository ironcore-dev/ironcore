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
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type GroupKinds map[schema.GroupKind]struct{}

func NewGroupKinds(ids ...schema.GroupKind) GroupKinds {
	set := GroupKinds{}
	for _, id := range ids {
		set.Add(id)
	}
	return set
}

func (o GroupKinds) Add(id schema.GroupKind) {
	o[id] = struct{}{}
}

func (o GroupKinds) AddAll(ids GroupKinds) {
	for id := range ids {
		o.Add(id)
	}
}

func (o GroupKinds) Remove(id schema.GroupKind) {
	delete(o, id)
}

func (o GroupKinds) Copy() GroupKinds {
	if o == nil {
		return nil
	}
	new := GroupKinds{}
	for id := range o {
		new.Add(id)
	}
	return new
}

func (o GroupKinds) String() string {
	s := "["
	sep := ""
	for id := range o {
		s = fmt.Sprintf("%s%s%s", s, sep, id)
		sep = ","
	}
	return s + "]"
}

func (o GroupKinds) Equal(ids GroupKinds) bool {
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

func (o GroupKinds) Join(ids GroupKinds) GroupKinds {
	new := o.Copy()
	for id := range ids {
		new.Add(id)
	}
	return new
}

func (o GroupKinds) Contains(id schema.GroupKind) bool {
	if o == nil {
		return false
	}
	_, ok := o[id]
	return ok
}
