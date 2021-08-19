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

package account

import (
	"context"
	api "github.com/onmetal/onmetal-api/apis/core/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Account webhook", func() {

	const (
		accountName        = "myaccount"
		accountDescription = "myaccount description"
		accountPurpose     = "development"

		//timeout  = time.Second * 10
		//interval = time.Second * 1
	)

	var account *api.Account
	//var accountLookUpKey types.NamespacedName

	Context("When creating an Account", func() {
		It("Should accept the Account creation", func() {
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
			//accountLookUpKey = types.NamespacedName{
			//	Name: accountName,
			//}
			Expect(k8sClient.Create(ctx, account)).Should(Succeed())

			//By(fmt.Sprintf("Expecting created and State set to %s", api.AccountStateInitial))
			//Eventually(func() bool {
			//	a := &api.Account{}
			//	if err := k8sClient.Get(context.Background(), accountLookUpKey, a); err != nil {
			//		return false
			//	}
			//	log.Info("account", "account status", a.Status)
			//	if a.Status.State == api.AccountStateInitial {
			//		return true
			//	}
			//	return false
			//}, timeout, interval).Should(BeTrue())
		})
	})
})
