// Package testutils provides utilities for writing (ginkgo / gomega - based) tests.
package testutils

import "github.com/onsi/ginkgo"

// NowAndBeforeEach runs the given body immediately and also before each test function.
func NowAndBeforeEach(body func(), timeout ...float64) {
	body()
	ginkgo.BeforeEach(body, timeout...)
}

// NowAndAfterEach runs the given body immediately and also after each test function.
func NowAndAfterEach(body func(), timeout ...float64) {
	body()
	ginkgo.AfterEach(body, timeout...)
}
