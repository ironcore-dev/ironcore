// Copyright 2022 IronCore authors
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

package renderers

import (
	"github.com/ironcore-dev/ironcore/orictl-bucket/tableconverters"
	"github.com/ironcore-dev/ironcore/orictl/renderer"
	"github.com/ironcore-dev/ironcore/orictl/tableconverter"
)

var (
	RegistryBuilder renderer.RegistryBuilder
	AddToRegistry   = RegistryBuilder.AddToRegistry
)

func init() {
	RegistryBuilder.Add(renderer.AddToRegistry)
	RegistryBuilder.Add(func(registry *renderer.Registry) error {
		tableConverter := tableconverter.NewRegistry()
		if err := tableconverters.AddToRegistry(tableConverter); err != nil {
			return err
		}
		return registry.Register("table", renderer.NewTable(tableConverter))
	})
}
