package condition_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCondition(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Condition Suite")
}
