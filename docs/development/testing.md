# Testing

This project is using [Ginkgo](https://onsi.github.io/ginkgo/) as it's primary testing framework in conjunction with
[Gomega](https://onsi.github.io/gomega/) matcher/assertion library.

## Unit Tests

Each package should consist of its own `suite_test` setup and the corresponding test cases for each component.

Example of test suite setup is below:

```go
package mypackage

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func Test(t *testing.T) {
	RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "MyComponent")
}
```

The testing code should meet the requirements of be common [Ginkgo](https://onsi.github.io/ginkgo/) format

```go
package mypackage

import
...

var _ = Describe("MyComponent", func() {

	BeforeEach(func() {
		// Code to run before each Context
	})

	Context("When doing x", func() {
		It("Should result in y", func() {
			By("Creating something in x")
			Expect(x.DoSomething()).To(Equal("expected result"))
		})
	})
})
```

!!! note More information on how to structure your tests can be found
here: [Ginkgo documentation](https://onsi.github.io/ginkgo/#structuring-your-specs). Assertion examples can be found
here: [Gomega documentation](https://onsi.github.io/gomega/#making-assertions).

## Controller Tests

Setup a local Kubernetes control plane in order to write controller tests. Use `envtest` as a part of
the [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) project.

Example of `suite_test.go` inside a controller package is below:

```go
package my_controller_package

import
...

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
	// Here is the actual envtest setup. Make sure that the path
	// to your generated CRDs is correct, as it will be injected
	// directly into the API server once the envtest environment comes up.
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}
	...
	// Define scheme
	err = api.AddToScheme(scheme.Scheme)
	...
	// Create a corresponding Kubernetes client.
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	...
	k8sManager, err := manager.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
		// On MacOS it might happen, that the firewall warnings will
		// popup if you open a port on your machine. It typically
		// happens due to the metrics endpoint of the controller-manager.
		// To prevent it, disable it in the local setup
		// and set the Host parameter to localhost.
		Host:               "127.0.0.1",
		MetricsBindAddress: "0",
	})
	...
	// Register our reconciler with the manager. In case if you want to test
	// multiple reconcilers at once you have to register them one by
	// one in the same fashion as is shown below.
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

The Ginkgo style tests can be now written in the same manner as described in the [Unit Test](#unit-tests) section. The
only difference now is, that you have a working controller manager in the background which is reacting on changes in the
Kubernetes API which you can access via the `k8sClient` to create or modify your resources.

More information on the envtest setup you can find in the CRD testing section
here: [Kubebuilder](https://book.kubebuilder.io/reference/envtest.html)

## Running Tests

Test run can be executed via:

```shell
make test
```

## Goland Integration

Running static Ginkgo/Gomega tests in Golang should work out of the box. However, in order to make the controller test
run from within your IDE you need to expose the following environment variable inside your 'Test Run Configuration'

```shell
KUBEBUILDER_ASSETS=/PATH_TO_MY_WORKSPACE/ironcore-dev/ironcore/testbin/bin
```

This is typically the location of the Kubernetes control plane binaries on your machine.
