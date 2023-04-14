// Copyright 2023 OnMetal authors
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

package v1alpha1

import (
	"encoding/json"
	"net/netip"

	"go4.org/netipx"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

// ObjectSelector specifies how to select objects of a certain kind.
type ObjectSelector struct {
	// Kind is the kind of object to select.
	Kind string `json:"kind"`
	// LabelSelector is the label selector to select objects of the specified Kind by.
	metav1.LabelSelector `json:",inline"`
}

// ConfigMapKeySelector is a reference to a specific 'key' within a ConfigMap resource.
// In some instances, `key` is a required field.
// +structType=atomic
type ConfigMapKeySelector struct {
	// Name of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string `json:"name,omitempty"`
	// The key of the entry in the ConfigMap resource's `data` field to be used.
	// Some instances of this field may be defaulted, in others it may be
	// required.
	// +optional
	Key string `json:"key,omitempty"`
}

// SecretKeySelector is a reference to a specific 'key' within a Secret resource.
// In some instances, `key` is a required field.
// +structType=atomic
type SecretKeySelector struct {
	// Name of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string `json:"name,omitempty"`
	// The key of the entry in the Secret resource's `data` field to be used.
	// Some instances of this field may be defaulted, in others it may be
	// required.
	// +optional
	Key string `json:"key,omitempty"`
}

// LocalUIDReference is a reference to another entity including its UID
// +structType=atomic
type LocalUIDReference struct {
	// Name is the name of the referenced entity.
	Name string `json:"name"`
	// UID is the UID of the referenced entity.
	UID types.UID `json:"uid"`
}

// UIDReference is a reference to another entity in a potentially different namespace including its UID.
// +structType=atomic
type UIDReference struct {
	// Namespace is the namespace of the referenced entity. If empty,
	// the same namespace as the referring resource is implied.
	Namespace string `json:"namespace,omitempty"`
	// Name is the name of the referenced entity.
	Name string `json:"name"`
	// UID is the UID of the referenced entity.
	UID types.UID `json:"uid,omitempty"`
}

// IP is an IP address.
type IP struct {
	netip.Addr `json:"-"`
}

func (in *IP) DeepCopyInto(out *IP) {
	*out = *in
}

func (in *IP) DeepCopy() *IP {
	return &IP{in.Addr}
}

func (i IP) GomegaString() string {
	return i.String()
}

func (i *IP) UnmarshalJSON(b []byte) error {
	if len(b) == 4 && string(b) == "null" {
		i.Addr = netip.Addr{}
		return nil
	}

	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	p, err := netip.ParseAddr(str)
	if err != nil {
		return err
	}

	i.Addr = p
	return nil
}

func (i IP) MarshalJSON() ([]byte, error) {
	if i.IsZero() {
		// Encode unset/nil objects as JSON's "null".
		return []byte("null"), nil
	}
	return json.Marshal(i.String())
}

func (i IP) ToUnstructured() interface{} {
	if i.IsZero() {
		return nil
	}
	return i.Addr.String()
}

func (i *IP) IsValid() bool {
	return i != nil && i.Addr.IsValid()
}

func (i *IP) IsZero() bool {
	return i == nil || !i.Addr.IsValid()
}

func (i IP) Family() corev1.IPFamily {
	switch {
	case i.Is4():
		return corev1.IPv4Protocol
	case i.Is6():
		return corev1.IPv6Protocol
	default:
		return ""
	}
}

func (_ IP) OpenAPISchemaType() []string { return []string{"string"} }

func (_ IP) OpenAPISchemaFormat() string { return "ip" }

func NewIP(ip netip.Addr) IP {
	return IP{ip}
}

func ParseIP(s string) (IP, error) {
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return IP{}, err
	}
	return IP{addr}, nil
}

func ParseNewIP(s string) (*IP, error) {
	ip, err := ParseIP(s)
	if err != nil {
		return nil, err
	}
	return &ip, nil
}

func MustParseIP(s string) IP {
	return IP{netip.MustParseAddr(s)}
}

func MustParseNewIP(s string) *IP {
	ip, err := ParseNewIP(s)
	utilruntime.Must(err)
	return ip
}

func NewIPPtr(ip netip.Addr) *IP {
	return &IP{ip}
}

func PtrToIP(addr IP) *IP {
	return &addr
}

func EqualIPs(a, b IP) bool {
	return a == b
}

// IPRange is an IP range.
type IPRange struct {
	From IP `json:"from"`
	To   IP `json:"to"`
}

func (i *IPRange) Range() netipx.IPRange {
	if i == nil {
		return netipx.IPRange{}
	}
	return netipx.IPRangeFrom(i.From.Addr, i.To.Addr)
}

func (i *IPRange) IsValid() bool {
	return i != nil && i.Range().IsValid()
}

func (i *IPRange) IsZero() bool {
	return i == nil || i.Range().IsZero()
}

func (i IPRange) String() string {
	return i.Range().String()
}

func (i IPRange) GomegaString() string {
	return i.String()
}

func NewIPRange(ipRange netipx.IPRange) IPRange {
	return IPRange{From: NewIP(ipRange.From()), To: NewIP(ipRange.To())}
}

func NewIPRangePtr(ipRange netipx.IPRange) *IPRange {
	r := NewIPRange(ipRange)
	return &r
}

func PtrToIPRange(ipRange IPRange) *IPRange {
	return &ipRange
}

func IPRangeFrom(from IP, to IP) IPRange {
	return NewIPRange(netipx.IPRangeFrom(from.Addr, to.Addr))
}

func ParseIPRange(s string) (IPRange, error) {
	rng, err := netipx.ParseIPRange(s)
	if err != nil {
		return IPRange{}, err
	}
	return IPRange{From: IP{rng.From()}, To: IP{rng.To()}}, nil
}

func ParseNewIPRange(s string) (*IPRange, error) {
	rng, err := ParseIPRange(s)
	if err != nil {
		return nil, err
	}
	return &rng, nil
}

func MustParseIPRange(s string) IPRange {
	rng := netipx.MustParseIPRange(s)
	return IPRange{From: IP{rng.From()}, To: IP{rng.To()}}
}

func MustParseNewIPRange(s string) *IPRange {
	rng, err := ParseNewIPRange(s)
	utilruntime.Must(err)
	return rng
}

func EqualIPRanges(a, b IPRange) bool {
	return EqualIPs(a.From, b.From) && EqualIPs(a.To, b.To)
}

// IPPrefix represents a network prefix.
// +nullable
type IPPrefix struct {
	netip.Prefix `json:"-"`
}

func (i IPPrefix) GomegaString() string {
	return i.String()
}

func (i IPPrefix) IP() IP {
	return IP{i.Prefix.Addr()}
}

func (i *IPPrefix) UnmarshalJSON(b []byte) error {
	if len(b) == 4 && string(b) == "null" {
		i.Prefix = netip.Prefix{}
		return nil
	}

	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	p, err := netip.ParsePrefix(str)
	if err != nil {
		return err
	}

	i.Prefix = p
	return nil
}

func (i IPPrefix) MarshalJSON() ([]byte, error) {
	if i.IsZero() {
		// Encode unset/nil objects as JSON's "null".
		return []byte("null"), nil
	}
	return json.Marshal(i.String())
}

func (i IPPrefix) ToUnstructured() interface{} {
	if i.IsZero() {
		return nil
	}
	return i.String()
}

func (in *IPPrefix) DeepCopyInto(out *IPPrefix) {
	*out = *in
}

func (in *IPPrefix) DeepCopy() *IPPrefix {
	return &IPPrefix{in.Prefix}
}

func (in *IPPrefix) IsValid() bool {
	return in != nil && in.Prefix.IsValid()
}

func (in *IPPrefix) IsZero() bool {
	return in == nil || !in.Prefix.IsValid()
}

func (_ IPPrefix) OpenAPISchemaType() []string { return []string{"string"} }

func (_ IPPrefix) OpenAPISchemaFormat() string { return "ip-prefix" }

func NewIPPrefix(prefix netip.Prefix) *IPPrefix {
	return &IPPrefix{Prefix: prefix}
}

func ParseIPPrefix(s string) (IPPrefix, error) {
	prefix, err := netip.ParsePrefix(s)
	if err != nil {
		return IPPrefix{}, err
	}
	return IPPrefix{prefix}, nil
}

func ParseNewIPPrefix(s string) (*IPPrefix, error) {
	prefix, err := ParseIPPrefix(s)
	if err != nil {
		return nil, err
	}
	return &prefix, nil
}

func MustParseIPPrefix(s string) IPPrefix {
	return IPPrefix{netip.MustParsePrefix(s)}
}

func MustParseNewIPPrefix(s string) *IPPrefix {
	prefix, err := ParseNewIPPrefix(s)
	utilruntime.Must(err)
	return prefix
}

func PtrToIPPrefix(prefix IPPrefix) *IPPrefix {
	return &prefix
}

func EqualIPPrefixes(a, b IPPrefix) bool {
	return a == b
}

// Taint is attached to a resource pool. It has the specified Effect on
// any resource that does not tolerate the Taint.
type Taint struct {
	// The taint key to be applied to a resource pool.
	Key string `json:"key"`
	// The taint value corresponding to the taint key.
	Value string `json:"value,omitempty"`
	// The effect of the taint on resources
	// that do not tolerate the taint.
	// Valid effects are NoSchedule, PreferNoSchedule and NoExecute.
	Effect TaintEffect `json:"effect"`
}

type TaintEffect string

const (
	// TaintEffectNoSchedule causes to not allow new resources to schedule onto the resource pool unless they tolerate
	// the taint, but allow all already-running resources to continue running.
	// Enforced by the scheduler.
	TaintEffectNoSchedule TaintEffect = "NoSchedule"
)

// Toleration makes the resource the toleration is attached to tolerate any taint that matches
// the triple <key,value,effect> using the matching operator <operator>.
type Toleration struct {
	// Key is the taint key that the toleration applies to. Empty means match all taint keys.
	// If the key is empty, operator must be Exists; this combination means to match all values and all keys.
	Key string `json:"key,omitempty"`
	// Operator represents a key's relationship to the value.
	// Valid operators are Exists and Equal. Defaults to Equal.
	// Exists is equivalent to wildcard for value, so that a resource can
	// tolerate all taints of a particular category.
	Operator TolerationOperator `json:"operator,omitempty"`
	// Value is the taint value the toleration matches to.
	// If the operator is Exists, the value should be empty, otherwise just a regular string.
	Value string `json:"value,omitempty"`
	// Effect indicates the taint effect to match. Empty means match all taint effects.
	// When specified, allowed values are NoSchedule.
	Effect TaintEffect `json:"effect,omitempty"`
}

// ToleratesTaint checks if the toleration tolerates the taint.
// The matching follows the rules below:
// (1) Empty toleration.effect means to match all taint effects,
//
//	otherwise taint effect must equal to toleration.effect.
//
// (2) If toleration.operator is 'Exists', it means to match all taint values.
// (3) Empty toleration.key means to match all taint keys.
//
//	If toleration.key is empty, toleration.operator must be 'Exists';
//	this combination means to match all taint values and all taint keys.
func (t *Toleration) ToleratesTaint(taint *Taint) bool {
	if len(t.Effect) > 0 && t.Effect != taint.Effect {
		return false
	}

	if len(t.Key) > 0 && t.Key != taint.Key {
		return false
	}

	switch t.Operator {
	case "", TolerationOpEqual: // empty operator means Equal
		return t.Value == taint.Value
	case TolerationOpExists:
		return true
	default:
		return false
	}
}

// TolerationOperator is the set of operators that can be used in a toleration.
type TolerationOperator string

const (
	TolerationOpEqual  TolerationOperator = "Equal"
	TolerationOpExists TolerationOperator = "Exists"
)

// TolerateTaints returns if tolerations tolerate all taints
func TolerateTaints(tolerations []Toleration, taints []Taint) bool {
Outer:
	for _, taint := range taints {
		for _, toleration := range tolerations {
			if toleration.ToleratesTaint(&taint) {
				continue Outer
			}
		}
		return false
	}
	return true
}
