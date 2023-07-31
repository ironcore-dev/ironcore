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

package claimmanager

import (
	"context"
	"fmt"

	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	utilslices "github.com/onmetal/onmetal-api/utils/slices"
	"golang.org/x/exp/slices"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClaimState uint8

const (
	ClaimStateFree ClaimState = iota
	ClaimStateClaimed
	ClaimStateTaken
)

type Selector[O client.Object] interface {
	Match(obj O) bool
}

type MatchingLabelSelector[O client.Object] struct {
	Selector labels.Selector
}

func (s MatchingLabelSelector[O]) Match(obj O) bool {
	return s.Selector.Matches(labels.Set(obj.GetLabels()))
}

func MatchingLabels[O client.Object](lbls map[string]string) MatchingLabelSelector[O] {
	return MatchingLabelSelector[O]{Selector: labels.SelectorFromSet(lbls)}
}

type ClaimStrategy[O client.Object] interface {
	ClaimState(claimer client.Object, obj O) ClaimState
	Adopt(ctx context.Context, claimer client.Object, obj O) error
	Release(ctx context.Context, claimer client.Object, obj O) error
}

type LocalUIDReferenceAccessor interface {
	IsTaken(otherRefs []commonv1alpha1.LocalUIDReference) bool
	GetLocalUIDReferences() []commonv1alpha1.LocalUIDReference
	SetLocalUIDReferences(refs []commonv1alpha1.LocalUIDReference)
}

type LocalUIDAccessorFor[O client.Object] func(obj O) LocalUIDReferenceAccessor

type localUIDReferenceClaimStrategy[O client.Object] struct {
	client      client.Client
	accessorFor LocalUIDAccessorFor[O]
}

func (l *localUIDReferenceClaimStrategy[O]) ClaimState(claimer client.Object, obj O) ClaimState {
	acc := l.accessorFor(obj)
	refs := acc.GetLocalUIDReferences()
	if len(refs) == 0 {
		return ClaimStateFree
	}
	var otherRefs []commonv1alpha1.LocalUIDReference
	for _, ref := range refs {
		if ref.UID == claimer.GetUID() {
			return ClaimStateClaimed
		}
		otherRefs = append(otherRefs, ref)
	}
	if acc.IsTaken(otherRefs) {
		return ClaimStateTaken
	}
	return ClaimStateFree
}

func (l *localUIDReferenceClaimStrategy[O]) Adopt(ctx context.Context, claimer client.Object, obj O) error {
	acc := l.accessorFor(obj)
	refs := slices.Clone(acc.GetLocalUIDReferences())

	base := obj.DeepCopyObject().(O)
	refs = append(refs, commonv1alpha1.LocalUIDReference{
		Name: claimer.GetName(),
		UID:  claimer.GetUID(),
	})
	acc.SetLocalUIDReferences(refs)
	return l.client.Patch(ctx, obj, client.MergeFromWithOptions(base, client.MergeFromWithOptimisticLock{}))
}

func (l *localUIDReferenceClaimStrategy[O]) Release(ctx context.Context, claimer client.Object, obj O) error {
	acc := l.accessorFor(obj)
	refs := slices.Clone(acc.GetLocalUIDReferences())

	base := obj.DeepCopyObject().(O)
	refs = utilslices.FilterNot(refs, commonv1alpha1.LocalUIDReference{
		Name: claimer.GetName(),
		UID:  claimer.GetUID(),
	})
	acc.SetLocalUIDReferences(refs)
	return l.client.Patch(ctx, obj, client.MergeFromWithOptions(base, client.MergeFromWithOptimisticLock{}))
}

type LocalUIDReferencePointerAccessor struct {
	LocalUIDReferenceField **commonv1alpha1.LocalUIDReference
}

func (l *LocalUIDReferencePointerAccessor) IsTaken([]commonv1alpha1.LocalUIDReference) bool {
	return true
}

func (l *LocalUIDReferencePointerAccessor) GetLocalUIDReferences() []commonv1alpha1.LocalUIDReference {
	ref := *l.LocalUIDReferenceField
	if ref == nil {
		return nil
	}
	return []commonv1alpha1.LocalUIDReference{*ref}
}

func (l *LocalUIDReferencePointerAccessor) SetLocalUIDReferences(refs []commonv1alpha1.LocalUIDReference) {
	if len(refs) > 1 {
		panic(fmt.Sprintf("cannot set more than one local uid reference (got %d): %v", len(refs), refs))
	}

	if len(refs) == 1 {
		ref := refs[0]
		*l.LocalUIDReferenceField = &ref
		return
	}

	*l.LocalUIDReferenceField = nil
}

func AccessViaLocalUIDReferenceField[O client.Object](f func(obj O) **commonv1alpha1.LocalUIDReference) LocalUIDAccessorFor[O] {
	return func(obj O) LocalUIDReferenceAccessor {
		return &LocalUIDReferencePointerAccessor{
			LocalUIDReferenceField: f(obj),
		}
	}
}

func LocalUIDReferenceClaimStrategy[O client.Object](
	c client.Client,
	accessorFor LocalUIDAccessorFor[O],
) ClaimStrategy[O] {
	return &localUIDReferenceClaimStrategy[O]{
		client:      c,
		accessorFor: accessorFor,
	}
}

type ClaimManager[O client.Object] struct {
	claimer  client.Object
	selector Selector[O]
	strategy ClaimStrategy[O]
}

func New[O client.Object](
	claimer client.Object,
	selector Selector[O],
	strategy ClaimStrategy[O],
) *ClaimManager[O] {
	return &ClaimManager[O]{
		claimer:  claimer,
		selector: selector,
		strategy: strategy,
	}
}

func (r *ClaimManager[O]) Claim(
	ctx context.Context,
	obj O,
) (bool, error) {
	switch claimState := r.strategy.ClaimState(r.claimer, obj); claimState {
	case ClaimStateTaken:
		// Claimed by someone else, ignore.
		return false, nil
	case ClaimStateClaimed:
		if r.selector.Match(obj) {
			// We own it and selector matches.
			// Even if we're deleting, we're allowed to own it.
			return true, nil
		}

		if !r.claimer.GetDeletionTimestamp().IsZero() {
			// We're already being deleted, don't try to release.
			return false, nil
		}

		// We own it but don't need it and are not being deleted - release it.
		if err := r.strategy.Release(ctx, r.claimer, obj); err != nil {
			if !apierrors.IsNotFound(err) {
				return false, err
			}
			// Ignore release error if it doesn't exist anymore.
			return false, nil
		}
		// Successfully released.
		return false, nil
	case ClaimStateFree:
		if !r.claimer.GetDeletionTimestamp().IsZero() || !r.selector.Match(obj) {
			// We are being deleted / don't want to claim it - skip it.
			return false, nil
		}
		if !obj.GetDeletionTimestamp().IsZero() {
			// Ignore it if it's being deleted
			return false, nil
		}

		if err := r.strategy.Adopt(ctx, r.claimer, obj); err != nil {
			if !apierrors.IsNotFound(err) {
				return false, err
			}
			// Ignore claim attempt if it wasn't found.
			return false, nil
		}

		// Successfully adopted.
		return true, nil
	default:
		return false, fmt.Errorf("invalid claim state %d", claimState)
	}
}
