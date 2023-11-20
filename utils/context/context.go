// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

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
