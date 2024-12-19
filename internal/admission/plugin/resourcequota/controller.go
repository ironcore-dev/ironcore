// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package resourcequota

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	"github.com/ironcore-dev/ironcore/utils/quota"
	ironcoreutilruntime "github.com/ironcore-dev/ironcore/utils/runtime"
	utilslices "github.com/ironcore-dev/ironcore/utils/slices"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Evaluator interface {
	Evaluate(ctx context.Context, a admission.Attributes) error
}

type EvaluatorController struct {
	workLock sync.Mutex

	quotaAccessor QuotaAccessor
	registry      quota.Registry

	queue      workqueue.TypedInterface[string]
	inProgress sets.Set[string]
	work       map[string][]*admissionWaiter
	dirtyWork  map[string][]*admissionWaiter
}

func NewEvaluatorController(quotaAccessor QuotaAccessor, registry quota.Registry) *EvaluatorController {
	return &EvaluatorController{
		quotaAccessor: quotaAccessor,
		registry:      registry,
		queue:         workqueue.NewTypedWithConfig[string](workqueue.TypedQueueConfig[string]{Name: "quota_admission_evaluator"}),
		inProgress:    sets.New[string](),
		work:          make(map[string][]*admissionWaiter),
		dirtyWork:     make(map[string][]*admissionWaiter),
	}
}

func (e *EvaluatorController) Evaluate(ctx context.Context, a admission.Attributes) error {
	obj, _, err := getObjects(a)
	if err != nil {
		return err
	}

	eval, err := e.registry.Get(obj)
	if err != nil || eval == nil {
		return err
	}

	waiter := &admissionWaiter{
		attributes: a,
		result:     defaultDeny{},
		finished:   make(chan struct{}),
	}
	e.addWork(waiter)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-waiter.finished:
		return waiter.result
	}
}

type admissionWaiter struct {
	attributes admission.Attributes
	result     error
	finished   chan struct{}
}

func (e *EvaluatorController) check(ctx context.Context, ns string, waiters []*admissionWaiter) {
	// Signal completion on exit
	defer func() {
		for _, waiter := range waiters {
			close(waiter.finished)
		}
	}()

	resourceQuotas, err := e.quotaAccessor.List(ctx, ns)
	if err != nil {
		err = fmt.Errorf("error getting resource quotas in namespace %s: %w", ns, err)
		for _, waiter := range waiters {
			waiter.result = err
		}
		return
	}

	e.checkResourceQuotas(ctx, resourceQuotas, waiters, 3)
}

func (e *EvaluatorController) checkResourceQuotas(ctx context.Context, quotas []corev1alpha1.ResourceQuota, waiters []*admissionWaiter, retries int) {
	originalQuotas := ironcoreutilruntime.DeepCopySliceRefs(quotas)

	var atLeastOneChanged bool
	for _, waiter := range waiters {
		newQuotas, err := e.checkRequest(ctx, quotas, waiter.attributes)
		if err != nil {
			waiter.result = err
			continue
		}

		if waiter.attributes.IsDryRun() {
			waiter.result = nil
			continue
		}

		var atLeastOneChangedWaiter bool
		for j := range newQuotas {
			if !quota.Equals(quotas[j].Status.Used, newQuotas[j].Status.Used) {
				atLeastOneChanged = true
				atLeastOneChangedWaiter = true
				break
			}
		}

		if !atLeastOneChangedWaiter {
			waiter.result = nil
		}

		quotas = newQuotas
	}

	if !atLeastOneChanged {
		return
	}

	var (
		updateFailedQuotaNames = sets.New[string]()
		lastErr                error
	)
	for i := range quotas {
		newQuota := quotas[i]

		if quota.Equals(originalQuotas[i].Status.Used, newQuota.Status.Used) {
			continue
		}

		if err := e.quotaAccessor.UpdateStatus(ctx, &newQuota, &originalQuotas[i]); err != nil {
			updateFailedQuotaNames.Insert(newQuota.Name)
			lastErr = err
		}
	}

	if len(updateFailedQuotaNames) == 0 {
		for _, waiter := range waiters {
			if isDefaultDeny(waiter.result) {
				waiter.result = nil
			}
		}
		return
	}

	if retries <= 0 {
		for _, waiter := range waiters {
			if isDefaultDeny(waiter.result) {
				waiter.result = lastErr
			}
		}
		return
	}

	quotas, err := e.quotaAccessor.List(ctx, quotas[0].Namespace)
	if err != nil {
		for _, waiter := range waiters {
			if isDefaultDeny(waiter.result) {
				waiter.result = lastErr
			}
		}
		return
	}

	quotasToCheck := utilslices.FilterFunc(quotas, func(quota corev1alpha1.ResourceQuota) bool {
		return updateFailedQuotaNames.Has(quota.Name)
	})
	e.checkResourceQuotas(ctx, quotasToCheck, waiters, retries-1)
}

func (e *EvaluatorController) addWork(a *admissionWaiter) {
	e.workLock.Lock()
	defer e.workLock.Unlock()

	ns := a.attributes.GetNamespace()
	e.queue.Add(ns)

	if e.inProgress.Has(ns) {
		e.dirtyWork[ns] = append(e.dirtyWork[ns], a)
		return
	}

	e.work[ns] = append(e.work[ns], a)
}

func (e *EvaluatorController) completeWork(ns string) {
	e.workLock.Lock()
	defer e.workLock.Unlock()

	e.queue.Done(ns)
	e.work[ns] = e.dirtyWork[ns]
	delete(e.dirtyWork, ns)
	e.inProgress.Delete(ns)
}

func (e *EvaluatorController) getWork() (string, []*admissionWaiter, bool) {
	ns, shutdown := e.queue.Get()
	if shutdown {
		return "", nil, true
	}

	e.workLock.Lock()
	defer e.workLock.Unlock()

	work := e.work[ns]
	delete(e.work, ns)
	delete(e.dirtyWork, ns)
	e.inProgress.Insert(ns)
	return ns, work, false
}

func (e *EvaluatorController) Start(ctx context.Context) error {
	go func() {
		defer e.queue.ShutDown()
		<-ctx.Done()
	}()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wait.UntilWithContext(ctx, e.doWork, 1*time.Second)
		}()
	}

	wg.Wait()
	return nil
}

func (e *EvaluatorController) doWork(ctx context.Context) {
	workFunc := func() bool {
		ns, admissionAttributes, quit := e.getWork()
		if quit {
			return true
		}
		defer e.completeWork(ns)
		if len(admissionAttributes) == 0 {
			return false
		}
		e.check(ctx, ns, admissionAttributes)
		return false
	}
	for {
		if quit := workFunc(); quit {
			return
		}
	}
}

type defaultDeny struct{}

func (defaultDeny) Error() string {
	return "DEFAULT DENY"
}

func isDefaultDeny(err error) bool {
	if err == nil {
		return false
	}

	_, ok := err.(defaultDeny)
	return ok
}

func getObjects(a admission.Attributes) (obj, oldObj client.Object, err error) {
	obj, ok := a.GetObject().(client.Object)
	if !ok {
		return nil, nil, fmt.Errorf("expected client.Object for Object but got %v", a.GetObject())
	}

	oldRuntimeObj := a.GetOldObject()
	if oldRuntimeObj != nil {
		old, ok := oldRuntimeObj.(client.Object)
		if !ok {
			return nil, nil, fmt.Errorf("expected client.Object for OldObject but got %v", oldRuntimeObj)
		}

		oldObj = old
	}

	return obj, oldObj, nil
}

func (e *EvaluatorController) checkRequest(ctx context.Context, quotas []corev1alpha1.ResourceQuota, a admission.Attributes) ([]corev1alpha1.ResourceQuota, error) {
	obj, oldObj, err := getObjects(a)
	if err != nil {
		return nil, err
	}

	evaluator, err := e.registry.Get(obj)
	if err != nil || evaluator == nil {
		return quotas, err
	}

	var indexes []int
	for i, resourceQuota := range quotas {
		match, err := quota.EvaluatorMatchesResourceScopeSelector(evaluator, obj, resourceQuota.Spec.ScopeSelector)
		if err != nil {
			return nil, fmt.Errorf("error matching scopes of quota %s: %w", resourceQuota.Name, err)
		}
		if !match {
			continue
		}

		if !quota.EvaluatorMatchesResourceList(evaluator, resourceQuota.Status.Hard) {
			continue
		}

		hardResources := quota.ResourceNames(resourceQuota.Status.Hard)
		restrictedResources := quota.EvaluatorMatchingResourceNames(evaluator, hardResources)
		if !hasUsageStats(&resourceQuota, restrictedResources) {
			return nil, admission.NewForbidden(a, fmt.Errorf("status unknown for quota: %s, resources: %s", resourceQuota.Name, prettyPrintResourceNames(restrictedResources)))
		}

		indexes = append(indexes, i)
	}

	deltaUsage, err := evaluator.Usage(ctx, obj)
	if err != nil {
		return nil, fmt.Errorf("error determining usage: %w", err)
	}
	if negativeUsage := quota.IsNegative(deltaUsage); negativeUsage.Len() > 0 {
		return nil, admission.NewForbidden(a, fmt.Errorf("quota usage is negative for resource(s): %v", sets.List(negativeUsage)))
	}

	if oldObj != nil {
		prevUsage, err := evaluator.Usage(ctx, oldObj)
		if err != nil {
			return nil, fmt.Errorf("error determining old usage: %w", err)
		}

		deltaUsage = quota.SubtractWithNonNegativeResult(deltaUsage, prevUsage)
	}

	deltaUsage = quota.RemoveZeros(deltaUsage)
	if len(deltaUsage) == 0 {
		return quotas, nil
	}

	if len(indexes) == 0 {
		return quotas, nil
	}

	outQuotas := ironcoreutilruntime.DeepCopySliceRefs(quotas)
	for _, i := range indexes {
		resourceQuota := outQuotas[i]

		hardResourceNames := quota.ResourceNames(resourceQuota.Status.Hard)
		maskedDeltaUsage := quota.Mask(deltaUsage, hardResourceNames)
		newUsage := quota.Add(resourceQuota.Status.Used, maskedDeltaUsage)
		maskedNewUsage := quota.Mask(newUsage, quota.ResourceNames(maskedDeltaUsage))

		if allowed, exceeded := quota.LessThanOrEqual(maskedNewUsage, resourceQuota.Status.Hard); !allowed {
			failedRequestedUsage := quota.Mask(maskedDeltaUsage, exceeded)
			failedUsed := quota.Mask(resourceQuota.Status.Used, exceeded)
			failedHard := quota.Mask(resourceQuota.Status.Hard, exceeded)
			return nil, admission.NewForbidden(a,
				fmt.Errorf("exceeded quota: %s, requested: %s, used: %s, limited: %s",
					resourceQuota.Name,
					prettyPrint(failedRequestedUsage),
					prettyPrint(failedUsed),
					prettyPrint(failedHard)))
		}

		outQuotas[i].Status.Used = newUsage
	}
	return outQuotas, nil
}

// prettyPrint formats a resource list for usage in errors
// it outputs resources sorted in increasing order
func prettyPrint(item corev1alpha1.ResourceList) string {
	keys := make([]string, 0, len(item))
	for key := range item {
		keys = append(keys, string(key))
	}
	sort.Strings(keys)

	parts := make([]string, len(keys))
	for i, key := range keys {
		value := item[corev1alpha1.ResourceName(key)]
		constraint := fmt.Sprintf("%s=%s", key, &value)
		parts[i] = constraint
	}
	return strings.Join(parts, ",")
}

// hasUsageStats returns true if for each hard constraint in interestingResources there is a value for its current usage
func hasUsageStats(resourceQuota *corev1alpha1.ResourceQuota, interestingResources sets.Set[corev1alpha1.ResourceName]) bool {
	for resourceName := range resourceQuota.Status.Hard {
		if !interestingResources.Has(resourceName) {
			continue
		}
		if _, found := resourceQuota.Status.Used[resourceName]; !found {
			return false
		}
	}
	return true
}

func prettyPrintResourceNames(a sets.Set[corev1alpha1.ResourceName]) string {
	var (
		values = make([]string, len(a))
		i      int
	)
	for value := range a {
		values[i] = string(value)
		i++
	}
	sort.Strings(values)
	return strings.Join(values, ",")
}
