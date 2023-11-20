// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package device

import (
	"fmt"
	"regexp"

	"github.com/bits-and-blooms/bitset"
)

const (
	letters    = "abcdefghijklmnopqrstuvwxyz"
	numLetters = len(letters)

	// MaxIndex is the maximum index usable for Name / returned by ParseName.
	MaxIndex = numLetters*numLetters + numLetters - 1

	// IronCorePrefix is the device prefix used by ironcore devices.
	IronCorePrefix = "od"
)

var (
	prefixRegex = regexp.MustCompile("^[a-z]{2}$")
	nameRegex   = regexp.MustCompile("^(?P<prefix>[a-z]{2})(?P<index>[a-z][a-z]?)$")
)

// Name creates a device name with the given prefix and index.
// If idx is greater than MaxIndex or if the prefix is not a valid prefix (two lowercase letters) it panics.
func Name(prefix string, idx int) string {
	if !prefixRegex.MatchString(prefix) {
		panic(fmt.Sprintf("invalid device prefix %s - must match regex %s", prefix, prefixRegex))
	}

	if idx > MaxIndex {
		panic(fmt.Sprintf("device index too large %d", idx))
	}
	if idx < 0 {
		panic(fmt.Sprintf("negative device index %d", idx))
	}

	if idx < numLetters {
		return prefix + string(letters[idx])
	}

	rmd := idx % numLetters
	idx = (idx / numLetters) - 1
	return prefix + string(letters[idx]) + string(letters[rmd])
}

// ParseName parses the name into its prefix and index. An error is returned if the name is not a valid device name.
func ParseName(name string) (string, int, error) {
	match := nameRegex.FindStringSubmatch(name)
	if match == nil {
		return "", 0, fmt.Errorf("%s does not match device name regex %s", name, nameRegex)
	}

	prefix := match[1]

	idxStr := match[2]
	if len(idxStr) == 1 {
		idx := int(idxStr[0] - 'a')
		return prefix, idx, nil
	}

	r1, r2 := int(idxStr[0]-'a'), int(idxStr[1]-'a')
	idx := (r1+1)*numLetters + r2

	return prefix, idx, nil
}

type taken struct {
	count uint
	set   *bitset.BitSet
}

func newTaken(bits uint) *taken {
	return &taken{set: bitset.New(bits)}
}

func (t *taken) claim(i uint) bool {
	if t.set.Test(i) {
		return false
	}

	t.count++
	t.set.Set(i)
	return true
}

func (t *taken) release(i uint) bool {
	if !t.set.Test(i) {
		return false
	}

	t.count--
	t.set.Clear(i)
	return true
}

func (t *taken) pop() (uint, bool) {
	if t.count == t.set.Len() {
		return 0, false
	}

	// We can safely assume that there is a next clear.
	idx, _ := t.set.NextClear(0)

	t.count++
	t.set.Set(idx)
	return idx, true
}

// Namer allows managing multiple device names. It remembers reserved ones and allows claiming / releasing new ones.
type Namer struct {
	takenByPrefix map[string]*taken
}

// NewNamer creates a new Namer that allows managing multiple device names.
func NewNamer() *Namer {
	return &Namer{takenByPrefix: make(map[string]*taken)}
}

// Observe marks the given name as already claimed. If it already has been claimed, an error is returned.
func (n *Namer) Observe(name string) error {
	prefix, idx, err := ParseName(name)
	if err != nil {
		return err
	}

	t := n.takenByPrefix[prefix]
	if t == nil {
		t = newTaken(uint(MaxIndex))
		n.takenByPrefix[prefix] = t
	}

	if !t.claim(uint(idx)) {
		return fmt.Errorf("index %d is already occupied", idx)
	}
	return nil
}

// Generate generates and claims a new free device name. An error is returned if no free device names are available.
func (n *Namer) Generate(prefix string) (string, error) {
	t := n.takenByPrefix[prefix]
	if t == nil {
		t = newTaken(uint(MaxIndex))
		n.takenByPrefix[prefix] = t
	}

	idx, ok := t.pop()
	if !ok {
		return "", fmt.Errorf("no free device names available")
	}

	return Name(prefix, int(idx)), nil
}

// Free releases the given name. If it has not been claimed before, an error is returned.
func (n *Namer) Free(name string) error {
	prefix, idx, err := ParseName(name)
	if err != nil {
		return err
	}

	t := n.takenByPrefix[prefix]
	if t == nil {
		t = newTaken(uint(MaxIndex))
		n.takenByPrefix[prefix] = t
	}

	if !t.release(uint(idx)) {
		return fmt.Errorf("index %d is not claimed", idx)
	}
	return nil
}
