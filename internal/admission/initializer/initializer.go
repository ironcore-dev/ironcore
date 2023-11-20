// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package initializer

import (
	ironcoreinformers "github.com/ironcore-dev/ironcore/client-go/informers"
	"github.com/ironcore-dev/ironcore/client-go/ironcore"
	"github.com/ironcore-dev/ironcore/utils/quota"
	"k8s.io/apiserver/pkg/admission"
)

type initializer struct {
	externalClient    ironcore.Interface
	externalInformers ironcoreinformers.SharedInformerFactory
	quotaRegistry     quota.Registry
}

func New(
	externalClient ironcore.Interface,
	externalInformers ironcoreinformers.SharedInformerFactory,
	quotaRegistry quota.Registry,
) admission.PluginInitializer {
	return &initializer{
		externalClient:    externalClient,
		externalInformers: externalInformers,
		quotaRegistry:     quotaRegistry,
	}
}

func (i *initializer) Initialize(plugin admission.Interface) {
	if wants, ok := plugin.(WantsExternalIronCoreClientSet); ok {
		wants.SetExternalIronCoreClientSet(i.externalClient)
	}

	if wants, ok := plugin.(WantsExternalInformers); ok {
		wants.SetExternalIronCoreInformerFactory(i.externalInformers)
	}

	if wants, ok := plugin.(WantsQuotaRegistry); ok {
		wants.SetQuotaRegistry(i.quotaRegistry)
	}
}

type WantsExternalIronCoreClientSet interface {
	SetExternalIronCoreClientSet(client ironcore.Interface)
	admission.InitializationValidator
}

type WantsExternalInformers interface {
	SetExternalIronCoreInformerFactory(f ironcoreinformers.SharedInformerFactory)
	admission.InitializationValidator
}

type WantsQuotaRegistry interface {
	SetQuotaRegistry(registry quota.Registry)
	admission.InitializationValidator
}
