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

package claimmanager

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClaimState uint8

const (
	ClaimStateInvalid ClaimState = iota
	ClaimStateFree
	ClaimStateClaimed
	ClaimStateTaken
)

type Selector interface {
	Match(obj client.Object) bool
}

type SelectorFunc func(obj client.Object) bool

func (f SelectorFunc) Match(obj client.Object) bool {
	return f(obj)
}

type everythingSelector struct{}

func (everythingSelector) Match(client.Object) bool {
	return true
}

var sharedEverythingSelector Selector = everythingSelector{}

func EverythingSelector() Selector {
	return sharedEverythingSelector
}

type nothingSelector struct{}

func (nothingSelector) Match(client.Object) bool {
	return false
}

var sharedNothingSelector Selector = nothingSelector{}

func NothingSelector() Selector {
	return sharedNothingSelector
}

type MatchingLabelSelector struct {
	Selector labels.Selector
}

func (s MatchingLabelSelector) Match(obj client.Object) bool {
	return s.Selector.Matches(labels.Set(obj.GetLabels()))
}

func MatchingLabels(lbls map[string]string) MatchingLabelSelector {
	return MatchingLabelSelector{Selector: labels.SelectorFromSet(lbls)}
}

type ClaimStrategy interface {
	ClaimState(claimer client.Object, obj client.Object) ClaimState
	Adopt(ctx context.Context, claimer client.Object, obj client.Object) error
	Release(ctx context.Context, claimer client.Object, obj client.Object) error
}

type ClaimStrategyFuncs struct {
	ClaimStateFunc func(claimer client.Object, obj client.Object) ClaimState
	AdoptFunc      func(ctx context.Context, claimer client.Object, obj client.Object) error
	ReleaseFunc    func(ctx context.Context, claimer client.Object, obj client.Object) error
}

func (f ClaimStrategyFuncs) ClaimState(claimer client.Object, obj client.Object) ClaimState {
	if f.ClaimStateFunc != nil {
		return f.ClaimStateFunc(claimer, obj)
	}
	return ClaimStateInvalid
}

func (f ClaimStrategyFuncs) Adopt(ctx context.Context, claimer client.Object, obj client.Object) error {
	if f.AdoptFunc != nil {
		return f.AdoptFunc(ctx, claimer, obj)
	}
	return nil
}

func (f ClaimStrategyFuncs) Release(ctx context.Context, claimer client.Object, obj client.Object) error {
	if f.ReleaseFunc != nil {
		return f.ReleaseFunc(ctx, claimer, obj)
	}
	return nil
}

type ClaimManager struct {
	claimer  client.Object
	selector Selector
	strategy ClaimStrategy
}

func New(
	claimer client.Object,
	selector Selector,
	strategy ClaimStrategy,
) *ClaimManager {
	return &ClaimManager{
		claimer:  claimer,
		selector: selector,
		strategy: strategy,
	}
}

func (r *ClaimManager) Claim(
	ctx context.Context,
	obj client.Object,
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
