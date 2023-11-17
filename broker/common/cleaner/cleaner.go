// Copyright 2022 IronCore authors
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

package cleaner

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Cleaner is an entity that allows tracking and executing cleanup methods.
type Cleaner struct {
	funcs []func(ctx context.Context) error
}

// New instantiates a new Cleaner.
func New() *Cleaner {
	return &Cleaner{}
}

func CleanupObject(c client.Client, object client.Object) func(ctx context.Context) error {
	// Create a copy so subsequent writes cannot change the object.
	object = object.DeepCopyObject().(client.Object)
	return func(ctx context.Context) error {
		if err := c.Delete(ctx, object); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error deleting object: %w", err)
		}
		return nil
	}
}

func CleanupOnError(ctx context.Context, c *Cleaner, err *error) {
	log := ctrl.LoggerFrom(ctx)

	if *err == nil {
		return
	}

	if err := c.Cleanup(ctx); err != nil {
		log.Error(err, "Error cleaning up")
	}
}

// Add adds the given function to the cleanup stack.
//
// Cleanup functions will be executed in reverse order.
func (c *Cleaner) Add(f func(ctx context.Context) error) {
	// funcs need to be added in reverse order (cleanup stack)
	c.funcs = append([]func(ctx context.Context) error{f}, c.funcs...)
}

// Cleanup runs all cleanup functions with the given context, existing on the first error occurred.
func (c *Cleaner) Cleanup(ctx context.Context) error {
	for _, f := range c.funcs {
		if err := f(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Reset removes all cleanup functions from the cleaner.
func (c *Cleaner) Reset() {
	c.funcs = nil
}
