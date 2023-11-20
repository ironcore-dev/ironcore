// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package debug

import ctrl "sigs.k8s.io/controller-runtime"

var (
	log = ctrl.Log.WithName("debug")

	handlerLog = log.WithName("handler")

	predicateLog = log.WithName("predicate")
)
