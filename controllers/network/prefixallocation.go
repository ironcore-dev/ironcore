// Copyright 2021 OnMetal authors
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

package network

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/onmetal/controller-utils/conditionutils"
	commonv1alpha1 "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	networkv1alpha1 "github.com/onmetal/onmetal-api/apis/network/v1alpha1"
	"inet.af/netaddr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"math/big"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type PrefixAllocationer interface {
	Object() client.Object
	PrefixRef() *networkv1alpha1.PrefixReference
	Request() networkv1alpha1.PrefixAllocationRequest
	Result() networkv1alpha1.PrefixAllocationResult
	ReadyState() PrefixAllocationReadyState
}

type PrefixAllocationReadyState int8

const (
	PrefixAllocationFailed PrefixAllocationReadyState = iota - 1
	PrefixAllocationUnknown
	PrefixAllocationPending
	PrefixAllocationSucceeded
)

type PrefixRequester interface {
	Object() client.Object
	Label() string
	Request() networkv1alpha1.PrefixAllocationRequest
	PrefixRef() *networkv1alpha1.PrefixReference
	PrefixSelector() *networkv1alpha1.PrefixSelector
}

type Prefixer interface {
	Object() client.Object
	Available() []commonv1alpha1.IPPrefix
}

type PrefixAllocation networkv1alpha1.PrefixAllocation

func (p *PrefixAllocation) Object() client.Object {
	return (*networkv1alpha1.PrefixAllocation)(p)
}

func (p *PrefixAllocation) PrefixRef() *networkv1alpha1.PrefixReference {
	return p.Spec.PrefixRef
}

func (p *PrefixAllocation) Request() networkv1alpha1.PrefixAllocationRequest {
	return p.Spec.PrefixAllocationRequest
}

func (p *PrefixAllocation) Result() networkv1alpha1.PrefixAllocationResult {
	return p.Status.PrefixAllocationResult
}

func (p *PrefixAllocation) ReadyState() PrefixAllocationReadyState {
	cond := &networkv1alpha1.PrefixAllocationCondition{}
	conditionutils.MustFindSlice(p.Status.Conditions, string(networkv1alpha1.PrefixAllocationReady), cond)
	switch {
	case cond.Status == corev1.ConditionTrue:
		return PrefixAllocationSucceeded
	case cond.Status == corev1.ConditionFalse && cond.Reason == networkv1alpha1.PrefixAllocationReadyReasonPending:
		return PrefixAllocationPending
	case cond.Status == corev1.ConditionFalse:
		return PrefixAllocationFailed
	default:
		return PrefixAllocationUnknown
	}
}

type ClusterPrefixAllocation networkv1alpha1.ClusterPrefixAllocation

func (a *ClusterPrefixAllocation) Object() client.Object {
	return (*networkv1alpha1.ClusterPrefixAllocation)(a)
}

func (a *ClusterPrefixAllocation) PrefixRef() *networkv1alpha1.PrefixReference {
	prefixRef := a.Spec.PrefixRef
	if prefixRef == nil {
		return nil
	}
	return &networkv1alpha1.PrefixReference{
		Kind: networkv1alpha1.ClusterPrefixKind,
		Name: prefixRef.Name,
	}
}

func (a *ClusterPrefixAllocation) Request() networkv1alpha1.PrefixAllocationRequest {
	return networkv1alpha1.PrefixAllocationRequest{
		Prefix:       a.Spec.Prefix,
		PrefixLength: a.Spec.PrefixLength,
	}
}

func (a *ClusterPrefixAllocation) Result() networkv1alpha1.PrefixAllocationResult {
	return networkv1alpha1.PrefixAllocationResult{
		Prefix: a.Status.Prefix,
	}
}

func (a *ClusterPrefixAllocation) ReadyState() PrefixAllocationReadyState {
	cond := &networkv1alpha1.ClusterPrefixAllocationCondition{}
	conditionutils.MustFindSlice(a.Status.Conditions, string(networkv1alpha1.ClusterPrefixAllocationReady), cond)
	switch {
	case cond.Status == corev1.ConditionTrue:
		return PrefixAllocationSucceeded
	case cond.Status == corev1.ConditionFalse && cond.Reason == networkv1alpha1.ClusterPrefixAllocationReadyReasonPending:
		return PrefixAllocationPending
	case cond.Status == corev1.ConditionFalse:
		return PrefixAllocationFailed
	default:
		return PrefixAllocationUnknown
	}
}

func IsOwned(scheme *runtime.Scheme, owner, controlled client.Object) (bool, error) {
	controller := metav1.GetControllerOf(controlled)
	if controller == nil {
		return false, nil
	}

	gvk, err := apiutil.GVKForObject(owner, scheme)
	if err != nil {
		return false, fmt.Errorf("error getting object kinds of owner: %w", err)
	}

	gv, err := schema.ParseGroupVersion(controller.APIVersion)
	if err != nil {
		return false, fmt.Errorf("could not parse controller api version: %w", err)
	}

	return gvk.GroupVersion() == gv &&
		controller.Kind == gvk.Kind &&
		controller.Name == owner.GetName() &&
		controller.UID == owner.GetUID(), nil
}

func RemoveIntersection(set *netaddr.IPSet, prefix netaddr.IPPrefix) ([]netaddr.IPPrefix, *netaddr.IPSet) {
	var prefixSetBuilder netaddr.IPSetBuilder
	prefixSetBuilder.AddPrefix(prefix)
	prefixSet, err := prefixSetBuilder.IPSet()
	if err != nil {
		return nil, set
	}

	var intersectBuilder netaddr.IPSetBuilder
	intersectBuilder.AddSet(set)
	intersectBuilder.Intersect(prefixSet)
	intersect, err := intersectBuilder.IPSet()
	if err != nil {
		return nil, set
	}

	prefixes := intersect.Prefixes()
	if len(prefixes) == 0 {
		return nil, set
	}

	var setBuilder netaddr.IPSetBuilder
	setBuilder.AddSet(set)
	setBuilder.RemoveSet(intersect)
	set, err = setBuilder.IPSet()
	if err != nil {
		return nil, set
	}
	return prefixes, set
}

func IPSetFromPrefixes(prefixes []commonv1alpha1.IPPrefix) (*netaddr.IPSet, error) {
	var builder netaddr.IPSetBuilder
	for _, prefix := range prefixes {
		builder.AddPrefix(prefix.IPPrefix)
	}
	return builder.IPSet()
}

func CanIPSetFitRequest(set *netaddr.IPSet, request networkv1alpha1.PrefixAllocationRequest) bool {
	switch {
	case request.Prefix.IsValid():
		return set.ContainsPrefix(request.Prefix.IPPrefix)
	case request.PrefixLength > 0:
		_, _, ok := set.RemoveFreePrefix(uint8(request.PrefixLength))
		return ok
	case request.Range.IsValid():
		return set.ContainsRange(request.Range.Range())
	case request.RangeLength > 0:
		cmp := big.NewInt(request.RangeLength)
		for _, rng := range set.Ranges() {
			if RangeSize(rng).Cmp(cmp) >= 0 {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func CanPrefixFitRequest(prefix *networkv1alpha1.Prefix, request networkv1alpha1.PrefixAllocationRequest) bool {
	set, err := IPSetFromPrefixes(prefix.Status.Available)
	if err != nil {
		return false
	}

	return CanIPSetFitRequest(set, request)
}

func CanClusterPrefixFitRequest(clusterPrefix *networkv1alpha1.ClusterPrefix, request networkv1alpha1.PrefixAllocationRequest) bool {
	set, err := IPSetFromPrefixes(clusterPrefix.Status.Available)
	if err != nil {
		return false
	}

	return CanIPSetFitRequest(set, request)
}

func IPSetFromNetaddrPrefix(prefix netaddr.IPPrefix) (*netaddr.IPSet, error) {
	var builder netaddr.IPSetBuilder
	builder.AddPrefix(prefix)
	return builder.IPSet()
}

func RemoveFreePrefixForRequest(set *netaddr.IPSet, allocation networkv1alpha1.PrefixAllocationRequest) (netaddr.IPPrefix, netaddr.IPRange, *netaddr.IPSet, bool) {
	switch {
	case allocation.Prefix.IsValid():
		if set, ok := IPSetRemovePrefix(set, allocation.Prefix.IPPrefix); ok {
			return allocation.Prefix.IPPrefix, netaddr.IPRange{}, set, true
		}
		return netaddr.IPPrefix{}, netaddr.IPRange{}, set, false
	case allocation.PrefixLength > 0:
		if prefix, set, ok := set.RemoveFreePrefix(uint8(allocation.PrefixLength)); ok {
			return prefix, netaddr.IPRange{}, set, true
		}
		return netaddr.IPPrefix{}, netaddr.IPRange{}, set, false
	case allocation.Range.IsValid():
		if set, ok := IPSetRemoveRange(set, allocation.Range.Range()); ok {
			return netaddr.IPPrefix{}, allocation.Range.Range(), set, true
		}
		return netaddr.IPPrefix{}, netaddr.IPRange{}, set, false
	case allocation.RangeLength > 0:
		if rng, set, ok := IPSetRemoveFreeRange(set, allocation.RangeLength); ok {
			return netaddr.IPPrefix{}, rng, set, true
		}
		return netaddr.IPPrefix{}, netaddr.IPRange{}, set, false
	default:
		return netaddr.IPPrefix{}, netaddr.IPRange{}, set, false
	}
}

func IPSetRemovePrefix(set *netaddr.IPSet, prefix netaddr.IPPrefix) (*netaddr.IPSet, bool) {
	if !prefix.IsValid() || !set.ContainsPrefix(prefix) {
		return set, false
	}
	var sb netaddr.IPSetBuilder
	sb.AddSet(set)
	sb.RemovePrefix(prefix)
	set, _ = sb.IPSet()
	return set, true
}

func IPSetRemoveRange(set *netaddr.IPSet, rng netaddr.IPRange) (*netaddr.IPSet, bool) {
	if !rng.IsValid() || !set.ContainsRange(rng) {
		return set, false
	}
	var sb netaddr.IPSetBuilder
	sb.AddSet(set)
	sb.RemoveRange(rng)
	set, _ = sb.IPSet()
	return set, true
}

func IPRangeFromIPPlusLength(from netaddr.IP, length int64) (netaddr.IPRange, error) {
	if !from.IsValid() {
		return netaddr.IPRange{}, fmt.Errorf("invalid ip")
	}
	if length == 0 {
		return netaddr.IPRange{}, fmt.Errorf("invalid range length %d: has to be >= 1", length)
	}

	// We have to decrement the length here as a single IP already is a range of length 1.
	length = length - 1

	if from.Is4() {
		data := from.As4()
		res := (&big.Int{}).Add((&big.Int{}).SetBytes(data[:]), big.NewInt(length))
		if res.BitLen() > 32 {
			return netaddr.IPRange{}, fmt.Errorf("%s plus length %d is not 32 bits / ipv4", from, length)
		}
		res.FillBytes(data[:])
		to := netaddr.IPFrom4(data)
		return netaddr.IPRangeFrom(from, to), nil
	} else {
		data := from.As16()
		res := (&big.Int{}).Add((&big.Int{}).SetBytes(data[:]), big.NewInt(length))
		if res.BitLen() > 128 {
			return netaddr.IPRange{}, fmt.Errorf("%s plus length %d is not 128 bits / ipv6", from, length)
		}
		res.FillBytes(data[:])
		to := netaddr.IPFrom16(data)
		return netaddr.IPRangeFrom(from, to), nil
	}
}

func IPSetRemoveFreeRange(set *netaddr.IPSet, rngLength int64) (netaddr.IPRange, *netaddr.IPSet, bool) {
	var (
		best     netaddr.IPRange
		bestSize *big.Int
		wanted   = big.NewInt(rngLength)
	)
	for _, rng := range set.Ranges() {
		size := RangeSize(rng)
		cmp := size.Cmp(wanted)
		if cmp >= 0 && (bestSize == nil || size.Cmp(bestSize) < 0) {
			best = rng
			bestSize = size
			if cmp == 0 {
				set, _ = IPSetRemoveRange(set, rng)
				return rng, set, true
			}
		}
	}
	if !best.IsValid() {
		return netaddr.IPRange{}, set, false
	}
	rng, _ := IPRangeFromIPPlusLength(best.From(), rngLength)
	set, _ = IPSetRemoveRange(set, rng)
	return rng, set, true
}

// RangeSize computes the number of ips in a range.
// TODO: we should come up with more efficient arithmetics as the netaddr package does.
func RangeSize(rng netaddr.IPRange) *big.Int {
	if !rng.IsValid() {
		return big.NewInt(0)
	}

	// We have to add '1' in both computation branches since an IP range is an inclusive range.
	if rng.From().Is4() {
		fromBytes := rng.From().As4()
		toBytes := rng.To().As4()
		from := binary.BigEndian.Uint32(fromBytes[:])
		to := binary.BigEndian.Uint32(toBytes[:])
		return big.NewInt(int64(to-from) + 1)
	} else {
		fromBytes := rng.From().As16()
		toBytes := rng.To().As16()
		from := &big.Int{}
		from = from.SetBytes(fromBytes[:])
		to := &big.Int{}
		to = to.SetBytes(toBytes[:])
		res := (&big.Int{}).Sub(to, from)
		return res.Add(res, big.NewInt(1))
	}
}

func PrefixAllocationRequestMatchesAllocator(request networkv1alpha1.PrefixAllocationRequest, allocation PrefixAllocationer) bool {
	if allocation.ReadyState() == PrefixAllocationSucceeded {
		switch {
		case request.Prefix.IsValid():
			return request.Prefix == allocation.Result().Prefix
		case request.PrefixLength > 0:
			return uint8(request.PrefixLength) == allocation.Result().Prefix.Bits()
		case request.Range.IsValid():
			return request.Range.Range() == allocation.Result().Range.Range()
		case request.RangeLength > 0:
			rng := allocation.Result().Range
			rngSize := RangeSize(rng.Range())
			if !rngSize.IsInt64() {
				return false
			}
			return rngSize.Int64() == request.RangeLength
		default:
			return false
		}
	}

	switch {
	case request.Prefix.IsValid():
		return request.Prefix == allocation.Request().Prefix
	case request.Range.IsValid():
		return request.Range.Range() == allocation.Request().Range.Range()
	case request.PrefixLength > 0:
		return request.PrefixLength == allocation.Request().PrefixLength
	case request.RangeLength > 0:
		return request.RangeLength == allocation.Request().RangeLength
	default:
		return false
	}
}

func (a *PrefixAllocator) splitAllocations(
	request networkv1alpha1.PrefixAllocationRequest,
	allocations []PrefixAllocationer,
) (active PrefixAllocationer, other []PrefixAllocationer) {
	var activeReadyState PrefixAllocationReadyState
	for _, allocator := range allocations {
		readyState := allocator.ReadyState()
		if PrefixAllocationRequestMatchesAllocator(request, allocator) && readyState >= PrefixAllocationUnknown {
			allocationCreationTimestamp := allocator.Object().GetCreationTimestamp()
			if active == nil || readyState >= activeReadyState &&
				allocationCreationTimestamp.Time.Before(active.Object().GetCreationTimestamp().Time) {
				if active != nil {
					other = append(other, active)
				}
				active = allocator
				activeReadyState = readyState
				continue
			}
		}

		other = append(other, allocator)
	}
	return active, other
}

func (a *PrefixAllocator) prunePrefixAllocators(ctx context.Context, allocators []PrefixAllocationer) error {
	var errs []error
	for _, allocation := range allocators {
		if err := a.Delete(ctx, allocation.Object()); client.IgnoreNotFound(err) != nil {
			errs = append(errs, fmt.Errorf("error deleting %s: %w", client.ObjectKeyFromObject(allocation.Object()), err))
		}
	}
	if len(errs) != 0 {
		return fmt.Errorf("error(s) occurred pruning prefix allocations: %v", errs)
	}
	return nil
}

// IsAPINamespaced returns true if the object is namespace scoped.
// For unstructured objects the gvk is found from the object itself.
func IsAPINamespaced(obj runtime.Object, scheme *runtime.Scheme, mapper meta.RESTMapper) (bool, error) {
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		return false, err
	}

	return IsAPINamespacedWithGVK(gvk, mapper)
}

// IsAPINamespacedWithGVK returns true if the object having the provided
// GVK is namespace scoped.
func IsAPINamespacedWithGVK(gk schema.GroupVersionKind, mapper meta.RESTMapper) (bool, error) {
	mapping, err := mapper.RESTMapping(schema.GroupKind{Group: gk.Group, Kind: gk.Kind})
	if err != nil {
		return false, fmt.Errorf("failed to get restmapping: %w", err)
	}

	scope := mapping.Scope.Name()

	if scope == "" {
		return false, fmt.Errorf("scope cannot be identified, empty scope returned")
	}

	if scope != meta.RESTScopeNameRoot {
		return true, nil
	}
	return false, nil
}

func CreateAllocator(scheme *runtime.Scheme, mapper meta.RESTMapper, requester PrefixRequester) (PrefixAllocationer, error) {
	ok, err := IsAPINamespaced(requester.Object(), scheme, mapper)
	if err != nil {
		return nil, fmt.Errorf("error determining whether object is namespaced: %w", err)
	}

	var (
		generateName = fmt.Sprintf("%s-", requester.Object().GetName())
		labels       = map[string]string{
			requester.Label(): requester.Object().GetName(),
		}
	)
	if ok {
		spec := networkv1alpha1.PrefixAllocationSpec{
			PrefixAllocationRequest: requester.Request(),
		}
		switch {
		case requester.PrefixRef() != nil:
			spec.PrefixRef = requester.PrefixRef()
		case requester.PrefixSelector() != nil:
			spec.PrefixSelector = requester.PrefixSelector()
		default:
			return nil, fmt.Errorf("neither prefix ref nor prefix selector specified")
		}

		return &PrefixAllocation{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:    requester.Object().GetNamespace(),
				GenerateName: generateName,
				Labels:       labels,
			},
			Spec: spec,
		}, nil
	} else {
		request := requester.Request()
		if request.Range.IsValid() || request.RangeLength > 0 {
			return nil, fmt.Errorf("cannot request ranges on cluster level")
		}

		spec := networkv1alpha1.ClusterPrefixAllocationSpec{
			ClusterPrefixAllocationRequest: networkv1alpha1.ClusterPrefixAllocationRequest{
				Prefix:       request.Prefix,
				PrefixLength: request.PrefixLength,
			},
		}
		switch {
		case requester.PrefixRef() != nil:
			spec.PrefixRef = &corev1.LocalObjectReference{Name: requester.PrefixRef().Name}
		case requester.PrefixSelector() != nil:
			labelSelector := requester.PrefixSelector().LabelSelector
			spec.PrefixSelector = &labelSelector
		default:
			return nil, fmt.Errorf("neither prefix ref nor prefix selector specified")
		}

		return &ClusterPrefixAllocation{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: generateName,
				Labels:       labels,
			},
			Spec: spec,
		}, nil
	}
}

func (a *PrefixAllocator) createAllocation(ctx context.Context, log logr.Logger, requester PrefixRequester) (PrefixAllocationer, error) {
	log.V(1).Info("Creating allocation")
	allocator, err := CreateAllocator(a.Scheme, a.RESTMapper(), requester)
	if err != nil {
		return nil, fmt.Errorf("error creating allocator: %w", err)
	}

	if err := ctrl.SetControllerReference(requester.Object(), allocator.Object(), a.Scheme); err != nil {
		return nil, fmt.Errorf("error owning allocator: %w", err)
	}
	if err := a.Create(ctx, allocator.Object()); err != nil {
		return nil, fmt.Errorf("error creating allocation: %w", err)
	}
	return allocator, nil
}

type PrefixAllocator struct {
	client.Client
	Scheme *runtime.Scheme
}

func NewPrefixAllocator(c client.Client, scheme *runtime.Scheme) *PrefixAllocator {
	return &PrefixAllocator{c, scheme}
}

func ListOwned(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, list client.ObjectList, opts ...client.ListOption) error {
	if err := c.List(ctx, list, opts...); err != nil {
		return fmt.Errorf("error listing: %w", err)
	}

	items, err := meta.ExtractList(list)
	if err != nil {
		return fmt.Errorf("error extracting list items: %w", err)
	}

	var res []runtime.Object
	for _, runtimeObj := range items {
		obj, ok := runtimeObj.(client.Object)
		if !ok {
			return fmt.Errorf("object is no client.Object: %v", runtimeObj)
		}

		ok, err := IsOwned(scheme, owner, obj)
		if err != nil {
			return fmt.Errorf("error checking whether object is owned: %w", err)
		}

		if ok {
			res = append(res, obj)
		}
	}

	if err := meta.SetList(list, res); err != nil {
		return fmt.Errorf("error setting list: %w", err)
	}
	return nil
}

func (a *PrefixAllocator) listOwnedAllocations(ctx context.Context, requester PrefixRequester) ([]PrefixAllocationer, error) {
	ok, err := IsAPINamespaced(requester.Object(), a.Scheme, a.RESTMapper())
	if err != nil {
		return nil, fmt.Errorf("error determining whether request is namespaced: %w", err)
	}

	if ok {
		list := &networkv1alpha1.PrefixAllocationList{}
		if err := ListOwned(ctx, a.Client, a.Scheme, requester.Object(), list, client.InNamespace(requester.Object().GetNamespace()), client.MatchingLabels{
			requester.Label(): requester.Object().GetName(),
		}); err != nil {
			return nil, fmt.Errorf("error listing prefix allocations: %w", err)
		}

		res := make([]PrefixAllocationer, 0, len(list.Items))
		for _, allocation := range list.Items {
			res = append(res, (*PrefixAllocation)(&allocation))
		}
		return res, nil
	} else {
		list := &networkv1alpha1.ClusterPrefixAllocationList{}
		if err := ListOwned(ctx, a.Client, a.Scheme, requester.Object(), list, client.MatchingLabels{
			requester.Label(): requester.Object().GetName(),
		}); err != nil {
			return nil, fmt.Errorf("error listing prefix allocations: %w", err)
		}

		res := make([]PrefixAllocationer, 0, len(list.Items))
		for _, allocation := range list.Items {
			res = append(res, (*ClusterPrefixAllocation)(&allocation))
		}
		return res, nil
	}
}

func (a *PrefixAllocator) Apply(ctx context.Context, requester PrefixRequester) (PrefixAllocationer, error) {
	log := ctrl.LoggerFrom(ctx)

	allocators, err := a.listOwnedAllocations(ctx, requester)
	if err != nil {
		return nil, fmt.Errorf("error listing owned prefix allocations: %w", err)
	}

	active, other := a.splitAllocations(requester.Request(), allocators)
	defer func() {
		if err := a.prunePrefixAllocators(ctx, other); err != nil {
			log.Error(err, "Error pruning prefix allocations")
		}
	}()

	if active == nil {
		active, err = a.createAllocation(ctx, log, requester)
		if err != nil {
			return nil, fmt.Errorf("error creating prefix allocation: %w", err)
		}
	}
	return active, nil
}
