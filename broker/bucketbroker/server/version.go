// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	"github.com/blang/semver/v4"
	"github.com/ironcore-dev/ironcore/broker/bucketbroker/version"
	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
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
