// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package apiserver

import (
	"context"
	"strings"
	"testing"

	config "github.com/ironcore-dev/ironcore/utils/client/config"
)

func TestSetupMachinePoolletClientRotationGating(t *testing.T) {
	tests := []struct {
		name        string
		opts        config.GetConfigOptions
		wantErr     bool
		errContains string
	}{
		{
			name: "no rotation flags set: no-op, no error",
			opts: config.GetConfigOptions{},
		},
		{
			name:        "bootstrap-kubeconfig without rotate-certificates: error",
			opts:        config.GetConfigOptions{BootstrapKubeconfig: "/nonexistent/bootstrap-kubeconfig"},
			wantErr:     true,
			errContains: "requires --machinepoollet-client-rotate-certificates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &IronCoreAPIServerOptions{
				MachinePoolletClientGetConfig: tt.opts,
			}
			err := o.setupMachinePoolletClientRotation(context.Background())
			if tt.wantErr && err == nil {
				t.Error("expected an error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if tt.errContains != "" && (err == nil || !strings.Contains(err.Error(), tt.errContains)) {
				t.Errorf("expected error to contain %q, got %v", tt.errContains, err)
			}
		})
	}
}
