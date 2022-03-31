// Copyright 2021 OnMetal authors
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

// Package tools

//go:build tools
// +build tools

package tools

import (
	// Use addlicense for adding license headers.
	_ "github.com/google/addlicense"
	// Use code-generator for generating aggregated-api code.
	_ "k8s.io/code-generator"
	// Use vgopath for setting up GOPATH to generate code with code-generator.
	_ "github.com/onmetal/vgopath"
)
