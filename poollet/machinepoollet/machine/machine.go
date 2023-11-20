// Copyright 2023 IronCore authors
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

package machine

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
		return fmt.Errorf("invalid machine ID: %q", s)
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
