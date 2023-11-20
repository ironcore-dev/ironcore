// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	"github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

type CacheStrategy interface {
	Key(instance *v1alpha1.Volume) (types.UID, error)
	ContainerKey(instance *v1alpha1.Volume) string
}

type defaultCacheStrategy struct{}

var DefaultCacheStrategy CacheStrategy = defaultCacheStrategy{}

func (defaultCacheStrategy) Key(instance *v1alpha1.Volume) (types.UID, error) {
	uid := instance.GetUID()
	if uid == "" {
		return "", fmt.Errorf("instance has no UID")
	}
	return uid, nil
}

func (defaultCacheStrategy) ContainerKey(instance *v1alpha1.Volume) string {
	if instance.Spec.VolumePoolRef == nil {
		return ""
	}
	return instance.Spec.VolumePoolRef.Name
}

type InstanceInfo struct {
	instance *v1alpha1.Volume
}

type ContainerInfo struct {
	node      *v1alpha1.VolumePool
	instances map[types.UID]*InstanceInfo
}

func newNodeInfo() *ContainerInfo {
	return &ContainerInfo{
		instances: make(map[types.UID]*InstanceInfo),
	}
}

func (n *ContainerInfo) Node() *v1alpha1.VolumePool {
	return n.node
}

func (n *ContainerInfo) MaxAllocatable(className string) resource.Quantity {
	var assigned = resource.NewQuantity(0, resource.BinarySI)
	for _, instance := range n.instances {
		if instance.instance.Spec.VolumeClassRef != nil && instance.instance.Spec.VolumeClassRef.Name == className {
			assigned.Add(*instance.instance.Spec.Resources.Storage())
		}
	}
	allocatable, ok := n.node.Status.Allocatable[corev1alpha1.ClassCountFor(corev1alpha1.ClassTypeVolumeClass, className)]
	if !ok {
		return *resource.NewQuantity(0, resource.BinarySI)
	}
	allocatable.Sub(*assigned)

	return allocatable
}

func (n *ContainerInfo) NumInstances() int {
	return len(n.instances)
}

func (n *ContainerInfo) shallowCopy() *ContainerInfo {
	return &ContainerInfo{
		node:      n.node,
		instances: maps.Clone(n.instances),
	}
}

type instanceState struct {
	instance        *v1alpha1.Volume
	bindingFinished bool
}

func NewCache(log logr.Logger, strategy CacheStrategy) *Cache {
	return &Cache{
		log:              log,
		assumedInstances: sets.New[types.UID](),
		instanceStates:   make(map[types.UID]*instanceState),
		nodes:            make(map[string]*ContainerInfo),
		strategy:         strategy,
	}
}

type Cache struct {
	mu sync.RWMutex

	log logr.Logger

	assumedInstances sets.Set[types.UID]
	instanceStates   map[types.UID]*instanceState
	nodes            map[string]*ContainerInfo

	strategy CacheStrategy
}

type Snapshot struct {
	cache *Cache

	nodes     map[string]*ContainerInfo
	nodesList []*ContainerInfo
}

func (s *Snapshot) Update() {
	s.cache.mu.RLock()
	defer s.cache.mu.RUnlock()

	s.nodes = make(map[string]*ContainerInfo, len(s.cache.nodes))
	s.nodesList = make([]*ContainerInfo, 0, len(s.cache.nodes))
	for key, node := range s.cache.nodes {
		if node.node == nil {
			continue
		}

		node := node.shallowCopy()
		s.nodes[key] = node
		s.nodesList = append(s.nodesList, node)
	}
}

func (s *Snapshot) NumNodes() int {
	return len(s.nodesList)
}

func (s *Snapshot) ListNodes() []*ContainerInfo {
	return s.nodesList
}

func (s *Snapshot) GetNode(name string) (*ContainerInfo, error) {
	node, ok := s.nodes[name]
	if !ok {
		return nil, fmt.Errorf("node %q not found", name)
	}
	return node, nil
}

func (c *Cache) Snapshot() *Snapshot {
	snapshot := &Snapshot{cache: c}
	snapshot.Update()
	return snapshot
}

func (c *Cache) IsAssumedInstance(instance *v1alpha1.Volume) (bool, error) {
	key, err := c.strategy.Key(instance)
	if err != nil {
		return false, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.assumedInstances.Has(key), nil
}

func (c *Cache) AssumeInstance(instance *v1alpha1.Volume) error {
	log := c.log.WithValues("Instance", klog.KObj(instance))
	key, err := c.strategy.Key(instance)
	if err != nil {
		return err
	}
	log = log.WithValues("InstanceKey", key)

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.instanceStates[key]; ok {
		return fmt.Errorf("instance %s(%v) is in the cache, so can't be assumed", key, klog.KObj(instance))
	}

	c.addInstance(log, key, instance, true)
	return nil
}

func (c *Cache) ForgetInstance(instance *v1alpha1.Volume) error {
	log := c.log.WithValues("Instance", klog.KObj(instance))
	key, err := c.strategy.Key(instance)
	if err != nil {
		return err
	}
	log = log.WithValues("InstanceKey", key)

	currState, ok := c.instanceStates[key]
	if ok {
		oldContainerKey := c.strategy.ContainerKey(currState.instance)
		newContainerKey := c.strategy.ContainerKey(instance)
		if oldContainerKey != newContainerKey {
			return fmt.Errorf("instance %s(%v) was assumed on container %s but assinged to %s", key, klog.KObj(instance), newContainerKey, oldContainerKey)
		}
	}

	if ok && c.assumedInstances.Has(key) {
		c.removeInstance(log, key, instance)
	}
	return fmt.Errorf("instance %s(%v) wasn't assumed so cannot be forgotten", key, klog.KObj(instance))
}

func (c *Cache) FinishBinding(instance *v1alpha1.Volume) error {
	log := c.log.WithValues("Instance", klog.KObj(instance))
	key, err := c.strategy.Key(instance)
	if err != nil {
		return err
	}
	log = log.WithValues("InstanceKey", key)

	c.mu.RLock()
	defer c.mu.RUnlock()

	log.V(5).Info("Finished binding for instance, can be expired")
	currState, ok := c.instanceStates[key]
	if ok && c.assumedInstances.Has(key) {
		currState.bindingFinished = true
	}
	return nil
}

func (c *Cache) AddContainer(node *v1alpha1.VolumePool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	n, ok := c.nodes[node.Name]
	if !ok {
		n = newNodeInfo()
		c.nodes[node.Name] = n
	}
	n.node = node
}

func (c *Cache) UpdateContainer(_, newNode *v1alpha1.VolumePool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	n, ok := c.nodes[newNode.Name]
	if !ok {
		n = newNodeInfo()
		c.nodes[newNode.Name] = n
	}
	n.node = newNode
}

func (c *Cache) RemoveContainer(node *v1alpha1.VolumePool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	n, ok := c.nodes[node.Name]
	if !ok {
		return fmt.Errorf("node %s not found", node.Name)
	}

	n.node = nil
	if len(n.instances) == 0 {
		delete(c.nodes, node.Name)
	}
	return nil
}

func (c *Cache) AddInstance(instance *v1alpha1.Volume) error {
	log := c.log.WithValues("Instance", klog.KObj(instance))
	key, err := c.strategy.Key(instance)
	if err != nil {
		return err
	}
	log = log.WithValues("InstanceKey", key)

	c.mu.Lock()
	defer c.mu.Unlock()

	currState, ok := c.instanceStates[key]
	switch {
	case ok && c.assumedInstances.Has(key):
		// The instance was previously assumed, but now we have actual knowledge.
		c.updateInstance(log, key, currState.instance, instance)
		oldContainerKey := c.strategy.ContainerKey(currState.instance)
		newContainerKey := c.strategy.ContainerKey(instance)
		if oldContainerKey != newContainerKey {
			log.Info("Instance was added to a different container than assumed",
				"AssumedContainer", oldContainerKey,
				"ActualContainer", newContainerKey,
			)
		}
		return nil
	case !ok:
		// Instance was expired, add it back to the cache.
		c.addInstance(log, key, instance, false)
		return nil
	default:
		return fmt.Errorf("instance %s(%s) was already in added state", key, klog.KObj(instance))
	}
}

func (c *Cache) UpdateInstance(oldInstance, newInstance *v1alpha1.Volume) error {
	log := c.log.WithValues("Instance", klog.KObj(oldInstance))
	key, err := c.strategy.Key(oldInstance)
	if err != nil {
		return err
	}
	log = log.WithValues("InstanceKey", key)

	c.mu.Lock()
	defer c.mu.Unlock()

	currState, ok := c.instanceStates[key]
	if !ok {
		return fmt.Errorf("instance %s is not present in the cache and thus cannot be updated", key)
	}

	if c.assumedInstances.Has(key) {
		// An assumed instance won't have an Update / Remove event. It needs to have an Add event
		// before an Update event, in which case the state would change from assumed to added.
		return fmt.Errorf("assumed instance %s should not be updated", key)
	}

	oldContainerKey := c.strategy.ContainerKey(currState.instance)
	newContainerKey := c.strategy.ContainerKey(newInstance)
	if oldContainerKey != newContainerKey {
		// In this case, the scheduler cache is corrupted, and we cannot handle this correctly in any way - panic to
		// signal abnormal exit.
		err := fmt.Errorf("instance %s updated on container %s which is different than the container %s it was previously added to",
			key, oldContainerKey, newContainerKey)
		panic(err)
	}
	c.updateInstance(log, key, oldInstance, newInstance)
	return nil
}

func (c *Cache) RemoveInstance(instance *v1alpha1.Volume) error {
	log := c.log.WithValues("Instance", klog.KObj(instance))
	key, err := c.strategy.Key(instance)
	if err != nil {
		return err
	}
	log = log.WithValues("InstanceKey", key)

	c.mu.Lock()
	defer c.mu.Unlock()

	currState, ok := c.instanceStates[key]
	if !ok {
		return fmt.Errorf("instance %s not found", key)
	}

	oldContainerKey := c.strategy.ContainerKey(currState.instance)
	newContainerKey := c.strategy.ContainerKey(instance)
	if oldContainerKey != newContainerKey {
		// In this case, the scheduler cache is corrupted, and we cannot handle this correctly in any way - panic to
		// signal abnormal exit.
		err := fmt.Errorf("instance %s updated on container %s which is different than the container %s it was previously added to",
			key, oldContainerKey, newContainerKey)
		panic(err)
	}
	c.removeInstance(log, key, instance)
	return nil
}

func (c *Cache) updateInstance(log logr.Logger, key types.UID, oldInstance, newInstance *v1alpha1.Volume) {
	c.removeInstance(log, key, oldInstance)
	c.addInstance(log, key, newInstance, false)
}

func (c *Cache) addInstance(_ logr.Logger, key types.UID, instance *v1alpha1.Volume, assume bool) {
	containerKey := c.strategy.ContainerKey(instance)
	n, ok := c.nodes[containerKey]
	if !ok {
		n = newNodeInfo()
		c.nodes[containerKey] = n
	}
	n.instances[key] = &InstanceInfo{instance: instance}
	is := &instanceState{
		instance: instance,
	}
	c.instanceStates[key] = is
	if assume {
		c.assumedInstances.Insert(key)
	}
}

func (c *Cache) removeInstance(log logr.Logger, key types.UID, instance *v1alpha1.Volume) {
	containerKey := c.strategy.ContainerKey(instance)
	n, ok := c.nodes[containerKey]
	if !ok {
		err := fmt.Errorf("container %s not found when trying to remove instance %s", containerKey, key)
		log.Error(err, "Container not found")
	} else {
		delete(n.instances, key)
		if len(n.instances) == 0 && n.node == nil {
			// Garbage collect container if it's not used anymore.
			delete(c.nodes, containerKey)
		}
	}

	c.assumedInstances.Delete(key)
	delete(c.instanceStates, key)
}

func (c *Cache) cleanupAssumedInstances() {
	log := c.log

	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.assumedInstances {
		log := log.WithValues("InstanceKey", key)
		is, ok := c.instanceStates[key]
		if !ok {
			err := fmt.Errorf("instance key %s is assumed but no state recorded, potential logical error", key)
			panic(err)
		}

		if !is.bindingFinished {
			log.V(5).Info("Won't expire cache for an instance where binding is still in progress")
			continue
		}

		log.V(5).Info("Removing expired instance")
		c.removeInstance(log, key, is.instance)
	}
}

func (c *Cache) Start(ctx context.Context) error {
	wait.UntilWithContext(ctx, func(ctx context.Context) {
		c.cleanupAssumedInstances()
	}, 1*time.Second)
	return nil
}
