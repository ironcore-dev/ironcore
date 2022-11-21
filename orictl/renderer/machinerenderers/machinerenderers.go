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

package machinerenderers

import (
	"github.com/onmetal/onmetal-api/orictl/renderer"
	"github.com/onmetal/onmetal-api/orictl/table/machinetableconverters"
)

var (
	RegistryBuilder renderer.RegistryBuilder
	AddToRegistry   = RegistryBuilder.AddToRegistry
	Registry        *renderer.Registry
)

func init() {
	RegistryBuilder.Add(renderer.AddToRegistry)
	RegistryBuilder.Register(
		"table", renderer.NewTable(machinetableconverters.Registry),
	)

	Registry = renderer.NewRegistry()
	if err := AddToRegistry(Registry); err != nil {
		panic(err)
	}
}
