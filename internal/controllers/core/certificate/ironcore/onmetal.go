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
