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

// FakeVolumes implements VolumeInterface
type FakeVolumes struct {
	Fake *FakeStorage
	ns   string
}

var volumesResource = schema.GroupVersionResource{Group: "storage.onmetal.de", Version: "", Resource: "volumes"}

var volumesKind = schema.GroupVersionKind{Group: "storage.onmetal.de", Version: "", Kind: "Volume"}

// Get takes name of the volume, and returns the corresponding volume object, and an error if there is any.
func (c *FakeVolumes) Get(ctx context.Context, name string, options v1.GetOptions) (result *storage.Volume, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(volumesResource, c.ns, name), &storage.Volume{})

	if obj == nil {
		return nil, err
	}
	return obj.(*storage.Volume), err
}

// List takes label and field selectors, and returns the list of Volumes that match those selectors.
func (c *FakeVolumes) List(ctx context.Context, opts v1.ListOptions) (result *storage.VolumeList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(volumesResource, volumesKind, c.ns, opts), &storage.VolumeList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &storage.VolumeList{ListMeta: obj.(*storage.VolumeList).ListMeta}
	for _, item := range obj.(*storage.VolumeList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested volumes.
func (c *FakeVolumes) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(volumesResource, c.ns, opts))

}

// Create takes the representation of a volume and creates it.  Returns the server's representation of the volume, and an error, if there is any.
func (c *FakeVolumes) Create(ctx context.Context, volume *storage.Volume, opts v1.CreateOptions) (result *storage.Volume, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(volumesResource, c.ns, volume), &storage.Volume{})

	if obj == nil {
		return nil, err
	}
	return obj.(*storage.Volume), err
}

// Update takes the representation of a volume and updates it. Returns the server's representation of the volume, and an error, if there is any.
func (c *FakeVolumes) Update(ctx context.Context, volume *storage.Volume, opts v1.UpdateOptions) (result *storage.Volume, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(volumesResource, c.ns, volume), &storage.Volume{})

	if obj == nil {
		return nil, err
	}
	return obj.(*storage.Volume), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeVolumes) UpdateStatus(ctx context.Context, volume *storage.Volume, opts v1.UpdateOptions) (*storage.Volume, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(volumesResource, "status", c.ns, volume), &storage.Volume{})

	if obj == nil {
		return nil, err
	}
	return obj.(*storage.Volume), err
}

// Delete takes name of the volume and deletes it. Returns an error if one occurs.
func (c *FakeVolumes) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(volumesResource, c.ns, name, opts), &storage.Volume{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVolumes) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(volumesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &storage.VolumeList{})
	return err
}

// Patch applies the patch and returns the patched volume.
func (c *FakeVolumes) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *storage.Volume, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(volumesResource, c.ns, name, pt, data, subresources...), &storage.Volume{})

	if obj == nil {
		return nil, err
	}
	return obj.(*storage.Volume), err
}
