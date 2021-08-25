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
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ObjectUsageInfo map[string]utils.ObjectIds

func NewObjectUsageInfo(config ...interface{}) ObjectUsageInfo {
	relation := ""
	info := ObjectUsageInfo{}
	for _, c := range config {
		switch v := c.(type) {
		case string:
			relation = v
		case utils.ObjectId:
			if relation == "" {
				panic("invalid UsageInfo spec")
			}
			old := info[relation]
			if old == nil {
				old = utils.ObjectIds{}
				info[relation] = old
			}
			old.Add(v)
		case utils.ObjectIds:
			if relation == "" {
				panic("invalid UsageInfo spec")
			}
			old := info[relation]
			if old == nil {
				info[relation] = v
			} else {
				old.AddAll(v)
			}
		}
	}
	return info
}

func (i ObjectUsageInfo) Add(relation string, ids utils.ObjectIds) {
	if len(ids) == 0 {
		return
	}
	old := i[relation]
	if old == nil {
		old = utils.ObjectIds{}
		i[relation] = old
	}
	old.AddAll(ids)
}

func (i ObjectUsageInfo) Equal(other ObjectUsageInfo) bool {
	if len(i) == 0 && len(other) == 0 {
		return true
	}
	if len(i) != len(other) {
		return false
	}
	for relation, ids := range other {
		if !ids.Equal(i[relation]) {
			return false
		}
	}
	for relation := range i {
		if _, ok := other[relation]; !ok {
			return false
		}

	}
	return true
}

func (i ObjectUsageInfo) GetObjectIdsForGK(gk schema.GroupKind) utils.ObjectIds {
	if len(i) == 0 {
		return nil
	}
	oids := utils.ObjectIds{}
	for _, ids := range i {
		for id := range ids {
			if id.GroupKind == gk {
				oids.Add(id)
			}
		}
	}
	if len(oids) == 0 {
		return nil
	}
	return oids
}

func (i ObjectUsageInfo) GetObjectIdsForRelationToGK(relation string, gk schema.GroupKind) utils.ObjectIds {
	if len(i) == 0 {
		return nil
	}
	oids := utils.ObjectIds{}
	ids := i[relation]
	if len(ids) == 0 {
		return nil
	}
	for id := range ids {
		if id.GroupKind == gk {
			oids.Add(id)
		}
	}
	return oids
}

func (i ObjectUsageInfo) GetObjectIdsForRelation(relation string) utils.ObjectIds {
	if len(i) == 0 {
		return nil
	}
	return i[relation].Copy()
}

func (i ObjectUsageInfo) GetObjectIds() utils.ObjectIds {
	if len(i) == 0 {
		return nil
	}
	oids := utils.ObjectIds{}
	for _, ids := range i {
		oids.AddAll(ids)
	}
	if len(oids) == 0 {
		return nil
	}
	return oids
}
