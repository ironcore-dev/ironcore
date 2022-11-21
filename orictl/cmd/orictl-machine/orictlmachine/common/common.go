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
	"fmt"
	"os"

	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"github.com/onmetal/onmetal-api/orictl/renderer"
	"github.com/onmetal/onmetal-api/orictl/renderer/machinerenderers"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientFactory interface {
	New() (ori.MachineRuntimeClient, func() error, error)
}

type ClientOptions struct {
	Address string
}

func (o *ClientOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Address, "address", "", "Address to the ori server.")
}

func (o *ClientOptions) New() (ori.MachineRuntimeClient, func() error, error) {
	address := o.Address
	if address == "" {
		address = os.Getenv("ORICTL_MACHINE_ADDRESS")
	}
	if address == "" {
		return nil, nil, fmt.Errorf("must specify address")
	}

	conn, err := grpc.Dial(o.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("error dialing: %w", err)
	}

	return ori.NewMachineRuntimeClient(conn), conn.Close, nil
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
	return machinerenderers.Registry.Get(output)
}

var (
	MachineAliases          = []string{"machines", "mach", "machs"}
	MachineClassAliases     = []string{"machineclasses", "mc", "mcs"}
	VolumeAliases           = []string{"volumes", "vol", "vols"}
	NetworkInterfaceAliases = []string{"networkinterfaces", "nic", "nics"}
)
