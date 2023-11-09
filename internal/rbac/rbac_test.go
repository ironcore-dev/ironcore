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
	"testing"

	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"
)

func TestUserIsMemberOf(t *testing.T) {
	type args struct {
		ctx    context.Context
		groups []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "User is always part of empty groups - 1",
			args: args{
				ctx:    request.WithUser(request.NewContext(), &user.DefaultInfo{}),
				groups: nil,
			},
			want: true,
		},
		{
			name: "User is always part of empty groups - 2",
			args: args{
				ctx: request.WithUser(request.NewContext(), &user.DefaultInfo{
					Name:   "test",
					Groups: []string{"group-1"},
					Extra:  nil,
				}),
				groups: nil,
			},
			want: true,
		},
		{
			name: "User is member of one of the required groups",
			args: args{
				ctx: request.WithUser(request.NewContext(), &user.DefaultInfo{
					Name:   "test",
					Groups: []string{"group-1"},
					Extra:  nil,
				}),
				groups: []string{"group-1", "group-2", "group-3"},
			},
			want: true,
		},
		{
			name: "User is not member of one of the required groups",
			args: args{
				ctx: request.WithUser(request.NewContext(), &user.DefaultInfo{
					Name:   "test",
					Groups: []string{"group-1"},
					Extra:  nil,
				}),
				groups: []string{"group-2", "group-3"},
			},
			want: false,
		},
		{
			name: "Context without user is not part of groups",
			args: args{
				ctx:    request.NewContext(),
				groups: []string{"group-1", "group-2", "group-3"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UserIsMemberOf(tt.args.ctx, tt.args.groups); got != tt.want {
				t.Errorf("UserIsMemberOf() = %v, want %v", got, tt.want)
			}
		})
	}
}
