// Copyright 2023 IronCore authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resourcequota

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	"github.com/ironcore-dev/ironcore/client-go/ironcore"
	corev1alpha1listers "github.com/ironcore-dev/ironcore/client-go/listers/core/v1alpha1"
	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/utils/lru"
)

type QuotaAccessor interface {
	List(ctx context.Context, namespace string) ([]corev1alpha1.ResourceQuota, error)
	UpdateStatus(ctx context.Context, newQuota, oldQuota *corev1alpha1.ResourceQuota) error
}

type quotaAccessor struct {
	client ironcore.Interface

	lister corev1alpha1listers.ResourceQuotaLister

	liveLookupCache *lru.Cache
	liveTTL         time.Duration

	updatedQuotas *lru.Cache
}

func NewQuotaAccessor(
	client ironcore.Interface,
	lister corev1alpha1listers.ResourceQuotaLister,
) (QuotaAccessor, error) {
	if client == nil {
		return nil, fmt.Errorf("must specify client")
	}
	if lister == nil {
		return nil, fmt.Errorf("must specify lister")
	}

	return &quotaAccessor{
		client:          client,
		lister:          lister,
		liveLookupCache: lru.New(100),
		liveTTL:         30 * time.Second,
		updatedQuotas:   lru.New(100),
	}, nil
}

type liveLookupEntry struct {
	expiry time.Time
	items  []*corev1alpha1.ResourceQuota
}

var apiObjectVersioner = storage.APIObjectVersioner{}

func (a *quotaAccessor) checkCache(quota *corev1alpha1.ResourceQuota) *corev1alpha1.ResourceQuota {
	key := a.quotaKey(quota)
	uncastCachedQuota, ok := a.updatedQuotas.Get(key)
	if !ok {
		return quota
	}

	cachedQuota := uncastCachedQuota.(*corev1alpha1.ResourceQuota)
	if apiObjectVersioner.CompareResourceVersion(quota, cachedQuota) >= 0 {
		a.updatedQuotas.Remove(key)
		return quota
	}
	return cachedQuota
}

func (a *quotaAccessor) quotaKey(quota *corev1alpha1.ResourceQuota) string {
	var sb strings.Builder
	sb.WriteString(quota.Namespace)
	sb.WriteString("/")
	sb.WriteString(quota.Name)
	return sb.String()
}

func (a *quotaAccessor) List(ctx context.Context, namespace string) ([]corev1alpha1.ResourceQuota, error) {
	items, err := a.lister.ResourceQuotas(namespace).List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("error listing quotas: %w", err)
	}

	if len(items) == 0 {
		lruItemObj, ok := a.liveLookupCache.Get(namespace)
		if !ok || lruItemObj.(liveLookupEntry).expiry.Before(time.Now()) {
			liveList, err := a.client.CoreV1alpha1().ResourceQuotas(namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				return nil, err
			}

			newEntry := liveLookupEntry{
				expiry: time.Now().Add(a.liveTTL),
				items:  make([]*corev1alpha1.ResourceQuota, len(liveList.Items)),
			}
			for i := range liveList.Items {
				newEntry.items[i] = &liveList.Items[i]
			}
			a.liveLookupCache.Add(namespace, newEntry)
			lruItemObj = newEntry
		}

		lruEntry := lruItemObj.(liveLookupEntry)
		items = slices.Clone(lruEntry.items)
	}

	res := make([]corev1alpha1.ResourceQuota, len(items))
	for i := range items {
		quota := items[i]
		quota = a.checkCache(quota)
		res[i] = *quota
	}
	return res, nil
}

func (a *quotaAccessor) UpdateStatus(ctx context.Context, newQuota, oldQuota *corev1alpha1.ResourceQuota) error {
	updatedQuota, err := a.client.CoreV1alpha1().ResourceQuotas(newQuota.Namespace).UpdateStatus(ctx, newQuota, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	key := a.quotaKey(updatedQuota)
	a.updatedQuotas.Add(key, updatedQuota)
	return nil
}
