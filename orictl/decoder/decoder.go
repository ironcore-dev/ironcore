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

package decoder

import (
	"encoding/json"

	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

type Decoder interface {
	Decode(data []byte, v any) error
}

type Recognizer interface {
	Recognizes(data []byte) (ok, unknown bool, err error)
}

type jsonDecoder struct{}

var JSONDecoder = jsonDecoder{}

func (jsonDecoder) Decode(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (jsonDecoder) Recognizes(data []byte) (ok, unknown bool, err error) {
	return utilyaml.IsJSONBuffer(data), false, nil
}

type yamlDecoder struct{}

var YAMLDecoder = yamlDecoder{}

func (yamlDecoder) Decode(data []byte, v any) error {
	data, err := yaml.YAMLToJSON(data)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

type multiDecoder []Decoder

func MultiDecoder(decoders ...Decoder) Decoder {
	return multiDecoder(decoders)
}

func (d multiDecoder) Decode(data []byte, v any) error {
	var skipped []Decoder
	for _, decoder := range d {
		if recognizer, ok := decoder.(Recognizer); ok {
			ok, unknown, err := recognizer.Recognizes(data)
			if err != nil {
				return err
			}
			if unknown {
				skipped = append(skipped, decoder)
				continue
			}
			if !ok {
				continue
			}
			return decoder.Decode(data, v)
		}

		skipped = append(skipped, decoder)
	}

	var lastErr error
	for _, decoder := range skipped {
		if err := decoder.Decode(data, v); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return lastErr
}

var Default = MultiDecoder(JSONDecoder, YAMLDecoder)

var (
	Decode = Default.Decode
)
