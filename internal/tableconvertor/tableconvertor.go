// Copyright 2022 IronCore authors
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

package tableconvertor

import (
	"fmt"
	"strings"
)

func JoinStringsMore(elems []string, sep string, max int) string {
	if max < 1 {
		panic(fmt.Sprintf("JoinStringsMore: max < 1 (%d)", max))
	}

	if len(elems) == 0 {
		return "<unset>"
	}

	diff := len(elems) - max
	if diff <= 0 {
		return strings.Join(elems, sep)
	}
	return fmt.Sprintf("%s + %d more", strings.Join(elems[:max], sep), diff)
}
