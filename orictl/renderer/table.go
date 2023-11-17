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

package renderer

import (
	"fmt"
	"io"

	oritable "github.com/ironcore-dev/ironcore/orictl/api"
	"github.com/ironcore-dev/ironcore/orictl/tableconverter"
	"github.com/ironcore-dev/ironcore/orictl/tabwriter"
)

type table struct {
	converter tableconverter.TableConverter[any]
}

func NewTable(converter tableconverter.TableConverter[any]) Renderer {
	return &table{converter: converter}
}

func (t *table) Render(v any, w io.Writer) error {
	tw := tabwriter.New(w)

	tab, err := t.converter.ConvertToTable(v)
	if err != nil {
		return err
	}

	if tab == nil || len(tab.Headers) == 0 && len(tab.Rows) == 0 {
		_, err := fmt.Fprintln(w, "No resources found")
		return err
	}

	oritable.Write(tab, tw)
	return tw.Flush()
}
