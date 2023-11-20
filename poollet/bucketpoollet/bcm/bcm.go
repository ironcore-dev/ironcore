// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package bcm

import (
	"context"
	"errors"

	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	ErrNoMatchingBucketClass        = errors.New("no matching bucket class")
	ErrAmbiguousMatchingBucketClass = errors.New("ambiguous matching bucket classes")
)

type BucketClassMapper interface {
	manager.Runnable
	GetBucketClassFor(ctx context.Context, name string, capabilities *iri.BucketClassCapabilities) (*iri.BucketClass, error)
	WaitForSync(ctx context.Context) error
}
