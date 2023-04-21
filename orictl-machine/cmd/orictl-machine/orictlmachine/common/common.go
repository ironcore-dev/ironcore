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
	"time"

	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	oriremotemachine "github.com/onmetal/onmetal-api/ori/remote/machine"
	"github.com/onmetal/onmetal-api/orictl-machine/clientcmd"
	"github.com/onmetal/onmetal-api/orictl-machine/tableconverters"
	"github.com/onmetal/onmetal-api/orictl/renderer"
	"github.com/onmetal/onmetal-api/orictl/tableconverter"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Factory interface {
	Client() (ori.MachineRuntimeClient, func() error, error)
	Config() (*clientcmd.Config, error)
	Registry() (*renderer.Registry, error)
	OutputOptions() *OutputOptions
}

type Options struct {
	Address    string
	ConfigFile string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ConfigFile, clientcmd.RecommendedConfigPathFlag, "", "Config file to use")
	fs.StringVar(&o.Address, "address", "", "Address to the ori server.")
}

func (o *Options) Config() (*clientcmd.Config, error) {
	return clientcmd.GetConfig(o.ConfigFile)
}

func (o *Options) tableConvertersOptions(cfg *clientcmd.Config) tableconverters.Options {
	tableCfg := cfg.TableConfig
	if tableCfg == nil {
		return tableconverters.Options{}
	}

	opts := tableconverters.Options{}
	if len(tableCfg.WellKnownMachineLabels) > 0 {
		opts.TransformMachine = func(funcs tableconverter.Funcs[*ori.Machine]) tableconverter.TableConverter[*ori.Machine] {
			return tableconverter.Merge[*ori.Machine](
				tableconverter.WellKnownLabels[*ori.Machine](tableCfg.WellKnownMachineLabels),
				funcs,
			)
		}

		opts.TransformMachineSlice = func(funcs tableconverter.SliceFuncs[*ori.Machine]) tableconverter.TableConverter[[]*ori.Machine] {
			itemFuncs := tableconverter.Funcs[*ori.Machine](funcs)
			merged := tableconverter.MergeFuncs[*ori.Machine](
				tableconverter.WellKnownLabels[*ori.Machine](tableCfg.WellKnownMachineLabels),
				itemFuncs,
			)
			return tableconverter.SliceFuncs[*ori.Machine](merged)
		}
	}

	return opts
}

func (o *Options) Registry() (*renderer.Registry, error) {
	cfg, err := o.Config()
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	registry := renderer.NewRegistry()
	if err := renderer.AddToRegistry(registry); err != nil {
		return nil, err
	}

	tableConv := tableconverter.NewRegistry()
	if err := tableconverters.MakeAddToRegistry(o.tableConvertersOptions(cfg))(tableConv); err != nil {
		return nil, err
	}

	if err := registry.Register("table", renderer.NewTable(tableConv)); err != nil {
		return nil, err
	}

	return registry, nil
}

func (o *Options) Client() (ori.MachineRuntimeClient, func() error, error) {
	address, err := oriremotemachine.GetAddressWithTimeout(3*time.Second, o.Address)
	if err != nil {
		return nil, nil, err
	}

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("error dialing: %w", err)
	}

	return ori.NewMachineRuntimeClient(conn), conn.Close, nil
}

func (o *Options) OutputOptions() *OutputOptions {
	return &OutputOptions{
		factory: o,
	}
}

type OutputOptions struct {
	factory Factory
	Output  string
}

func (o *OutputOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Output, "output", "o", o.Output, "Output format.")
}

func (o *OutputOptions) Renderer(ifEmpty string) (renderer.Renderer, error) {
	output := o.Output
	if output == "" {
		output = ifEmpty
	}

	r, err := o.factory.Registry()
	if err != nil {
		return nil, err
	}

	return r.Get(output)
}

func (o *OutputOptions) RendererOrNil() (renderer.Renderer, error) {
	output := o.Output
	if output == "" {
		return nil, nil
	}

	r, err := o.factory.Registry()
	if err != nil {
		return nil, err
	}

	return r.Get(output)
}

var (
	MachineAliases          = []string{"machines", "mach", "machs"}
	MachineClassAliases     = []string{"machineclasses", "mc", "mcs"}
	VolumeAliases           = []string{"volumes", "vol", "vols"}
	NetworkInterfaceAliases = []string{"networkinterfaces", "nic", "nics"}
)
