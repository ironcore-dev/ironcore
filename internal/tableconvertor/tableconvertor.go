// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

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
