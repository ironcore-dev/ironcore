// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"
	"text/template"
	"time"

	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	iriremotemachine "github.com/ironcore-dev/ironcore/iri/remote/machine"
	"github.com/ironcore-dev/ironcore/irictl-machine/clientcmd"
	"github.com/ironcore-dev/ironcore/irictl-machine/tableconverters"
	"github.com/ironcore-dev/ironcore/irictl/renderer"
	"github.com/ironcore-dev/ironcore/irictl/tableconverter"
	"github.com/ironcore-dev/ironcore/utils/generic"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Factory interface {
	Client() (iri.MachineRuntimeClient, func() error, error)
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
	fs.StringVar(&o.Address, "address", "", "Address to the iri server.")
}

func (o *Options) Config() (*clientcmd.Config, error) {
	return clientcmd.GetConfig(o.ConfigFile)
}

func TemplateTableBuilderFromColumns[E any](columns []clientcmd.Column) (tableconverter.Funcs[E], error) {
	tColumns := make([]tableconverter.TemplateTableColumn, len(columns))
	for i, col := range columns {
		tmpl, err := template.New(col.Name).Parse(col.Template)
		if err != nil {
			return tableconverter.Funcs[E]{}, fmt.Errorf("[column %s] error parsing template: %w", col.Name, err)
		}

		tColumns[i] = tableconverter.TemplateTableColumn{
			Name:     col.Name,
			Template: tmpl,
		}
	}

	return tableconverter.TemplateTableBuilder[E](tColumns...), nil
}

func modifyTableConverter[E any](
	reg *tableconverter.Registry,
	modifyFunc func(conv tableconverter.TableConverter[any]) tableconverter.TableConverter[any],
) error {
	tag := generic.ReflectType[E]()
	conv, err := reg.Lookup(tag)
	if err != nil {
		return err
	}

	oldConv := conv
	conv = modifyFunc(conv)

	// Fast track: no modification
	if conv == oldConv {
		return nil
	}

	if err := reg.Delete(tag); err != nil {
		return err
	}

	return reg.Register(tag, conv)
}

func applyTableConverterOverlay[E any](
	reg *tableconverter.Registry,
	prependColumns, appendColumns []clientcmd.Column,
) error {
	if len(prependColumns) == 0 && len(appendColumns) == 0 {
		return nil
	}

	var (
		toPrepend      tableconverter.Funcs[E]
		toPrependSlice tableconverter.SliceFuncs[E]
	)
	if len(prependColumns) > 0 {
		conv, err := TemplateTableBuilderFromColumns[E](prependColumns)
		if err != nil {
			return err
		}

		toPrepend = conv
		toPrependSlice = tableconverter.SliceFuncs[E](toPrepend)
	}

	var (
		toAppend      tableconverter.Funcs[E]
		toAppendSlice tableconverter.SliceFuncs[E]
	)
	if len(appendColumns) > 0 {
		conv, err := TemplateTableBuilderFromColumns[E](appendColumns)
		if err != nil {
			return err
		}

		toAppend = conv
		toAppendSlice = tableconverter.SliceFuncs[E](toAppend)
	}

	if err := modifyTableConverter[E](reg, func(conv tableconverter.TableConverter[any]) tableconverter.TableConverter[any] {
		var convs []tableconverter.TableConverter[any]
		if len(prependColumns) > 0 {
			convs = append(convs, tableconverter.TypedAnyConverter[E, tableconverter.Funcs[E]]{Converter: toPrepend})
		}
		convs = append(convs, conv)
		if len(appendColumns) > 0 {
			convs = append(convs, tableconverter.TypedAnyConverter[E, tableconverter.Funcs[E]]{Converter: toAppend})
		}

		return tableconverter.Zip(convs...)
	}); err != nil {
		return err
	}

	if err := modifyTableConverter[[]E](reg, func(conv tableconverter.TableConverter[any]) tableconverter.TableConverter[any] {
		var convs []tableconverter.TableConverter[any]
		if len(prependColumns) > 0 {
			convs = append(convs, tableconverter.TypedAnyConverter[[]E, tableconverter.SliceFuncs[E]]{Converter: toPrependSlice})
		}
		convs = append(convs, conv)
		if len(appendColumns) > 0 {
			convs = append(convs, tableconverter.TypedAnyConverter[[]E, tableconverter.SliceFuncs[E]]{Converter: toAppendSlice})
		}

		return tableconverter.Zip(convs...)
	}); err != nil {
		return err
	}

	return nil
}

func (o *Options) applyTableConvertersOverlay(cfg *clientcmd.TableConfig, reg *tableconverter.Registry) error {
	if err := applyTableConverterOverlay[*iri.Machine](reg, cfg.PrependMachineColumns, cfg.AppendMachineColumns); err != nil {
		return fmt.Errorf("error applying machine table converter overlay: %w", err)
	}

	return nil
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

	tableConvRegistry := tableconverter.NewRegistry()
	if err := tableconverters.AddToRegistry(tableConvRegistry); err != nil {
		return nil, err
	}

	if tableCfg := cfg.TableConfig; tableCfg != nil {
		if err := o.applyTableConvertersOverlay(tableCfg, tableConvRegistry); err != nil {
			return nil, err
		}
	}

	if err := registry.Register("table", renderer.NewTable(tableConvRegistry)); err != nil {
		return nil, err
	}

	return registry, nil
}

func (o *Options) Client() (iri.MachineRuntimeClient, func() error, error) {
	address, err := iriremotemachine.GetAddressWithTimeout(3*time.Second, o.Address)
	if err != nil {
		return nil, nil, err
	}

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("error dialing: %w", err)
	}

	return iri.NewMachineRuntimeClient(conn), conn.Close, nil
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
	VolumeAliases           = []string{"volumes", "vol", "vols"}
	NetworkInterfaceAliases = []string{"networkinterfaces", "nic", "nics"}
)
