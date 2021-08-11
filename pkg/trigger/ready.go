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

package trigger

import "sync"

type Ready struct {
	count   int
	lock    sync.Mutex
	waiters []*sync.Mutex
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
		for _, l := range r.waiters {
			l.Unlock()
		}
		r.waiters = nil
	}
}
