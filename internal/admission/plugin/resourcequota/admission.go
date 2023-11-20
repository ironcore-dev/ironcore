// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package resourcequota

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ironcore-dev/ironcore/client-go/informers"
	"github.com/ironcore-dev/ironcore/client-go/ironcore"
	corev1alpha1listers "github.com/ironcore-dev/ironcore/client-go/listers/core/v1alpha1"
	utilcontext "github.com/ironcore-dev/ironcore/utils/context"
	"github.com/ironcore-dev/ironcore/utils/quota"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/admission"
)

const PluginName = "ResourceQuota"

func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		return NewResourceQuota(), nil
	})
}

type ResourceQuota struct {
	initOnce sync.Once
	initErr  error

	ctx context.Context

	client   ironcore.Interface
	lister   corev1alpha1listers.ResourceQuotaLister
	registry quota.Registry

	*admission.Handler

	evaluator Evaluator
}

func (r *ResourceQuota) SetDrainedNotification(stopCh <-chan struct{}) {
	r.ctx = utilcontext.FromStopChannel(stopCh)
}

func (r *ResourceQuota) SetQuotaRegistry(registry quota.Registry) {
	r.registry = registry
}

func (r *ResourceQuota) SetExternalIronCoreClientSet(client ironcore.Interface) {
	r.client = client
}

func (r *ResourceQuota) SetExternalIronCoreInformerFactory(f informers.SharedInformerFactory) {
	r.lister = f.Core().V1alpha1().ResourceQuotas().Lister()
}

func (r *ResourceQuota) init() error {
	r.initOnce.Do(func() {
		r.initErr = func() error {
			if r.ctx == nil {
				return fmt.Errorf("missing context")
			}
			if r.client == nil {
				return fmt.Errorf("missing client")
			}
			if r.lister == nil {
				return fmt.Errorf("missing lister")
			}
			if r.registry == nil {
				return fmt.Errorf("missing registry")
			}

			quotaAcc, err := NewQuotaAccessor(r.client, r.lister)
			if err != nil {
				return fmt.Errorf("error creating new quota accessor: %w", err)
			}

			r.evaluator = &startOnceEvaluatorController{
				ctx:  r.ctx,
				ctrl: NewEvaluatorController(quotaAcc, r.registry),
			}
			return nil
		}()
	})
	return r.initErr
}

type startOnceEvaluatorController struct {
	once sync.Once
	ctx  context.Context
	ctrl *EvaluatorController
}

func (c *startOnceEvaluatorController) Evaluate(ctx context.Context, a admission.Attributes) error {
	c.once.Do(func() {
		go func() {
			_ = c.ctrl.Start(c.ctx)
		}()
	})
	return c.ctrl.Evaluate(ctx, a)
}

func (r *ResourceQuota) ValidateInitialization() error {
	return r.init()
}

func NewResourceQuota() *ResourceQuota {
	return &ResourceQuota{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}
}

func shouldHandle(a admission.Attributes) bool {
	if a.GetSubresource() != "" {
		return false
	}

	if a.GetNamespace() == "" || isNamespaceCreation(a) {
		return false
	}

	return true
}

func isNamespaceCreation(a admission.Attributes) bool {
	return a.GetOperation() == admission.Create &&
		a.GetKind().GroupKind() == corev1.SchemeGroupVersion.WithKind("Namespace").GroupKind()
}

func (r *ResourceQuota) Validate(ctx context.Context, a admission.Attributes, o admission.ObjectInterfaces) error {
	if !shouldHandle(a) {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := r.evaluator.Evaluate(ctx, a); err != nil {
		if errors.Is(err, context.Canceled) {
			return apierrors.NewInternalError(fmt.Errorf("resource quota evaluation timed out"))
		}
		return err
	}
	return nil
}
