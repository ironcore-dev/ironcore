// Copyright 2023 OnMetal authors
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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ResourceName is the name of a resource, most often used alongside a resource.Quantity.
type ResourceName string

const (
	// ResourceCPU is the amount of cpu in cores.
	ResourceCPU ResourceName = "cpu"
	// ResourceMemory is the amount of memory in bytes.
	ResourceMemory ResourceName = "memory"
	// ResourceStorage is the amount of storage, in bytes.
	ResourceStorage ResourceName = "storage"
	// ResourceTPS defines max throughput per second. (e.g. 1Gi)
	ResourceTPS ResourceName = "tps"
	// ResourceIOPS defines max IOPS in input/output operations per second.
	ResourceIOPS ResourceName = "iops"

	// ResourcesRequestsPrefix is the prefix used for limiting resource requests in ResourceQuota.
	ResourcesRequestsPrefix = "requests."

	// SharedResourcesPrefix is the prefix used for the shared (virtual) resources.
	SharedResourcesPrefix = "shared-"

	// ResourceSharedCPU is the amount of virtual cpu in cores.
	ResourceSharedCPU ResourceName = SharedResourcesPrefix + ResourceCPU
	// ResourceSharedMemory is the amount of virtual memory in bytes.
	ResourceSharedMemory ResourceName = SharedResourcesPrefix + ResourceMemory

	// ResourceRequestsCPU is the amount of requested cpu in cores.
	ResourceRequestsCPU = ResourcesRequestsPrefix + ResourceCPU
	// ResourceRequestsMemory is the amount of requested memory in bytes.
	ResourceRequestsMemory = ResourcesRequestsPrefix + ResourceMemory
	// ResourceRequestsStorage is the amount of requested storage in bytes.
	ResourceRequestsStorage = ResourcesRequestsPrefix + ResourceStorage

	// ResourceCountNamespacePrefix is resource namespace prefix for counting resources.
	ResourceCountNamespacePrefix = "count/"
)

// ObjectCountQuotaResourceNameFor returns the ResourceName for counting the given groupResource.
func ObjectCountQuotaResourceNameFor(groupResource schema.GroupResource) ResourceName {
	if len(groupResource.Group) == 0 {
		return ResourceName("count/" + groupResource.Resource)
	}
	return ResourceName(ResourceCountNamespacePrefix + groupResource.Resource + "." + groupResource.Group)
}

// ResourceList is a list of ResourceName alongside their resource.Quantity.
type ResourceList map[ResourceName]resource.Quantity

// Name returns the resource with name if specified, otherwise it returns a nil quantity with default format.
func (rl *ResourceList) Name(name ResourceName, defaultFormat resource.Format) *resource.Quantity {
	if val, ok := (*rl)[name]; ok {
		return &val
	}
	return &resource.Quantity{Format: defaultFormat}
}

// Storage is a shorthand for getting the quantity associated with ResourceStorage.
func (rl *ResourceList) Storage() *resource.Quantity {
	return rl.Name(ResourceStorage, resource.BinarySI)
}

// Memory is a shorthand for getting the quantity associated with ResourceMemory.
func (rl *ResourceList) Memory() *resource.Quantity {
	return rl.Name(ResourceMemory, resource.BinarySI)
}

// CPU is a shorthand for getting the quantity associated with ResourceCPU.
func (rl *ResourceList) CPU() *resource.Quantity {
	return rl.Name(ResourceCPU, resource.DecimalSI)
}

// TPS is a shorthand for getting the quantity associated with ResourceTPS.
func (rl *ResourceList) TPS() *resource.Quantity {
	return rl.Name(ResourceTPS, resource.DecimalSI)
}

// IOPS is a shorthand for getting the quantity associated with ResourceIOPS.
func (rl *ResourceList) IOPS() *resource.Quantity {
	return rl.Name(ResourceIOPS, resource.DecimalSI)
}
