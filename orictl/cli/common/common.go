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

package common

import (
	"io"
	"os"

	"github.com/onmetal/onmetal-api/orictl/renderer"
	"github.com/spf13/pflag"
)

type Streams struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

var OSStreams = Streams{
	In:  os.Stdin,
	Out: os.Stdout,
	Err: os.Stderr,
}

const ReaderIdent = "-"

func ReadFileOrReader(filename string, orStream io.Reader) ([]byte, error) {
	if filename == ReaderIdent {
		return io.ReadAll(orStream)
	}
	return os.ReadFile(filename)
}

type OutputOptions struct {
	Registry *renderer.Registry
	Output   string
}

func (o *OutputOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Output, "output", "o", o.Output, "Output format.")
}

func (o *OutputOptions) Renderer(ifEmpty string) (renderer.Renderer, error) {
	output := o.Output
	if output == "" {
		output = ifEmpty
	}
	return o.Registry.Get(output)
}

func (o *OutputOptions) RendererOrNil() (renderer.Renderer, error) {
	output := o.Output
	if output == "" {
		return nil, nil
	}
	return o.Registry.Get(output)
}

type OutputOptionsFactory struct {
	registry *renderer.Registry
}

func (f *OutputOptionsFactory) NewOutputOptions() *OutputOptions {
	return &OutputOptions{
		Registry: f.registry,
	}
}
