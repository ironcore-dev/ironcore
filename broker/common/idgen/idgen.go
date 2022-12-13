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

package idgen

import (
	"encoding/hex"
	"io"
	"math/rand"
	"strconv"
)

type randReader struct{}

func (randReader) Read(p []byte) (int, error) {
	return rand.Read(p)
}

type IDGen interface {
	Generate() string
}

type idGen struct {
	reader io.Reader
	length int
}

// NewIDGen creates a new IDGen. The length parameter is the desired length of the generated id.
func NewIDGen(r io.Reader, length int) IDGen {
	return &idGen{
		reader: r,
		length: length,
	}
}

func (g *idGen) Generate() string {
	data := make([]byte, (g.length/2)+1)
	for {
		_, _ = g.reader.Read(data)
		id := hex.EncodeToString(data)

		// Truncated versions of the id should not be numerical.
		if _, err := strconv.ParseInt(id[:12], 10, 64); err != nil {
			continue
		}

		return id[:g.length]
	}
}

var (
	Default = NewIDGen(randReader{}, DefaultIDLength)
)

const (
	DefaultIDLength = 63
)
