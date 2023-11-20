// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package sync

import (
	"fmt"
	"sync"
)

// MutexMap is a map of mutexes by a given key type [K].
//
// The map keeps track of entities trying to lock a certain key. Upon releasing a key,
// if no entity wants to lock a key, the key is deleted from the map, keeping it as small as possible.
type MutexMap[K comparable] struct {
	mu      sync.RWMutex
	entries map[K]*mutexMapEntry
}

// NewMutexMap creates a new MutexMap.
func NewMutexMap[K comparable]() *MutexMap[K] {
	return &MutexMap[K]{
		entries: make(map[K]*mutexMapEntry),
	}
}

type mutexMapEntry struct {
	mu    sync.Mutex
	count int
}

// Lock locks the given key.
func (m *MutexMap[K]) Lock(key K) {
	m.mu.Lock()
	entry := m.entries[key]
	if entry == nil {
		entry = &mutexMapEntry{}
		m.entries[key] = entry
	}
	entry.count++
	m.mu.Unlock()

	entry.mu.Lock()
}

// Unlock unlocks the given key.
func (m *MutexMap[K]) Unlock(key K) {
	m.mu.Lock()
	entry := m.entries[key]
	if entry == nil {
		m.mu.Unlock()
		panic(fmt.Errorf("unlock: key %v not found", key))
	}

	entry.count--
	if entry.count == 0 {
		delete(m.entries, key)
	}
	m.mu.Unlock()
	entry.mu.Unlock()
}

type mutexMapLocker[K comparable] struct {
	m   *MutexMap[K]
	key K
}

func (m *mutexMapLocker[K]) Lock() {
	m.m.Lock(m.key)
}

func (m *mutexMapLocker[K]) Unlock() {
	m.m.Unlock(m.key)
}

// Locker returns a sync.Locker that locks / unlocks the given key in the MutexMap.
func (m *MutexMap[K]) Locker(key K) sync.Locker {
	return &mutexMapLocker[K]{
		m:   m,
		key: key,
	}
}

// Len returns the number of entries in the MutexMap.
func (m *MutexMap[K]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.entries)
}

// Count returns the number of locking entities for a given key.
// If a key is not present, 0 is returned.
func (m *MutexMap[K]) Count(key K) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry := m.entries[key]
	if entry == nil {
		return 0
	}
	return entry.count
}
