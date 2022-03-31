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

	v1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeMachineClasses implements MachineClassInterface
type FakeMachineClasses struct {
	Fake *FakeComputeV1alpha1
	ns   string
}

var machineclassesResource = schema.GroupVersionResource{Group: "compute.onmetal.de", Version: "v1alpha1", Resource: "machineclasses"}

var machineclassesKind = schema.GroupVersionKind{Group: "compute.onmetal.de", Version: "v1alpha1", Kind: "MachineClass"}

// Get takes name of the machineClass, and returns the corresponding machineClass object, and an error if there is any.
func (c *FakeMachineClasses) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.MachineClass, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(machineclassesResource, c.ns, name), &v1alpha1.MachineClass{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MachineClass), err
}

// List takes label and field selectors, and returns the list of MachineClasses that match those selectors.
func (c *FakeMachineClasses) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.MachineClassList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(machineclassesResource, machineclassesKind, c.ns, opts), &v1alpha1.MachineClassList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.MachineClassList{ListMeta: obj.(*v1alpha1.MachineClassList).ListMeta}
	for _, item := range obj.(*v1alpha1.MachineClassList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested machineClasses.
func (c *FakeMachineClasses) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(machineclassesResource, c.ns, opts))

}

// Create takes the representation of a machineClass and creates it.  Returns the server's representation of the machineClass, and an error, if there is any.
func (c *FakeMachineClasses) Create(ctx context.Context, machineClass *v1alpha1.MachineClass, opts v1.CreateOptions) (result *v1alpha1.MachineClass, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(machineclassesResource, c.ns, machineClass), &v1alpha1.MachineClass{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MachineClass), err
}

// Update takes the representation of a machineClass and updates it. Returns the server's representation of the machineClass, and an error, if there is any.
func (c *FakeMachineClasses) Update(ctx context.Context, machineClass *v1alpha1.MachineClass, opts v1.UpdateOptions) (result *v1alpha1.MachineClass, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(machineclassesResource, c.ns, machineClass), &v1alpha1.MachineClass{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MachineClass), err
}

// Delete takes name of the machineClass and deletes it. Returns an error if one occurs.
func (c *FakeMachineClasses) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(machineclassesResource, c.ns, name, opts), &v1alpha1.MachineClass{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMachineClasses) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(machineclassesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.MachineClassList{})
	return err
}

// Patch applies the patch and returns the patched machineClass.
func (c *FakeMachineClasses) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.MachineClass, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(machineclassesResource, c.ns, name, pt, data, subresources...), &v1alpha1.MachineClass{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MachineClass), err
}
