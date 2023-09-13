// Copyright 2023 OnMetal authors
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

package rbac

import (
	"context"

	"golang.org/x/exp/slices"
	"k8s.io/apiserver/pkg/endpoints/request"
)

// UserIsMemberOf returns true if the ctx contains user.Info and if the user is member of one of the provided groups
func UserIsMemberOf(ctx context.Context, groups []string) bool {
	if groups == nil {
		return true
	}

	user, ok := request.UserFrom(ctx)
	if !ok {
		return false
	}

	for _, group := range groups {
		if slices.Contains(user.GetGroups(), group) {
			return true
		}
	}
	return false
}
