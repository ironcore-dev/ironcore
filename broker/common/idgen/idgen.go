// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package idgen

import (
	"crypto/rand"
	"encoding/hex"
	"io"
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
