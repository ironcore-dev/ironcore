// Copyright 2022 OnMetal authors
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

package transaction

import (
	"errors"
	"fmt"
	"sync"
)

var (
	// ErrClosed is an error that is returned when a transaction is closed.
	ErrClosed = errors.New("transaction closed")
)

func IgnoreClosedError(err error) error {
	if errors.Is(err, ErrClosed) {
		return nil
	}
	return err
}

// Transaction allows building 'transaction-like' function complexes around shared objects.
//
// Commit should be called if a relation should be 'persisted' (whatever makes sense here for the relation)
// and the given id should be written as dependent. If Commit fails, a Transaction is expected to automatically
// roll back so calling Rollback is not required.
//
// Rollback should be called if relation should not be 'persisted' and any changes related should be rolled back.
//
// It should only be possible to call Commit or Rollback once. After that, ErrClosed should be returned.
type Transaction[E any] interface {
	Commit(value E) error
	Rollback() error
}

type Callback[E any] func() (value E, rollback bool)

// Build builds a transaction, allowing the caller to linearly implement the transaction function f.
//
// Build waits for the function to call the supplied Callback. If it returns an error before that,
// the construction of the transaction is considered to be failed and the error is returned as a result of Build.
// Otherwise, a Transaction object will be created that returns the values grabbed by a user either calling
// Transaction.Commit or Transaction.Rollback.
func Build[E any](f func(callback Callback[E]) error) (Transaction[E], error) {
	var (
		supply   = make(chan valueOrCallback[E])
		called   = make(chan struct{})
		once     sync.Once
		value    E
		rollback bool
	)
	callback := func() (E, bool) {
		once.Do(func() {
			close(called)
			r := <-supply
			value = r.value
			rollback = r.rollback
		})
		return value, rollback
	}

	res := make(chan error)
	go func() {
		res <- f(callback)
	}()

	select {
	case <-called:
		return &transactor[E]{
			supply: supply,
			res:    res,
		}, nil
	case err := <-res:
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("building the transaction exited prematurely")
	}
}

type valueOrCallback[E any] struct {
	value    E
	rollback bool
}

type transactor[E any] struct {
	mu sync.Mutex

	closed bool

	supply chan valueOrCallback[E]
	res    chan error
}

func (t *transactor[E]) Commit(value E) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return ErrClosed
	}
	defer func() {
		t.closed = true
	}()

	t.supply <- valueOrCallback[E]{value: value}

	return <-t.res
}

func (t *transactor[E]) Rollback() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return ErrClosed
	}

	t.supply <- valueOrCallback[E]{rollback: true}

	return <-t.res
}
