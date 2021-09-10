// Package testutils provides utilities for writing (ginkgo / gomega - based) tests.
package testutils

import (
	"fmt"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

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

type beMetaV1TemporallyOverlay struct {
	matchers.BeTemporallyMatcher
}

func (o *beMetaV1TemporallyOverlay) FailureMessage(actual interface{}) (message string) {
	return o.BeTemporallyMatcher.FailureMessage(actual.(metav1.Time).Time)
}

func (o *beMetaV1TemporallyOverlay) NegatedFailureMessage(actual interface{}) (message string) {
	return o.BeTemporallyMatcher.NegatedFailureMessage(actual.(metav1.Time).Time)
}

func (o *beMetaV1TemporallyOverlay) Match(actual interface{}) (success bool, err error) {
	t, ok := actual.(metav1.Time)
	if !ok {
		return false, fmt.Errorf("wxpected a metav1.Time. Got:\n%s", format.Object(actual, 1))
	}

	return o.BeTemporallyMatcher.Match(t.Time)
}

func (o *beMetaV1TemporallyOverlay) GomegaString() string {
	if len(o.Threshold) > 0 {
		return fmt.Sprintf("%s %s threshold %s", o.Comparator, o.CompareTo, o.Threshold[0])
	}
	return fmt.Sprintf("%s %s", o.Comparator, o.CompareTo)
}

func BeMetaV1Temporally(comparator string, compareTo metav1.Time, threshold ...time.Duration) types.GomegaMatcher {
	return &beMetaV1TemporallyOverlay{
		BeTemporallyMatcher: matchers.BeTemporallyMatcher{
			Comparator: comparator,
			CompareTo:  compareTo.Time,
			Threshold:  threshold,
		},
	}
}
