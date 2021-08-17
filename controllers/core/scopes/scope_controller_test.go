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

package scopes

import (
	"context"
	api "github.com/onmetal/onmetal-api/apis/core/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var _ = Describe("Scope controller", func() {

	const (
		scopeName        = "myscope"
		scopeDescription = "myaccount description"
		scopeRegion      = "myregion"

		accountName        = "myaccount"
		accountDescription = "myaccount description"
		accountPurpose     = "development"

		timeout  = time.Second * 30
		interval = time.Second * 1
	)

	var scope *api.Scope
	var scopeLookUpKey types.NamespacedName
	var scopeNamespace string

	var account *api.Account
	var accountLookUpKey types.NamespacedName
	var accountNamespace string

	//TODO: factor out Account creation into a BeforeEach

	Context("When creating an Account and a Scope", func() {
		It("Should create a corresponding Scope Namespace, Finalizer and set the correct Status", func() {
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

			By("Creating a new Scope")
			scope = &api.Scope{
				TypeMeta: metav1.TypeMeta{
					Kind:       api.AccountGK.Kind,
					APIVersion: api.AccountGK.Group,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      scopeName,
					Namespace: accountNamespace,
				},
				Spec: api.ScopeSpec{
					Description: scopeDescription,
					Region:      scopeRegion,
				},
			}
			scopeLookUpKey = types.NamespacedName{
				Name:      scopeName,
				Namespace: accountNamespace,
			}
			Expect(k8sClient.Create(ctx, scope)).Should(Succeed())

			By("Expecting created")
			Eventually(func() bool {
				s := &api.Scope{}
				if err := k8sClient.Get(context.Background(), scopeLookUpKey, s); err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("Expecting description")
			Eventually(func() string {
				s := &api.Scope{}
				if err := k8sClient.Get(context.Background(), scopeLookUpKey, s); err != nil {
					return ""
				}
				return s.Spec.Description
			}, timeout, interval).Should(Equal(scopeDescription))

			By("Expecting region")
			Eventually(func() string {
				s := &api.Scope{}
				if err := k8sClient.Get(context.Background(), scopeLookUpKey, s); err != nil {
					return ""
				}
				return s.Spec.Region
			}, timeout, interval).Should(Equal(scopeRegion))

			By("Expecting Namespace in Status not to be empty")
			Eventually(func() string {
				s := &api.Scope{}
				if err := k8sClient.Get(context.Background(), scopeLookUpKey, s); err != nil {
					return ""
				}
				return s.Status.Namespace
			}, timeout, interval).Should(Not(BeEmpty()))

			By("Expecting Scope Namespace to be created")
			Eventually(func() bool {
				s := &api.Scope{}
				if err := k8sClient.Get(context.Background(), scopeLookUpKey, s); err != nil {
					return false
				}
				scopeNamespace = s.Status.Namespace
				n := &v1.Namespace{}
				if err := k8sClient.Get(context.Background(), client.ObjectKey{
					Name: scopeNamespace,
				}, n); err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			By("Expecting the State to be Ready")
			Eventually(func() string {
				s := &api.Scope{}
				if err := k8sClient.Get(context.Background(), scopeLookUpKey, s); err != nil {
					return ""
				}
				return s.Status.State
			}, timeout, interval).Should(Equal(api.AccountReady))

			By("Expecting the Account name in State to be set")
			Eventually(func() string {
				s := &api.Scope{}
				if err := k8sClient.Get(context.Background(), scopeLookUpKey, s); err != nil {
					return ""
				}
				return s.Status.Account
			}, timeout, interval).Should(Equal(accountName))

			By("Expecting the ParentScope in State to be set")
			Eventually(func() string {
				s := &api.Scope{}
				if err := k8sClient.Get(context.Background(), scopeLookUpKey, s); err != nil {
					return ""
				}
				return s.Status.ParentScope
			}, timeout, interval).Should(Equal(accountName))

			By("Expecting the ParentNamespace in State to be set")
			Eventually(func() string {
				s := &api.Scope{}
				if err := k8sClient.Get(context.Background(), scopeLookUpKey, s); err != nil {
					return ""
				}
				return s.Status.ParentNamespace
			}, timeout, interval).Should(Equal(accountNamespace))

			By("Expecting finalizer")
			Eventually(func() []string {
				s := &api.Scope{}
				if err := k8sClient.Get(context.Background(), scopeLookUpKey, s); err != nil {
					return []string{}
				}
				return s.GetFinalizers()
			}, timeout, interval).Should(ContainElements(scopeFinilizerName))
		})
	})

	//TODO:
	// Delete Scope -> Terminating namespace
	// Delete Account -> Remove scope + terminating account and scope namespaces
})
