// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

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
