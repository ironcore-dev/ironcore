// Copyright 2023 IronCore authors
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

package client

import (
	"context"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func gvkForList(list client.ObjectList, scheme *runtime.Scheme) (schema.GroupVersionKind, error) {
	gvk, err := apiutil.GVKForObject(list, scheme)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}

	gvk.Kind = strings.TrimSuffix(gvk.Kind, "List")
	return gvk, nil
}

type groupKindEntry struct {
	objects map[client.ObjectKey]client.Object
}

type ReaderCache struct {
	scheme     *runtime.Scheme
	groupKinds map[schema.GroupKind]*groupKindEntry
}

func NewReaderCache(scheme *runtime.Scheme) *ReaderCache {
	return &ReaderCache{
		scheme:     scheme,
		groupKinds: make(map[schema.GroupKind]*groupKindEntry),
	}
}

func (c *ReaderCache) CanList(list client.ObjectList) (bool, error) {
	gvk, err := gvkForList(list, c.scheme)
	if err != nil {
		return false, fmt.Errorf("error getting gvk for list: %w", err)
	}

	_, ok := c.groupKinds[gvk.GroupKind()]
	return ok, nil
}

func (c *ReaderCache) entryForGVK(gk schema.GroupKind) (*groupKindEntry, error) {
	entry, ok := c.groupKinds[gk]
	if ok {
		return entry, nil
	}

	entry = &groupKindEntry{
		objects: make(map[client.ObjectKey]client.Object),
	}
	c.groupKinds[gk] = entry
	return entry, nil
}

func (c *ReaderCache) InsertList(list client.ObjectList) error {
	gvk, err := gvkForList(list, c.scheme)
	if err != nil {
		return fmt.Errorf("error getting gvk for list: %w", err)
	}

	gk := gvk.GroupKind()
	entry, err := c.entryForGVK(gk)
	if err != nil {
		return err
	}

	if err := meta.EachListItem(list, func(obj runtime.Object) error {
		cObj := obj.(client.Object)
		entry.objects[client.ObjectKeyFromObject(cObj)] = cObj
		return nil
	}); err != nil {
		return fmt.Errorf("error iterating list: %w", err)
	}

	return nil
}

func (c *ReaderCache) Insert(obj client.Object) error {
	gvk, err := apiutil.GVKForObject(obj, c.scheme)
	if err != nil {
		return fmt.Errorf("error getting gvk for object: %w", err)
	}

	gk := gvk.GroupKind()
	entry, err := c.entryForGVK(gk)
	if err != nil {
		return err
	}

	entry.objects[client.ObjectKeyFromObject(obj)] = obj
	return nil
}

func (c *ReaderCache) Get(_ context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	o := &client.GetOptions{}
	o.ApplyOptions(opts)

	gvk, err := apiutil.GVKForObject(obj, c.scheme)
	if err != nil {
		return fmt.Errorf("error getting gvk for object: %w", err)
	}

	gk := gvk.GroupKind()
	entry, ok := c.groupKinds[gk]
	if !ok {
		return apierrors.NewNotFound(schema.GroupResource{Group: gk.Group, Resource: gk.Kind}, key.Name)
	}

	found, ok := entry.objects[key]
	if !ok {
		return apierrors.NewNotFound(schema.GroupResource{Group: gk.Group, Resource: gk.Kind}, key.Name)
	}

	return c.scheme.Convert(found, obj, nil)
}

func genericFieldsFromObject(obj client.Object) fields.Set {
	set := fields.Set{
		"metadata.name": obj.GetName(),
	}
	if namespace := obj.GetNamespace(); namespace != "" {
		set["metadata.namespace"] = namespace
	}
	return set
}

func matchesListOptions(opts *client.ListOptions, obj client.Object) bool {
	if opts.Namespace != "" && obj.GetNamespace() != opts.Namespace {
		return false
	}

	if opts.FieldSelector != nil {
		if !opts.FieldSelector.Matches(genericFieldsFromObject(obj)) {
			return false
		}
	}

	if opts.LabelSelector != nil {
		if !opts.LabelSelector.Matches(labels.Set(obj.GetLabels())) {
			return false
		}
	}

	return true
}

func (c *ReaderCache) List(_ context.Context, list client.ObjectList, opts ...client.ListOption) error {
	o := &client.ListOptions{}
	o.ApplyOptions(opts)

	gvk, err := gvkForList(list, c.scheme)
	if err != nil {
		return fmt.Errorf("error getting gvk for list: %w", err)
	}

	gk := gvk.GroupKind()
	entry, ok := c.groupKinds[gk]
	if !ok {
		return nil
	}

	var (
		rObjs []runtime.Object
		gv    = gvk.GroupVersion()
	)
	for key, obj := range entry.objects {
		if o.Limit > 0 && int64(len(rObjs)) >= o.Limit {
			break
		}

		if !matchesListOptions(o, obj) {
			continue
		}

		rObj, err := c.scheme.ConvertToVersion(obj, gv)
		if err != nil {
			return fmt.Errorf("error converting object %s: %w", key, err)
		}

		rObjs = append(rObjs, rObj)
	}

	return meta.SetList(list, rObjs)
}
