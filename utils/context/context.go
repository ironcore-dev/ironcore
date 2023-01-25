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

package context

import (
	"context"
	"time"
)

type stopChannelContext <-chan struct{}

func (c stopChannelContext) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (c stopChannelContext) Err() error {
	select {
	case _, ok := <-c:
		if ok {
			panic("stop channel sent item instead of being close-only")
		}
		return context.Canceled
	default:
		return nil
	}
}

func (c stopChannelContext) Value(key any) any {
	return nil
}

func (c stopChannelContext) Done() <-chan struct{} {
	return c
}

// FromStopChannel creates a new context.Context from the given stopChan.
// If a value is sent from the channel, the context.Context.Err() method panics.
// The channel should only be closed to signal completion.
func FromStopChannel(stopChan <-chan struct{}) context.Context {
	return stopChannelContext(stopChan)
}
