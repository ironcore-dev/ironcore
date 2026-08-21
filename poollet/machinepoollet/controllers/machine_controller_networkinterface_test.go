// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	irimeta "github.com/ironcore-dev/ironcore/iri/apis/meta/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("patchNetworkInterfaceStatus", func() {
	// The no-op detection in patchNetworkInterfaceStatus compares two statuses containing
	// commonv1alpha1.IP / commonv1alpha1.IPPrefix (which embed netip.Addr with unexported
	// fields). The comparison must not panic on those.
	It("skips unchanged values", func() {
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
		Expect(patchNetworkInterfaceStatus(context.Background(), nil, nic, networkingv1alpha1.NetworkInterfaceStatus{
			State:     networkingv1alpha1.NetworkInterfaceStateAvailable,
			IPs:       []commonv1alpha1.IP{*commonv1alpha1.MustParseNewIP("10.0.0.11")},
			Prefixes:  []commonv1alpha1.IPPrefix{commonv1alpha1.MustParseIPPrefix("10.0.0.0/24")},
			VirtualIP: commonv1alpha1.MustParseNewIP("10.0.1.1"),
		})).To(Succeed())

		Expect(nic.Status.LastStateTransitionTime).To(BeIdenticalTo(&transitionTime))
	})
})

var _ = Describe("iriNetworkInterfaceUpToDate", func() {
	// Metadata must not cause an endless detach/attach loop with providers that do not know
	// (and thus never echo) the metadata field.
	metadata := &irimeta.ObjectMetadata{
		Id:     "some-uid",
		Labels: map[string]string{"foo": "bar"},
	}

	DescribeTable("reports whether the existing IRI network interface matches the desired one",
		func(desired, existing *iri.NetworkInterface, upToDate bool) {
			Expect(iriNetworkInterfaceUpToDate(desired, existing)).To(Equal(upToDate))
		},
		Entry("no metadata on either side",
			&iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.1"}},
			&iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.1"}},
			true,
		),
		Entry("desired has metadata, provider does not report it",
			&iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.1"}, Metadata: metadata},
			&iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.1"}},
			true,
		),
		Entry("provider echoes equal metadata",
			&iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.1"}, Metadata: metadata},
			&iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.1"}, Metadata: metadata},
			true,
		),
		Entry("provider reports differing metadata",
			&iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.1"}, Metadata: metadata},
			&iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.1"}, Metadata: &irimeta.ObjectMetadata{
				Id:     "some-uid",
				Labels: map[string]string{"foo": "changed"},
			}},
			false,
		),
		Entry("content drift is detected despite absent metadata",
			&iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.2"}, Metadata: metadata},
			&iri.NetworkInterface{Name: "nic", NetworkId: "network", Ips: []string{"10.0.0.1"}},
			false,
		),
	)
})
