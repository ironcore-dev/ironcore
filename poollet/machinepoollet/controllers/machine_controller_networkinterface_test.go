// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"testing"

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// The no-op detection in patchNetworkInterfaceStatus compares two statuses containing
// commonv1alpha1.IP / commonv1alpha1.IPPrefix (which embed netip.Addr with unexported
// fields). The comparison must not panic on those.
func TestPatchNetworkInterfaceStatusSkipsUnchangedValues(t *testing.T) {
	transitionTime := metav1.Now()
	nic := &networkingv1alpha1.NetworkInterface{
		Status: networkingv1alpha1.NetworkInterfaceStatus{
			State:                   networkingv1alpha1.NetworkInterfaceStateAvailable,
			LastStateTransitionTime: &transitionTime,
			IPs:                     []commonv1alpha1.IP{*commonv1alpha1.MustParseNewIP("10.0.0.11")},
			Prefixes:                []commonv1alpha1.IPPrefix{commonv1alpha1.MustParseIPPrefix("10.0.0.0/24")},
			VirtualIP:               commonv1alpha1.MustParseNewIP("10.0.1.1"),
		},
	}

	// Passing a nil client: if the status is considered unchanged, no client interaction happens.
	if err := patchNetworkInterfaceStatus(context.Background(), nil, nic, networkingv1alpha1.NetworkInterfaceStatus{
		State:     networkingv1alpha1.NetworkInterfaceStateAvailable,
		IPs:       []commonv1alpha1.IP{*commonv1alpha1.MustParseNewIP("10.0.0.11")},
		Prefixes:  []commonv1alpha1.IPPrefix{commonv1alpha1.MustParseIPPrefix("10.0.0.0/24")},
		VirtualIP: commonv1alpha1.MustParseNewIP("10.0.1.1"),
	}); err != nil {
		t.Fatalf("patchNetworkInterfaceStatus returned error: %v", err)
	}

	if nic.Status.LastStateTransitionTime != &transitionTime {
		t.Errorf("LastStateTransitionTime was unexpectedly changed: %v", nic.Status.LastStateTransitionTime)
	}
}

// Metadata must not cause an endless detach/attach loop with providers that do not know
// (and thus never echo) the metadata field.
func TestIRINetworkInterfaceUpToDate(t *testing.T) {
	metadata := &irimeta.ObjectMetadata{
		Id:     "some-uid",
		Labels: map[string]string{"foo": "bar"},
	}

	tests := []struct {
		name     string
		desired  *iri.NetworkInterface
		existing *iri.NetworkInterface
		upToDate bool
	}{
		{
			name:     "no metadata on either side",
			desired:  &iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.1"}},
			existing: &iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.1"}},
			upToDate: true,
		},
		{
			name:     "desired has metadata, provider does not report it",
			desired:  &iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.1"}, Metadata: metadata},
			existing: &iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.1"}},
			upToDate: true,
		},
		{
			name:     "provider echoes equal metadata",
			desired:  &iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.1"}, Metadata: metadata},
			existing: &iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.1"}, Metadata: metadata},
			upToDate: true,
		},
		{
			name:    "provider reports differing metadata",
			desired: &iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.1"}, Metadata: metadata},
			existing: &iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.1"}, Metadata: &irimeta.ObjectMetadata{
				Id:     "some-uid",
				Labels: map[string]string{"foo": "changed"},
			}},
			upToDate: false,
		},
		{
			name:     "content drift is detected despite absent metadata",
			desired:  &iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.2"}, Metadata: metadata},
			existing: &iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.1"}},
			upToDate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := iriNetworkInterfaceUpToDate(tt.desired, tt.existing); got != tt.upToDate {
				t.Errorf("iriNetworkInterfaceUpToDate() = %v, want %v", got, tt.upToDate)
			}
		})
	}
}
