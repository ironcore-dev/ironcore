// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"io"
	"os"

	"github.com/ironcore-dev/ironcore/irictl/renderer"
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
