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

package tableconverters

import (
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"github.com/onmetal/onmetal-api/orictl/tableconverter"
)

type Options struct {
	TransformMachine      func(funcs tableconverter.Funcs[*ori.Machine]) tableconverter.TableConverter[*ori.Machine]
	TransformMachineSlice func(funcs tableconverter.SliceFuncs[*ori.Machine]) tableconverter.TableConverter[[]*ori.Machine]
}

func transformIfNotNil[TC tableconverter.TableConverter[E], E any](tc TC, f func(TC) tableconverter.TableConverter[E]) tableconverter.TableConverter[E] {
	if f != nil {
		return f(tc)
	}
	return tc
}

func MakeAddToRegistry(opts Options) func(*tableconverter.Registry) error {
	var rb tableconverter.RegistryBuilder

	machine := transformIfNotNil[tableconverter.Funcs[*ori.Machine]](Machine, opts.TransformMachine)
	machineSlice := transformIfNotNil[tableconverter.SliceFuncs[*ori.Machine]](MachineSlice, opts.TransformMachineSlice)

	rb.Register(tableconverter.ToTagAndTypedAny[*ori.Machine](machine))
	rb.Register(tableconverter.ToTagAndTypedAny[[]*ori.Machine](machineSlice))
	rb.Register(tableconverter.ToTagAndTypedAny[*ori.MachineClass](MachineClass))
	rb.Register(tableconverter.ToTagAndTypedAny[[]*ori.MachineClass](MachineClassSlice))
	rb.Register(tableconverter.ToTagAndTypedAny[*ori.NetworkInterface](NetworkInterface))
	rb.Register(tableconverter.ToTagAndTypedAny[[]*ori.NetworkInterface](NetworkInterfaceSlice))
	rb.Register(tableconverter.ToTagAndTypedAny[*ori.Volume](Volume))
	rb.Register(tableconverter.ToTagAndTypedAny[[]*ori.Volume](VolumeSlice))
	return rb.AddToRegistry
}
