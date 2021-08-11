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
	"github.com/go-logr/logr"
	"sync"
)

type Ready struct {
	log     logr.Logger
	name    string
	count   int
	lock    sync.Mutex
	waiters []*sync.Mutex
}

func NewReady(log logr.Logger, name string) *Ready {
	return &Ready{
		log:  log,
		name: name,
	}
}

func (r *Ready) Add() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.count++
}

func (r *Ready) IsReady() bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.count == 0
}

func (r *Ready) Wait() {
	r.lock.Lock()
	if r.count == 0 {
		r.lock.Unlock()
		return
	}
	if r.log != nil {
		r.log.Info(fmt.Sprintf("waiting for %s to be finished (%d resources pending)", r.name, r.count))
	}
	lock := &sync.Mutex{}
	lock.Lock()
	r.waiters = append(r.waiters, lock)
	r.lock.Unlock()
	lock.Lock()
}

func (r *Ready) Remove() {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.count > 0 {
		r.count--
	} else {
		panic("too many removes")
	}
	if r.count == 0 {
		if r.log != nil {
			r.log.Info(fmt.Sprintf("%s done -> wakeup %d waiters", r.name, len(r.waiters)))
		}
		for _, l := range r.waiters {
			l.Unlock()
		}
		r.waiters = nil
	}
}
