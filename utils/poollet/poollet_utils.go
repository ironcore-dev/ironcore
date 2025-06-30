// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package poollet

import (
	"fmt"
	"strings"
)

type ID struct {
	Type string
	ID   string
}

func MakeID(typ, id string) ID {
	return ID{
		Type: typ,
		ID:   id,
	}
}

func (i *ID) UnmarshalText(data []byte) error {
	s := string(data)

	// Trim the quotes and split the type and ID.
	parts := strings.Split(strings.Trim(s, "\""), "://")
	if len(parts) != 2 {
		return fmt.Errorf("invalid ID: %q", s)
	}
	i.Type, i.ID = parts[0], parts[1]
	return nil
}

func (i *ID) String() string {
	return fmt.Sprintf("%s://%s", i.Type, i.ID)
}

func ParseID(s string) (ID, error) {
	var id ID
	return id, id.UnmarshalText([]byte(s))
}

// DownwardAPILabel makes a downward api label name from the given name.
func DownwardAPILabel(label_prefix, name string) string {
	return label_prefix + name
}

// DownwardAPIAnnotation makes a downward api annotation name from the given name.
func DownwardAPIAnnotation(annotation_prefix, name string) string {
	return annotation_prefix + name
}
