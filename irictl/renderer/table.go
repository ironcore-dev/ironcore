// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package renderer

import (
	"fmt"
	"io"

	iritable "github.com/ironcore-dev/ironcore/irictl/api"
	"github.com/ironcore-dev/ironcore/irictl/tableconverter"
	"github.com/ironcore-dev/ironcore/irictl/tabwriter"
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

	iritable.Write(tab, tw)
	return tw.Flush()
}
