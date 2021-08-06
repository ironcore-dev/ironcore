/*
 * Copyright (c) 2021 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ipam

import (
	"net"
)

////////////////////////////////////////////////////////////////////////////////

func sub(a, b net.IPMask) []byte {
	if len(a) != len(b) {
		return nil
	}
	new := make(net.IPMask, len(a), len(a))
	for i, v := range a {
		new[i] = v - b[i]
	}
	return new
}

func inv(a []byte) []byte {
	new := make(net.IPMask, len(a), len(a))
	for i, v := range a {
		new[i] = ^v
	}
	return new
}

func or(a, b []byte) []byte {
	if len(a) != len(b) {
		return nil
	}
	new := make(net.IPMask, len(a), len(a))
	for i, v := range a {
		new[i] = v | b[i]
	}
	return new
}

func and(a, b []byte) []byte {
	if len(a) != len(b) {
		return nil
	}
	new := make(net.IPMask, len(a), len(a))
	for i, v := range a {
		new[i] = v & b[i]
	}
	return new
}

func isZero(a []byte) bool {
	for _, v := range a {
		if v != 0 {
			return false
		}
	}
	return true
}
