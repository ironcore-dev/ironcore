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
	"context"
	"fmt"
	"io"
	"os"
	"time"

	ori "github.com/onmetal/onmetal-api/ori/apis/compute/v1alpha1"
	"github.com/onmetal/onmetal-api/orictl/renderer"
	"github.com/onmetal/onmetal-api/orictl/renderer/renderers"
	"github.com/onmetal/onmetal-api/orictl/table/tableconverter"
	"github.com/onmetal/onmetal-api/orictl/table/tableconverters"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

type ClientFactory = func(ctx context.Context) (ori.MachineRuntimeClient, func() error, error)

type DialOptions struct {
	Address     string
	DialTimeout time.Duration
}

func (o *DialOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Address, "address", "", "Address to the ori server.")
	fs.DurationVar(&o.DialTimeout, "dial-timeout", 1*time.Second, "Timeout while dialing")
}

func (o *DialOptions) Dial(ctx context.Context) (ori.MachineRuntimeClient, func() error, error) {
	dialCtx, cancel := context.WithTimeout(ctx, o.DialTimeout)
	defer cancel()

	if o.Address == "" {
		o.Address = os.Getenv("ORI_MACHINE_RUNTIME")
	}

	conn, err := grpc.DialContext(dialCtx, o.Address,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithReturnConnectionError(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("error dialing: %w", err)
	}

	return ori.NewMachineRuntimeClient(conn), conn.Close, nil
}

type Streams struct {
	In  io.Writer
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
	Output string
}

func (o *OutputOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Output, "output", "o", o.Output, "Output format.")
}

func (o *OutputOptions) Renderer(ifEmpty string) (renderer.Renderer, error) {
	output := o.Output
	if output == "" {
		output = ifEmpty
	}
	return rendererRegistry.Get(output)
}

var rendererRegistry = renderer.NewRegistry()

func init() {
	tableConverter := tableconverter.NewRegistry()
	utilruntime.Must(tableconverters.AddToRegistry(tableConverter))
	utilruntime.Must(renderers.AddToRegistry(rendererRegistry))
	utilruntime.Must(rendererRegistry.Register("table", renderers.NewTable(tableConverter)))
}

var (
	MachineAliases          = []string{"machines", "mach", "machs"}
	MachineClassAliases     = []string{"machineclasses", "mc", "mcs"}
	VolumeAliases           = []string{"volumes", "vol", "vols"}
	NetworkInterfaceAliases = []string{"networkinterfaces", "nic", "nics"}
)
