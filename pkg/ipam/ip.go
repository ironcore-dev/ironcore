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

func ParseIP(s string) net.IP {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil
	}
	if v4 := ip.To4(); v4 != nil {
		return v4
	}
	return ip
}

func IPClone(ip net.IP) net.IP {
	return append(ip[:0:0], ip...)
}

func IPAdd(ip net.IP, n int64) net.IP {
	ip = IPClone(ip)
	for i := len(ip) - 1; n != 0; i-- {
		n += int64(ip[i])
		ip[i] = uint8(n & 0xff)
		n >>= 8
	}
	return ip
}

func IPAddInt(ip net.IP, n Int) net.IP {
	ip = IPClone(ip)
	for i := len(ip) - 1; n.Cmp(IntZero) > 0; i-- {
		n = n.Add(Int64(int64(ip[i])))
		ip[i] = byte(n.And(IntTwoFiveFive).Uint64())
		n = n.RShift(8)
	}
	return ip
}

func IPDiff(a, b net.IP) Int {
	d := IntZero

	a = a.To16()
	b = b.To16()
	for i, _ := range a {
		db := int64(a[i]) - int64(b[i])
		d = d.LShift(8).Add(Int64(db))
	}
	return d
}

func IPCmp(a, b net.IP) int {
	d := 0
	a = a.To16()
	b = b.To16()
	for i, _ := range a {
		d = int(a[i]) - int(b[i])
		if d != 0 {
			break
		}
	}
	switch {
	case d < 0:
		return -1
	case d > 0:
		return 1
	default:
		return 0
	}
}

func IPtoCIDR(ip net.IP) *net.IPNet {
	return &net.IPNet{
		IP:   ip,
		Mask: net.CIDRMask(len(ip)*8, len(ip)*8),
	}
}
