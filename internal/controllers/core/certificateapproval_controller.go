// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"context"
	"fmt"

	"github.com/ironcore-dev/ironcore/internal/controllers/core/certificate/generic"
	authv1 "k8s.io/api/authorization/v1"
	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CertificateApprovalReconciler struct {
	client.Client

	Recognizers []generic.CertificateSigningRequestRecognizer
}

//+kubebuilder:rbac:groups=certificates.k8s.io,resources=signers,resourceNames=kubernetes.io/kube-apiserver-client,verbs=approve
//+kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests,verbs=get;list;watch
//+kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests/approval,verbs=get;update;patch
//+kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=create

func (r *CertificateApprovalReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	csr := &certificatesv1.CertificateSigningRequest{}
	if err := r.Get(ctx, req.NamespacedName, csr); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if len(csr.Status.Certificate) > 0 {
		log.V(1).Info("Certificate already present, nothing to do")
		return ctrl.Result{}, nil
	}

	if approved, denied := generic.GetCertificateSigningRequestApprovalCondition(&csr.Status); approved || denied {
		log.V(1).Info("Certificate approval already marked", "Approved", approved, "Denied", denied)
		return ctrl.Result{}, nil
	}

	x509CR, err := generic.ParseCertificateRequest(csr.Spec.Request)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error parsing csr: %w", err)
	}

	var tried []string
	for _, recognizer := range r.Recognizers {
		if recognizer.Recognize(csr, x509CR) {
			continue
		}

		permission := recognizer.Permission()
		tried = append(tried, permission.Subresource)

		approved, err := r.authorize(ctx, csr, permission)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("error approving permission subresource %s: %w",
				permission.Subresource,
				err,
			)
		}
		if approved {
			appendApprovalCondition(csr, recognizer.SuccessMessage())
			if err := r.Client.SubResource("approval").Update(ctx, csr); err != nil {
				return ctrl.Result{}, fmt.Errorf("error updating approval for certificate signing request: %w", err)
			}

			log.V(1).Info("Approved certificate", "Subresource", permission.Subresource)
			return ctrl.Result{}, nil
		}
	}
	if len(tried) > 0 {
		log.V(1).Info("Recognized certificate signing request but access review was not approved", "Tried", tried)
	}

	return ctrl.Result{}, nil
}

func (r *CertificateApprovalReconciler) authorize(ctx context.Context, csr *certificatesv1.CertificateSigningRequest, attrs authv1.ResourceAttributes) (bool, error) {
	extra := make(map[string]authv1.ExtraValue)
	for k, v := range csr.Spec.Extra {
		extra[k] = authv1.ExtraValue(v)
	}

	sar := &authv1.SubjectAccessReview{
		Spec: authv1.SubjectAccessReviewSpec{
			User:               csr.Spec.Username,
			UID:                csr.Spec.UID,
			Groups:             csr.Spec.Groups,
			Extra:              extra,
			ResourceAttributes: &attrs,
		},
	}
	if err := r.Create(ctx, sar); err != nil {
		return false, fmt.Errorf("error creating subject access review: %w", err)
	}
	return sar.Status.Allowed, nil
}

func (r *CertificateApprovalReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&certificatesv1.CertificateSigningRequest{}).
		Complete(r)
}

func appendApprovalCondition(csr *certificatesv1.CertificateSigningRequest, message string) {
	csr.Status.Conditions = append(csr.Status.Conditions, certificatesv1.CertificateSigningRequestCondition{
		Type:    certificatesv1.CertificateApproved,
		Status:  corev1.ConditionTrue,
		Reason:  "AutoApproved",
		Message: message,
	})
}
