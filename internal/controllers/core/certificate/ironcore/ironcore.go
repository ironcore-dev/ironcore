// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package ironcore

import (
	"github.com/ironcore-dev/ironcore/internal/controllers/core/certificate/compute"
	"github.com/ironcore-dev/ironcore/internal/controllers/core/certificate/generic"
	"github.com/ironcore-dev/ironcore/internal/controllers/core/certificate/networking"
	"github.com/ironcore-dev/ironcore/internal/controllers/core/certificate/storage"
)

var Recognizers []generic.CertificateSigningRequestRecognizer

func init() {
	Recognizers = append(Recognizers, compute.Recognizers...)
	Recognizers = append(Recognizers, storage.Recognizers...)
	Recognizers = append(Recognizers, networking.Recognizers...)
}
