// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package certificate

import (
	"crypto/x509/pkix"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
)

// MachinePoolletClientSignerName routes the apiserver machinepoollet-client CSR
// to the cert-manager CA ClusterIssuer named "ironcore-machinepoollet-client".
// cert-manager's CSR controller signs k8s CertificateSigningRequests whose
// spec.signerName is clusterissuers.cert-manager.io/<clusterissuer-name>, and
// only after the CSR carries an Approved condition (set by the ironcore
// recognizer). The cert is thus signed by the ironcore client CA, not the shoot
// cluster CA.
const MachinePoolletClientSignerName = "clusterissuers.cert-manager.io/ironcore-machinepoollet-client"

// APIServerMachinePoolletClientCSRSubject returns the CSR subject for the
// apiserver's machinepoollet client cert. It is shared by the rotator and the
// recognizer so both agree on the exact identity. The organization is the
// dedicated ironcore group (authorized by the poollet's delegating SAR for
// machinepools/proxy), not the Kubernetes superuser group system:masters.
func APIServerMachinePoolletClientCSRSubject() pkix.Name {
	return pkix.Name{
		CommonName:   computev1alpha1.APIServerMachinePoolletClientCommonName,
		Organization: []string{computev1alpha1.APIServerMachinePoolletClientsGroup},
	}
}
