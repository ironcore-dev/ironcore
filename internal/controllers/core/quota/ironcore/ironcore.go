// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package ironcore

import (
	"github.com/ironcore-dev/ironcore/internal/controllers/core/quota/compute"
	"github.com/ironcore-dev/ironcore/internal/controllers/core/quota/generic"
	"github.com/ironcore-dev/ironcore/internal/controllers/core/quota/storage"
)

var (
	replenishReconcilersBuilder generic.ReplenishReconcilersBuilder
	NewReplenishReconcilers     = replenishReconcilersBuilder.NewReplenishReconcilers
)

func init() {
	replenishReconcilersBuilder.Add(
		compute.NewReplenishReconcilers,
		storage.NewReplenishReconcilers,
	)
}
