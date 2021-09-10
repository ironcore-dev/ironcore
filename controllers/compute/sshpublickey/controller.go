/*
 * Copyright (c) 2021 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sshpublickey

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/onmetal/onmetal-api/pkg/utils"
	"github.com/onmetal/onmetal-api/pkg/utils/condition"
	"golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciler reconciles a SSHPublicKey object
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=compute.onmetal.de,resources=sshpublickeys,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=compute.onmetal.de,resources=sshpublickeys/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=compute.onmetal.de,resources=sshpublickeys/finalizers,verbs=update

// Reconcile reconciles the computev1alpha1.SSHPublicKey.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	sshPublicKey := &computev1alpha1.SSHPublicKey{}
	if err := r.Get(ctx, req.NamespacedName, sshPublicKey); err != nil {
		return utils.SucceededIfNotFound(err)
	}

	if !sshPublicKey.DeletionTimestamp.IsZero() {
		return r.delete(ctx, log, sshPublicKey)
	}
	return r.reconcile(ctx, log, sshPublicKey)
}

func (r *Reconciler) resolvePublicKeyData(ctx context.Context, sshPublicKey *computev1alpha1.SSHPublicKey) ([]byte, error) {
	ref := sshPublicKey.Spec.ConfigMapRef
	configMap := &corev1.ConfigMap{}
	configMapKey := client.ObjectKey{Namespace: sshPublicKey.Namespace, Name: ref.Name}
	if err := r.Get(ctx, configMapKey, configMap); err != nil {
		return nil, fmt.Errorf("error retrieving config map %s: %w", configMapKey, err)
	}

	dataKey := ref.Key
	if dataKey == "" {
		dataKey = computev1alpha1.DefaultSSHPublicKeyDataKey
	}

	keyData, ok := configMap.Data[dataKey]
	if !ok || keyData == "" {
		return nil, fmt.Errorf("public key data of config map %s key %s is empty", configMapKey, dataKey)
	}

	return []byte(keyData), nil
}

func publicKeyDataFingerprint(data []byte) string {
	h := md5.New()
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (r *Reconciler) reconcile(ctx context.Context, log logr.Logger, sshPublicKey *computev1alpha1.SSHPublicKey) (ctrl.Result, error) {
	mergeFromCurrent := client.MergeFrom(sshPublicKey.DeepCopy())

	log.V(1).Info("Resolving public key data")
	keyData, err := r.resolvePublicKeyData(ctx, sshPublicKey)
	if err != nil {
		condition.MustUpdateSlice(&sshPublicKey.Status.Conditions, string(computev1alpha1.SSHPublicKeyAvailable),
			condition.UpdateStatus(corev1.ConditionFalse),
			condition.UpdateReason("ResolutionFailed"),
			condition.UpdateMessage("Resolving the public key resulted in an error."),
		)
		if err := r.Status().Patch(ctx, sshPublicKey, mergeFromCurrent); err != nil {
			log.Error(err, "Error updating status")
		}
		return ctrl.Result{}, fmt.Errorf("error resolving public key data: %w", err)
	}

	log.V(1).Info("Parsing public key")
	key, _, _, _, err := ssh.ParseAuthorizedKey(keyData)
	if err != nil {
		condition.MustUpdateSlice(&sshPublicKey.Status.Conditions, string(computev1alpha1.SSHPublicKeyAvailable),
			condition.UpdateStatus(corev1.ConditionFalse),
			condition.UpdateReason("InvalidData"),
			condition.UpdateMessage("The key data is invalid."),
			condition.UpdateObserved(sshPublicKey),
		)
		if err := r.Status().Patch(ctx, sshPublicKey, mergeFromCurrent); err != nil {
			log.Error(err, "Error updating status")
		}
		return ctrl.Result{}, fmt.Errorf("error parsing public key: %w", err)
	}

	log.V(1).Info("Key is valid, updating status")
	canonicalData := key.Marshal()
	sshPublicKey.Status.Algorithm = key.Type()
	sshPublicKey.Status.KeyLength = len(canonicalData)
	sshPublicKey.Status.Fingerprint = publicKeyDataFingerprint(canonicalData)
	condition.MustUpdateSlice(&sshPublicKey.Status.Conditions, string(computev1alpha1.SSHPublicKeyAvailable),
		condition.UpdateStatus(corev1.ConditionTrue),
		condition.UpdateReason("Valid"),
		condition.UpdateMessage("The key is well-formed."),
		condition.UpdateObserved(sshPublicKey),
	)
	if err := r.Status().Patch(ctx, sshPublicKey, mergeFromCurrent); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating status: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) delete(ctx context.Context, log logr.Logger, sshPublicKey *computev1alpha1.SSHPublicKey) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&computev1alpha1.SSHPublicKey{}).
		Complete(r)
}
