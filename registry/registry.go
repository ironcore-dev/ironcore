// Copyright 2022 OnMetal authors
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

package registry

import (
	"os"

	"k8s.io/apiserver/pkg/registry/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logf = ctrl.Log.WithName("registry")
)

func RESTOrDie(storage rest.StandardStorage, err error) rest.StandardStorage {
	if err != nil {
		logf.Error(err, "Unable to create REST storage for a resource")
		os.Exit(1)
	}
	return storage
}
