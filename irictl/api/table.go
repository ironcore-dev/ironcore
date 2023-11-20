// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"fmt"
	"io"
)

type Header struct {
	Name string
}

type Cell any

type Row []Cell

type Table struct {
	Headers []Header
	Rows    []Row
}

func Write(table *Table, w io.Writer) {
	for i, header := range table.Headers {
		if i != 0 {
			_, _ = fmt.Fprint(w, "\t")
		}
		_, _ = fmt.Fprint(w, header.Name)
	}
	if len(table.Headers) > 0 {
		_, _ = fmt.Fprintln(w)
	}
	for i, row := range table.Rows {
		if i != 0 {
			_, _ = fmt.Fprintln(w)
		}
		for j, cell := range row {
			if j != 0 {
				_, _ = fmt.Fprint(w, "\t")
			}
			_, _ = fmt.Fprint(w, cell)
		}
	}
	if len(table.Rows) > 0 {
		_, _ = fmt.Fprintln(w)
	}
}
