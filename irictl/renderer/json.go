// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package renderer

import (
	gojson "encoding/json"
	"io"
)

type json struct{}

func (json) Render(v any, w io.Writer) error {
	enc := gojson.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

var JSON = json{}

func init() {
	LocalRegistryBuilder.Register("json", JSON)
}
