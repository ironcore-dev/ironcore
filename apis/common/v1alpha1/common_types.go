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

package v1alpha1

import (
	"encoding/json"
	"inet.af/netaddr"
	corev1 "k8s.io/api/core/v1"
)

// ConfigMapKeySelector is a reference to a specific 'key' within a ConfigMap resource.
// In some instances, `key` is a required field.
type ConfigMapKeySelector struct {
	// The name of the ConfigMap resource being referred to.
	corev1.LocalObjectReference `json:",inline"`
	// The key of the entry in the ConfigMap resource's `data` field to be used.
	// Some instances of this field may be defaulted, in others it may be
	// required.
	// +optional
	Key string `json:"key,omitempty"`
}

// SecretKeySelector is a reference to a specific 'key' within a Secret resource.
// In some instances, `key` is a required field.
type SecretKeySelector struct {
	// The name of the Secret resource being referred to.
	corev1.LocalObjectReference `json:",inline"`
	// The key of the entry in the Secret resource's `data` field to be used.
	// Some instances of this field may be defaulted, in others it may be
	// required.
	// +optional
	Key string `json:"key,omitempty"`
}

// IP is an IP address.
//+kubebuilder:validation:Type=string
type IP struct {
	netaddr.IP `json:"-"`
}

func (in *IP) DeepCopyInto(out *IP) {
	*out = *in
}

func (i IP) GomegaString() string {
	return i.String()
}

func (i *IP) UnmarshalJSON(b []byte) error {
	if len(b) == 4 && string(b) == "null" {
		i.IP = netaddr.IP{}
		return nil
	}

	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	p, err := netaddr.ParseIP(str)
	if err != nil {
		return err
	}

	i.IP = p
	return nil
}

func (i *IP) MarshalJSON() ([]byte, error) {
	if i.IsZero() {
		// Encode unset/nil objects as JSON's "null".
		return []byte("null"), nil
	}
	return json.Marshal(i.String())
}

func (IP) OpenAPISchemaType() []string { return []string{"string"} }

func NewIP(ip netaddr.IP) IP {
	return IP{ip}
}

func ParseIP(s string) (IP, error) {
	addr, err := netaddr.ParseIP(s)
	if err != nil {
		return IP{}, err
	}
	return IP{addr}, nil
}

func MustParseIP(s string) IP {
	return IP{netaddr.MustParseIP(s)}
}

func NewIPPtr(ip netaddr.IP) *IP {
	return &IP{ip}
}

func PtrToIP(addr IP) *IP {
	return &addr
}

// IPRange is an IP range.
type IPRange struct {
	From IP `json:"from"`
	To   IP `json:"to"`
}

func (i *IPRange) Range() netaddr.IPRange {
	if i == nil {
		return netaddr.IPRange{}
	}
	return netaddr.IPRangeFrom(i.From.IP, i.To.IP)
}

func (i *IPRange) IsValid() bool {
	if i == nil {
		return false
	}
	return i.Range().IsValid()
}

func (i *IPRange) IsZero() bool {
	if i == nil {
		return true
	}
	return i.Range().IsZero()
}

func (i IPRange) String() string {
	return i.Range().String()
}

func (i IPRange) GomegaString() string {
	return i.String()
}

func NewIPRange(ipRange netaddr.IPRange) IPRange {
	return IPRange{From: NewIP(ipRange.From()), To: NewIP(ipRange.To())}
}

func NewIPRangePtr(ipRange netaddr.IPRange) *IPRange {
	r := NewIPRange(ipRange)
	return &r
}

func PtrToIPRange(ipRange IPRange) *IPRange {
	return &ipRange
}

func IPRangeFrom(from IP, to IP) IPRange {
	return NewIPRange(netaddr.IPRangeFrom(from.IP, to.IP))
}

func ParseIPRange(s string) (IPRange, error) {
	rng, err := netaddr.ParseIPRange(s)
	if err != nil {
		return IPRange{}, err
	}
	return IPRange{From: IP{rng.From()}, To: IP{rng.To()}}, nil
}

func MustParseIPRange(s string) IPRange {
	rng := netaddr.MustParseIPRange(s)
	return IPRange{From: IP{rng.From()}, To: IP{rng.To()}}
}

// IPPrefix represents a network prefix.
//+kubebuilder:validation:Type=string
//+nullable
type IPPrefix struct {
	netaddr.IPPrefix `json:"-"`
}

func (i IPPrefix) GomegaString() string {
	return i.String()
}

func (i *IPPrefix) UnmarshalJSON(b []byte) error {
	if len(b) == 4 && string(b) == "null" {
		i.IPPrefix = netaddr.IPPrefix{}
		return nil
	}

	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	p, err := netaddr.ParseIPPrefix(str)
	if err != nil {
		return err
	}

	i.IPPrefix = p
	return nil
}

func (i *IPPrefix) MarshalJSON() ([]byte, error) {
	if i.IsZero() {
		// Encode unset/nil objects as JSON's "null".
		return []byte("null"), nil
	}
	return json.Marshal(i.String())
}

func (in *IPPrefix) DeepCopyInto(out *IPPrefix) {
	*out = *in
}

func (IPPrefix) OpenAPISchemaType() []string { return []string{"string"} }

func NewIPPrefix(prefix netaddr.IPPrefix) IPPrefix {
	return IPPrefix{IPPrefix: prefix}
}

func ParseIPPrefix(s string) (IPPrefix, error) {
	prefix, err := netaddr.ParseIPPrefix(s)
	if err != nil {
		return IPPrefix{}, err
	}
	return IPPrefix{prefix}, nil
}

func MustParseIPPrefix(s string) IPPrefix {
	return IPPrefix{netaddr.MustParseIPPrefix(s)}
}

func NewIPPrefixPtr(prefix netaddr.IPPrefix) *IPPrefix {
	c := NewIPPrefix(prefix)
	return &c
}

func PtrToIPPrefix(prefix IPPrefix) *IPPrefix {
	return &prefix
}
