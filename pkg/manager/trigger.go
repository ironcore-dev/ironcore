package manager

import (
	"context"
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/onmetal/onmetal-api/pkg/utils"
)

// ReconcilationTrigger is used to register controllers
// to be externally triggerable for a dedicated object.
// given by its object id.
type ReconcilationTrigger interface {
	// RegisterControllerFor registers a controller to receice
	// reconcilation requests for a dedicated group kind, if
	// the Trigger method is called for such an object id.
	//
	// ATTENTION: Because a controller is always responsible for ONLY
	//   one object type, but is does not provide a method to
	//   get this type, this is not validated during the registration.
	//   Therefore this method MUST only be called for a controller
	//   if it is handling this type of object.
	RegisterControllerFor(gk schema.GroupKind, c controller.Controller)

	// Trigger triggers all controllers registered for the GroupKind of the given id
	Trigger(id utils.ObjectId)
}

type triggersources map[schema.GroupKind]*triggerSource

// triggerSource is a fake source to catch watch start calls from controllers
// to get access to the queue to feed with reconcilation requests.
// Each source will be responsible for a dedicated group kind and used
// to add watches to any interested controller.
// It will never deliver events to registered event handlers, but directly
// add requests to the registered queues.
// Predicates are not supported, because no object is available during this
// request generation process.
type triggerSource struct {
	lock   sync.RWMutex
	queues map[workqueue.RateLimitingInterface]struct{}
}

func (s *triggerSource) Start(ctx context.Context, handler handler.EventHandler, queue workqueue.RateLimitingInterface, predicate ...predicate.Predicate) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.queues[queue] = struct{}{}
	return nil
}

func (s *triggerSource) Trigger(key client.ObjectKey) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for q := range s.queues {
		q.Add(reconcile.Request{NamespacedName: key})
	}
}

type reconcilationTrigger struct {
	lock    sync.RWMutex
	sources map[schema.GroupKind]*triggerSource
}

func NewReconcilationTrigger() ReconcilationTrigger {
	return &reconcilationTrigger{
		sources: map[schema.GroupKind]*triggerSource{},
	}
}

func (t *reconcilationTrigger) RegisterControllerFor(gk schema.GroupKind, c controller.Controller) {
	t.lock.Lock()
	defer t.lock.Unlock()

	s := t.sources[gk]
	if s == nil {
		s = &triggerSource{
			queues: map[workqueue.RateLimitingInterface]struct{}{},
		}
		t.sources[gk] = s
	}
	c.Watch(s, nil)
}

func (t *reconcilationTrigger) Trigger(id utils.ObjectId) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	s := t.sources[id.GroupKind]
	if s != nil {
		s.Trigger(id.ObjectKey)
	}
}
