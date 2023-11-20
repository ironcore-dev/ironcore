// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package renderer

import (
	gojson "encoding/json"
	"io"

	sigsyaml "sigs.k8s.io/yaml"
)

type yaml struct{}

func (yaml) Render(v any, w io.Writer) error {
	jsonData, err := gojson.Marshal(v)
	if err != nil {
		return err
	}

	data, err := sigsyaml.JSONToYAML(jsonData)
	if err != nil {
		return err
	}

	n, err := w.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return io.ErrShortWrite
	}
	return nil
}

var YAML = yaml{}

func init() {
	LocalRegistryBuilder.Register("yaml", YAML)
}
