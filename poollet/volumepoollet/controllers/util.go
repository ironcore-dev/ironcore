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

package controllers

import (
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type dependencyNotReadyError struct {
	group    string
	resource string
	name     string

	cause error
}

func (d *dependencyNotReadyError) Error() string {
	return fmt.Sprintf("dependency %s %s not ready: %v",
		schema.GroupResource{
			Group:    d.group,
			Resource: d.resource,
		},
		d.name,
		d.cause,
	)
}

func (d *dependencyNotReadyError) Unwrap() error {
	return d.cause
}

func IsDependencyNotReadyError(err error) bool {
	return errors.As(err, new(*dependencyNotReadyError))
}

func IgnoreDependencyNotReadyError(err error) error {
	if IsDependencyNotReadyError(err) {
		return nil
	}
	return err
}

func NewDependencyNotReadyError(gr schema.GroupResource, name string, cause error) error {
	return &dependencyNotReadyError{
		group:    gr.Group,
		resource: gr.Resource,
		name:     name,
		cause:    cause,
	}
}
