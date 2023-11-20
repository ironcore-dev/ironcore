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

package server

import (
	"context"

	"github.com/blang/semver/v4"
	"github.com/ironcore-dev/ironcore/broker/machinebroker/version"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
)

func (s *Server) Version(ctx context.Context, req *iri.VersionRequest) (*iri.VersionResponse, error) {
	var runtimeVersion string
	switch {
	case version.Version != "":
		runtimeVersion = version.Version
	case version.Commit != "":
		v, err := semver.NewBuildVersion(version.Commit)
		if err != nil {
			runtimeVersion = "0.0.0"
		} else {
			runtimeVersion = v
		}
	default:
		runtimeVersion = "0.0.0"
	}

	return &iri.VersionResponse{
		RuntimeName:    version.RuntimeName,
		RuntimeVersion: runtimeVersion,
	}, nil
}
