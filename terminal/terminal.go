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

package terminal

import (
	"io"

	"k8s.io/client-go/tools/remotecommand"
)

// Terminal is a terminal to run.
type Terminal interface {
	// Run runs the terminal with the given streams. At least one of in, out and/or err has to be specified.
	// Depending on the underlying implementation some streams might have to be omitted / set to nil
	// (e.g. in case of a tty-terminal, err has to be nil).
	Run(in io.Reader, out, err io.WriteCloser, resize <-chan remotecommand.TerminalSize) error
}
