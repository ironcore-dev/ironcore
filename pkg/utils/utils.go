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
	"context"
	"fmt"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AssureDeleting ensures that an object is being deleted
func AssureDeleting(ctx context.Context, clt client.Client, object client.Object) error {
	log := logr.FromContextOrDiscard(ctx)
	if !object.GetDeletionTimestamp().IsZero() {
		return nil
	}
	log.Info("deleting object", "objectID", NewObjectId(object))
	return client.IgnoreNotFound(clt.Delete(ctx, object, client.PropagationPolicy(metav1.DeletePropagationBackground)))
}

// ContainsString is a helper functions to check and remove string from a slice of strings.
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// GetLabel returns the value of a given object and label name
func GetLabel(object client.Object, name string, defs ...string) string {
	def := ""
	if len(defs) > 0 {
		def = defs[0]
	}
	if object != nil && object.GetLabels() != nil {
		if found := object.GetLabels()[name]; found != "" {
			return found
		}
	}
	return def
}

func GetObjectForGroupKind(clt client.Client, gk schema.GroupKind) client.Object {
	list := clt.Scheme().VersionsForGroupKind(gk)
	if len(list) == 0 {
		return nil
	}
	for _, gv := range clt.Scheme().PreferredVersionAllGroups() {
		if gv.Group == gk.Group {
			t := clt.Scheme().KnownTypes(gv)[gk.Kind]
			return reflect.New(t).Interface().(client.Object)
		}
	}
	return nil
}

func GetObjectListForGroupKind(clt client.Client, gk schema.GroupKind) client.ObjectList {
	list := clt.Scheme().VersionsForGroupKind(gk)
	if len(list) == 0 {
		return nil
	}
	kind := gk.Kind + "List"
	for _, gv := range clt.Scheme().PreferredVersionAllGroups() {
		if gv.Group == gk.Group {
			t := clt.Scheme().KnownTypes(gv)[kind]
			return reflect.New(t).Interface().(client.ObjectList)
		}
	}
	return nil
}

type ItemListIterator struct {
	list  client.ObjectList
	index int
	size  int
	value reflect.Value
}

func MustItemListIterator(list client.ObjectList) *ItemListIterator {
	it, err := NewItemListIterator(list)
	if err != nil {
		panic(err.Error())
	}
	return it
}

func NewItemListIterator(list client.ObjectList) (*ItemListIterator, error) {
	value := reflect.ValueOf(list)
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	value = value.FieldByName("Items")
	if value.IsZero() {
		return nil, fmt.Errorf("%T: no list object", list)
	}
	if value.Kind() != reflect.Slice {
		return nil, fmt.Errorf("%T: no list object", list)
	}
	return &ItemListIterator{
		list:  list,
		index: -1,
		size:  value.Len(),
		value: value,
	}, nil
}

func (i *ItemListIterator) HasNext() bool {
	return i.index < i.size-1
}

func (i *ItemListIterator) Next() client.Object {
	if !i.HasNext() {
		return nil
	}
	i.index++
	return i.Current()
}

func (i *ItemListIterator) Current() client.Object {
	value := i.value.Index(i.index)
	return value.Addr().Interface().(client.Object)
}
