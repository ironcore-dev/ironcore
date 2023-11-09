// Copyright 2023 OnMetal authors
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

package context_test

import (
	"context"

	. "github.com/onmetal/onmetal-api/utils/context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Context", func() {
	Describe("FromStopChannel", func() {
		It("should create a context from the given stop channel", func() {
			stopChan := make(chan struct{})
			ctx := FromStopChannel(stopChan)

			Expect(ctx.Done()).NotTo(BeClosed())
			Expect(ctx.Err()).To(Succeed())

			close(stopChan)

			Expect(ctx.Done()).To(BeClosed())
			Expect(ctx.Err()).To(Equal(context.Canceled))
		})

		It("should panic if a value is sent through the channel when determining whether it's closed", func() {
			stopChan := make(chan struct{}, 1)
			ctx := FromStopChannel(stopChan)

			stopChan <- struct{}{}
			Expect(func() { _ = ctx.Err() }).To(Panic())
		})
	})
})
