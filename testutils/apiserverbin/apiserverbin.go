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

// Package apiserverbin is a test-only package that provides the path to a
// compiled binary of the onmetal-api API server. This is to speed up tests
// by not requiring the compilation each time.
package apiserverbin

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

var (
	Path string
)

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("apiserverbin: unable to determine filename")
	}

	Path = filepath.Join(filename, "..", "..", "..", "testbin", "apiserver")

	var out bytes.Buffer
	cmd := exec.Command("go", "build", "-o",
		Path,
		filepath.Join(filename, "..", "..", "..", "cmd", "apiserver"),
	)
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		panic(fmt.Errorf("error running command: %w\noutput: %s", err, out.String()))
	}
}
