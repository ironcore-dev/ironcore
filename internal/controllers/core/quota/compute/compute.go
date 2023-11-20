// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package compute

import (
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"github.com/ironcore-dev/ironcore/internal/controllers/core/quota/generic"
)

var (
	replenishReconcilersBuilder generic.ReplenishReconcilersBuilder
	NewReplenishReconcilers     = replenishReconcilersBuilder.NewReplenishReconcilers
)

func init() {
	replenishReconcilersBuilder.Register(
		&computev1alpha1.Machine{},
	)
}
