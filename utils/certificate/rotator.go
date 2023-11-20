// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package certificate

import (
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	certificatesv1 "k8s.io/api/certificates/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/server/healthz"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(certificatesv1.AddToScheme(scheme))
}

type rotator struct {
	startedMu sync.Mutex
	started   bool

	name              string
	newClient         func(cert *tls.Certificate) (client.WithWatch, error)
	logConstructor    func() logr.Logger
	signerName        string
	template          *x509.CertificateRequest
	getUsages         func(privateKey any) []certificatesv1.KeyUsage
	requestedDuration *time.Duration
	initContext       func(ctx context.Context) (context.Context, context.CancelFunc)

	listenersMu sync.RWMutex
	listeners   sets.Set[*handle]

	forceInitial bool
	certificate  atomic.Pointer[tls.Certificate]

	queue workqueue.RateLimitingInterface
}

type Rotator interface {
	manager.Runnable
	healthz.HealthChecker
	Init(ctx context.Context, force bool) error
	Certificate() *tls.Certificate
	AddListener(listener RotatorListener) RotatorListenerRegistration
	RemoveListener(reg RotatorListenerRegistration)
}

type RotatorOptions struct {
	Name              string
	NewClient         func(cert *tls.Certificate) (client.WithWatch, error)
	LogConstructor    func() logr.Logger
	SignerName        string
	Template          *x509.CertificateRequest
	GetUsages         func(privateKey any) []certificatesv1.KeyUsage
	RequestedDuration *time.Duration
	ForceInitial      bool
	InitCertificate   *tls.Certificate
	InitContext       func(ctx context.Context) (context.Context, context.CancelFunc)
}

func ConstantLogger(logger logr.Logger) func() logr.Logger {
	return func() logr.Logger {
		return logger
	}
}

func setRotatorOptionsDefaults(o *RotatorOptions) {
	if o.LogConstructor == nil {
		o.LogConstructor = func() logr.Logger {
			return ctrl.Log.WithName(o.Name)
		}
	}
	if o.InitContext == nil {
		o.InitContext = func(ctx context.Context) (context.Context, context.CancelFunc) {
			return context.WithTimeout(ctx, 30*time.Second)
		}
	}
}

func NewRotator(opts RotatorOptions) (Rotator, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("must specify Name")
	}
	if opts.NewClient == nil {
		return nil, fmt.Errorf("must specify NewClient")
	}
	if opts.SignerName == "" {
		return nil, fmt.Errorf("must specify SignerName")
	}
	if opts.Template == nil {
		return nil, fmt.Errorf("must specify Template")
	}
	if opts.GetUsages == nil {
		return nil, fmt.Errorf("must specify GetUsages")
	}
	setRotatorOptionsDefaults(&opts)

	r := &rotator{
		newClient:         opts.NewClient,
		name:              opts.Name,
		logConstructor:    opts.LogConstructor,
		initContext:       opts.InitContext,
		signerName:        opts.SignerName,
		template:          opts.Template,
		getUsages:         opts.GetUsages,
		requestedDuration: opts.RequestedDuration,
		forceInitial:      opts.ForceInitial,
		listeners:         sets.New[*handle](),
	}

	if opts.InitCertificate != nil {
		// Create shallow copy to safely mutate the leaf.
		cert := *opts.InitCertificate
		if err := setCertificateLeaf(&cert); err != nil {
			return nil, fmt.Errorf("error setting certificate leaf: %w", err)
		}

		r.certificate.Store(&cert)
	}

	return r, nil
}

func TLSCertificateLeaf(cert *tls.Certificate) (*x509.Certificate, error) {
	if cert.Leaf != nil {
		return cert.Leaf, nil
	}

	if err := setCertificateLeaf(cert); err != nil {
		return nil, err
	}
	return cert.Leaf, nil
}

// parseCSR extracts the CSR from the API object and decodes it.
func parseCSR(pemData []byte) (*x509.CertificateRequest, error) {
	// extract PEM from request object
	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "CERTIFICATE REQUEST" {
		return nil, fmt.Errorf("PEM block type must be CERTIFICATE REQUEST")
	}
	return x509.ParseCertificateRequest(block.Bytes)
}

// ensureCompatible ensures that a CSR object is compatible with an original CSR
func ensureCompatible(new, orig *certificatesv1.CertificateSigningRequest, privateKey interface{}) error {
	newCSR, err := parseCSR(new.Spec.Request)
	if err != nil {
		return fmt.Errorf("unable to parse new csr: %v", err)
	}
	origCSR, err := parseCSR(orig.Spec.Request)
	if err != nil {
		return fmt.Errorf("unable to parse original csr: %v", err)
	}
	if !reflect.DeepEqual(newCSR.Subject, origCSR.Subject) {
		return fmt.Errorf("csr subjects differ: new: %#v, orig: %#v", newCSR.Subject, origCSR.Subject)
	}
	if len(new.Spec.SignerName) > 0 && len(orig.Spec.SignerName) > 0 && new.Spec.SignerName != orig.Spec.SignerName {
		return fmt.Errorf("csr signerNames differ: new %q, orig: %q", new.Spec.SignerName, orig.Spec.SignerName)
	}
	signer, ok := privateKey.(crypto.Signer)
	if !ok {
		return fmt.Errorf("privateKey is not a signer")
	}
	newCSR.PublicKey = signer.Public()
	if err := newCSR.CheckSignature(); err != nil {
		return fmt.Errorf("error validating signature new CSR against old key: %v", err)
	}
	if len(new.Status.Certificate) > 0 {
		certs, err := certutil.ParseCertsPEM(new.Status.Certificate)
		if err != nil {
			return fmt.Errorf("error parsing signed certificate for CSR: %v", err)
		}
		now := time.Now()
		for _, cert := range certs {
			if now.After(cert.NotAfter) {
				return fmt.Errorf("one of the certificates for the CSR has expired: %s", cert.NotAfter)
			}
		}
	}
	return nil
}

func (r *rotator) requestCertificate(ctx context.Context) (*tls.Certificate, error) {
	c, err := r.newClient(r.certificate.Load())
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	return RequestCertificate(ctx, c, r.signerName, r.template, r.getUsages, r.requestedDuration)
}

func setCertificateLeaf(cert *tls.Certificate) error {
	if len(cert.Certificate) == 0 {
		return fmt.Errorf("no certificates in certificate chain")
	}

	certs, err := x509.ParseCertificates(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("error parsing certificate data: %w", err)
	}

	cert.Leaf = certs[0]
	return nil
}

func (r *rotator) nextRotationDeadline(force bool) time.Time {
	if force {
		return time.Now()
	}

	cert := r.certificate.Load()
	if cert == nil {
		return time.Now()
	}

	notAfter := cert.Leaf.NotAfter
	totalDuration := float64(notAfter.Sub(cert.Leaf.NotBefore))
	deadline := cert.Leaf.NotBefore.Add(jitteryDuration(totalDuration))
	return deadline
}

func jitteryDuration(totalDuration float64) time.Duration {
	return wait.Jitter(time.Duration(totalDuration), 0.2) - time.Duration(totalDuration*0.3)
}

func (r *rotator) getListeners() []RotatorListener {
	r.listenersMu.RLock()
	defer r.listenersMu.RUnlock()

	res := make([]RotatorListener, 0, len(r.listeners))
	for hdl := range r.listeners {
		res = append(res, hdl.RotatorListener)
	}
	return res
}

func (r *rotator) rotate(ctx context.Context) error {
	r.logConstructor().V(1).Info("Rotating certificate")
	newCert, err := r.requestCertificate(ctx)
	if err != nil {
		return err
	}

	r.logConstructor().V(1).Info("Certificate rotated, storing updated certificate")
	r.certificate.Store(newCert)

	// Copy active listeners once to release lock as soon as possible.
	listeners := r.getListeners()
	r.logConstructor().V(1).Info("Notifying listeners")
	for _, listener := range listeners {
		listener.Enqueue()
	}
	return nil
}

func (r *rotator) enqueue(key interface{}) {
	// If we want a forced initial rotation, respect it / reset it afterwards.
	force := r.forceInitial
	if force {
		r.forceInitial = false
	}

	deadline := r.nextRotationDeadline(force)
	if duration := time.Until(deadline); duration > 0 {
		r.logConstructor().Info("Enqueuing for planned rotation", "Deadline", deadline, "Duration", duration)
		r.queue.AddAfter(key, duration)
	} else {
		r.logConstructor().Info("Enqueueing for immediate rotation")
		r.queue.Add(key)
	}
}

func (r *rotator) processNextWorkItem(ctx context.Context) bool {
	key, quit := r.queue.Get()
	if quit {
		return false
	}
	defer r.queue.Done(key)

	if err := r.rotate(ctx); err != nil {
		r.logConstructor().Error(err, "Error rotating certificate")
		r.queue.AddRateLimited(key)
		return true
	}

	r.queue.Forget(key)
	r.enqueue(key)
	return true
}

func (r *rotator) Certificate() *tls.Certificate {
	return r.certificate.Load()
}

func (r *rotator) Init(ctx context.Context, force bool) error {
	ctx, cancel := r.initContext(ctx)
	defer cancel()

	if time.Until(r.nextRotationDeadline(force)) <= 0 {
		r.logConstructor().Info("Initial certificate rotation required")

		var lastErr error
		if err := wait.PollUntilContextCancel(ctx, 1*time.Second, true, func(ctx context.Context) (done bool, err error) {
			lastErr = r.rotate(ctx)
			if lastErr != nil {
				r.logConstructor().Error(lastErr, "Error initially rotating certificates")
				return false, nil
			}
			return true, nil
		}); err != nil {
			return fmt.Errorf("error doing initial rotation: %w", lastErr)
		}
	}
	return nil
}

const workItemKey = "key"

func (r *rotator) Start(ctx context.Context) error {
	r.startedMu.Lock()
	if r.started {
		r.startedMu.Unlock()
		return fmt.Errorf("rotator was already started")
	}

	r.started = true
	r.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	r.startedMu.Unlock()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for r.processNextWorkItem(ctx) {
		}
	}()

	go func() {
		defer r.queue.ShutDown()
		<-ctx.Done()
	}()

	// Kick off initial enqueue
	r.enqueue(workItemKey)

	wg.Wait()
	return nil
}

type RotatorListener interface {
	Enqueue()
}

type RotatorListenerFunc func()

func (f RotatorListenerFunc) Enqueue() {
	f()
}

type RotatorListenerRegistration interface{}

type handle struct {
	RotatorListener
}

func (r *rotator) AddListener(listener RotatorListener) RotatorListenerRegistration {
	r.listenersMu.Lock()
	defer r.listenersMu.Unlock()

	h := &handle{listener}
	r.listeners.Insert(h)
	return h
}

func (r *rotator) RemoveListener(reg RotatorListenerRegistration) {
	r.listenersMu.Lock()
	defer r.listenersMu.Unlock()

	h, ok := reg.(*handle)
	if !ok {
		return
	}

	r.listeners.Delete(h)
}

func (r *rotator) Name() string {
	return r.name
}

func (r *rotator) Check(_ *http.Request) error {
	cert := r.Certificate()
	if cert == nil {
		return fmt.Errorf("certificate has not yet been issued")
	}

	leaf, err := TLSCertificateLeaf(cert)
	if err != nil {
		return fmt.Errorf("error getting certificate leaf: %w", err)
	}

	if time.Now().After(leaf.NotAfter) {
		return fmt.Errorf("certificate is expired")
	}
	return nil
}
