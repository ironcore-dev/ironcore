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

package accounts

import (
	"context"
	api "github.com/onmetal/onmetal-api/apis/core/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

var _ = Describe("Account controller", func() {

	log := logf.Log.WithName("Account controller test")

	const (
		accountName        = "myaccount"
		accountDescription = "myaccount description"
		accountPurpose     = "development"

		timeout         = time.Second * 30
		deletionTimeout = time.Second * 30
		interval        = time.Second * 1
	)

	var account *api.Account
	var accountLookUpKey types.NamespacedName
	var accountNamespace string

	Context("When creating an Account", func() {
		It("Should create a corresponding Account Namespace, Finalizer and set the correct Status", func() {
			ctx := context.Background()
			By("Creating a new Account")
			account = &api.Account{
				TypeMeta: metav1.TypeMeta{
					Kind:       api.AccountGK.Kind,
					APIVersion: api.AccountGK.Group,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: accountName,
				},
				Spec: api.AccountSpec{
					CreatedBy:   nil,
					Description: accountDescription,
					Owner:       nil,
					Purpose:     accountPurpose,
				},
			}
			accountLookUpKey = types.NamespacedName{
				Name: accountName,
			}
			Expect(k8sClient.Create(ctx, account)).Should(Succeed())

			By("Expecting created")
			Eventually(func() bool {
				a := &api.Account{}
				if err := k8sClient.Get(context.Background(), accountLookUpKey, a); err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("Expecting finalizer")
			Eventually(func() []string {
				a := &api.Account{}
				if err := k8sClient.Get(context.Background(), accountLookUpKey, a); err != nil {
					return []string{}
				}
				return a.GetFinalizers()
			}, timeout, interval).Should(ContainElements(accountFinilizerName))

			By("Expecting description")
			Eventually(func() string {
				a := &api.Account{}
				if err := k8sClient.Get(context.Background(), accountLookUpKey, a); err != nil {
					return ""
				}
				return a.Spec.Description
			}, timeout, interval).Should(Equal(accountDescription))

			By("Expecting purpose")
			Eventually(func() string {
				a := &api.Account{}
				if err := k8sClient.Get(context.Background(), accountLookUpKey, a); err != nil {
					return ""
				}
				return a.Spec.Purpose
			}, timeout, interval).Should(Equal(accountPurpose))

			By("Expecting Namespace in Status not to be empty")
			Eventually(func() string {
				a := &api.Account{}
				if err := k8sClient.Get(context.Background(), accountLookUpKey, a); err != nil {
					return ""
				}
				return a.Status.Namespace
			}, timeout, interval).Should(Not(BeEmpty()))

			By("Expecting Account Namespace to be created")
			Eventually(func() bool {
				a := &api.Account{}
				if err := k8sClient.Get(context.Background(), accountLookUpKey, a); err != nil {
					return false
				}
				accountNamespace = a.Status.Namespace
				n := &v1.Namespace{}
				if err := k8sClient.Get(context.Background(), client.ObjectKey{
					Name: accountNamespace,
				}, n); err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("Expecting the State to be Ready")
			Eventually(func() string {
				a := &api.Account{}
				if err := k8sClient.Get(context.Background(), accountLookUpKey, a); err != nil {
					return ""
				}
				return a.Status.State
			}, timeout, interval).Should(Equal(api.AccountReady))
		})
	})

	Context("When deleting an Account", func() {
		It("Should delete the Account Namespace and remove the Finalizer and Account object", func() {
			ctx := context.Background()

			By("Expecting deleting")
			// Delete the Account object and start the reconciliation loop
			Expect(k8sClient.Delete(ctx, account)).Should(Succeed())

			By("Expecting the Account Namespace to be in the Terminating state")
			Eventually(func() bool {
				n := &v1.Namespace{}
				err := k8sClient.Get(context.Background(), client.ObjectKey{
					Name: accountNamespace,
				}, n)
				if n.Status.Phase == v1.NamespaceTerminating || errors.IsNotFound(err) {
					log.Info("deleting account namespace", "account", account.Name, "namespace", n.Name)
					return true
				}
				return false
			}, deletionTimeout, interval).Should(BeTrue())
		})
	})
})
