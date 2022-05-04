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
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	storage "github.com/onmetal/onmetal-api/apis/storage"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeVolumePools implements VolumePoolInterface
type FakeVolumePools struct {
	Fake *FakeStorage
}

var volumepoolsResource = schema.GroupVersionResource{Group: "storage.api.onmetal.de", Version: "", Resource: "volumepools"}

var volumepoolsKind = schema.GroupVersionKind{Group: "storage.api.onmetal.de", Version: "", Kind: "VolumePool"}

// Get takes name of the volumePool, and returns the corresponding volumePool object, and an error if there is any.
func (c *FakeVolumePools) Get(ctx context.Context, name string, options v1.GetOptions) (result *storage.VolumePool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(volumepoolsResource, name), &storage.VolumePool{})
	if obj == nil {
		return nil, err
	}
	return obj.(*storage.VolumePool), err
}

// List takes label and field selectors, and returns the list of VolumePools that match those selectors.
func (c *FakeVolumePools) List(ctx context.Context, opts v1.ListOptions) (result *storage.VolumePoolList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(volumepoolsResource, volumepoolsKind, opts), &storage.VolumePoolList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &storage.VolumePoolList{ListMeta: obj.(*storage.VolumePoolList).ListMeta}
	for _, item := range obj.(*storage.VolumePoolList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested volumePools.
func (c *FakeVolumePools) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(volumepoolsResource, opts))
}

// Create takes the representation of a volumePool and creates it.  Returns the server's representation of the volumePool, and an error, if there is any.
func (c *FakeVolumePools) Create(ctx context.Context, volumePool *storage.VolumePool, opts v1.CreateOptions) (result *storage.VolumePool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(volumepoolsResource, volumePool), &storage.VolumePool{})
	if obj == nil {
		return nil, err
	}
	return obj.(*storage.VolumePool), err
}

// Update takes the representation of a volumePool and updates it. Returns the server's representation of the volumePool, and an error, if there is any.
func (c *FakeVolumePools) Update(ctx context.Context, volumePool *storage.VolumePool, opts v1.UpdateOptions) (result *storage.VolumePool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(volumepoolsResource, volumePool), &storage.VolumePool{})
	if obj == nil {
		return nil, err
	}
	return obj.(*storage.VolumePool), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeVolumePools) UpdateStatus(ctx context.Context, volumePool *storage.VolumePool, opts v1.UpdateOptions) (*storage.VolumePool, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(volumepoolsResource, "status", volumePool), &storage.VolumePool{})
	if obj == nil {
		return nil, err
	}
	return obj.(*storage.VolumePool), err
}

// Delete takes name of the volumePool and deletes it. Returns an error if one occurs.
func (c *FakeVolumePools) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(volumepoolsResource, name, opts), &storage.VolumePool{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVolumePools) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(volumepoolsResource, listOpts)

	_, err := c.Fake.Invokes(action, &storage.VolumePoolList{})
	return err
}

// Patch applies the patch and returns the patched volumePool.
func (c *FakeVolumePools) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *storage.VolumePool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(volumepoolsResource, name, pt, data, subresources...), &storage.VolumePool{})
	if obj == nil {
		return nil, err
	}
	return obj.(*storage.VolumePool), err
}
