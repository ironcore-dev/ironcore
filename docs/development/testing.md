# Testing

This project is using [Ginkgo](https://onsi.github.io/ginkgo/) as it's primary testing framework in conjunctions with 
the [Gomega](https://onsi.github.io/gomega/) matcher/assertion library.

## Unit Tests

Each package should come with its own `suite_test` setup and the corresponding test cases for each component.

The test suite setup typically looks like the following 

```go
package mypackage

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func Test(t *testing.T) {
	RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "MyComponent")
}
```

The testing code is self is should be in the common [Ginkgo](https://onsi.github.io/ginkgo/) format

```go
package mypackage

import ...

var _ = Describe("MyComponent", func() {

	BeforeEach(func() {
		// code which should run before each Context
	})

	Context("When doing x", func() {
		It("Should result in y", func() {
			By("Creating something in x")
			Expect(x.DoSomething()).To(Equal("expected result"))
		})
	})
})
```

More information on how to structure your tests can be found in the [Ginkgo documentation](https://onsi.github.io/ginkgo/#structuring-your-specs).
Assertion examples can be found in the the [Gomega documentation](https://onsi.github.io/gomega/#making-assertions).

## Controller Tests

Writing controller tests is a little more involved as we first need to setup a local Kubernetes control plane.
Here we are using `envtest` which is part of the [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) project.

The `suite_test.go` inside a controller package looks like the following

```go
package my_controller_package

import ...

// Those global vars are needed later.
var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	...
	// Here the actual envtest setup is happening. Make sure that the 
	// path to your generated CRDs is correct as we will inject them 
	// directly into the API server once the envtest environment comes up.
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}
	...
	// Now we need to define our scheme ...
	err = api.AddToScheme(scheme.Scheme)
	...
	// ... and can create a corresponding Kubernetes client.
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	...
	k8sManager, err := manager.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		// On MacOS it can happen, that the firewall will popup you 
		// a waring if you open a port on your machine. This is 
		// typically due to the metrics endpoint of the controller-manager.
		// To be on the safe side you can disable it in the local setup 
		// and also set the Host parameter to localhost.
		Host:               "127.0.0.1",
		MetricsBindAddress: "0",
	})
    ...
	// Register our reconciler with the manager. In case you want to test 
	// multiple reconciler at once you have to register them one after 
	// another in the same fashion as shown below.
	err = (&MyObjectReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
		Log:    ctrl.Log.WithName("controllers").WithName("MyObject"),
	}).SetupWithManager(k8sManager)
	...

	// Start the manager
	go func() {
		err = k8sManager.Manager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
```

The Ginkgo style tests can be now written in the same manner as described in the [Unit Test](#unit-tests) section. The only
difference now is, that you have a working controller manager in the background which is reacting on changes in the
Kubernetes API which you can access via the `k8sClient` to create or modify your resources.

For more information on the `envtest` setup please consult the [Kubebuilder](https://book.kubebuilder.io/reference/envtest.html) 
documentation on CRD testing.

## Running Tests

A test run can be executed via

```shell
make test
```

## Goland Integration

Running static Ginkgo/Gomega tests in Golang should work out of the box. However, in order to make the controller 
test run from within your IDE you need to expose the following environment variable inside your 'Test Run Configuration'

```shell
KUBEBUILDER_ASSETS=/PATH_TO_MY_WORKSPACE/onmetal/onmetal-api/testbin/bin
```

This is typically the location of the Kubernetes control plane binaries on your machine.