// Copyright 2022 OnMetal authors
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

package fake

import (
	"bytes"
	"io"
	"sync"

	"github.com/onmetal/onmetal-api/poollet/machinepoollet/terminal/session"
	"k8s.io/client-go/tools/remotecommand"
)

// Terminal is a fake implementation of a terminal.Terminal.
//
// Usually, mocks are a better choice, however, due to the heavily asynchronous environment, a fake
// makes testing way more convenient.
//
// It mimics the real-life behavior as close as possible, exhibiting the following traits:
// * If Run.in closes, the terminal shuts down.
// * If a write-error to out / err occur, the terminal shuts down.
// * Reads from in only happen if in is non-nil.
// * Resizes are only read if resizes is non-nil.
type Terminal struct {
	mu sync.RWMutex

	done    chan struct{}
	doneErr error

	connected chan struct{}
	in        bytes.Buffer
	out       io.Writer
	err       io.Writer
	resizes   []remotecommand.TerminalSize
}

func (f *Terminal) Resizes() []remotecommand.TerminalSize {
	f.mu.RLock()
	defer f.mu.RUnlock()
	res := make([]remotecommand.TerminalSize, len(f.resizes))
	copy(res, f.resizes)
	return res
}

func (f *Terminal) InBytes() []byte {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.in.Bytes()
}

func (f *Terminal) WriteOut(data []byte) error {
	if _, err := f.out.Write(data); err != nil {
		f.exit(err)
		return err
	}
	return nil
}

func (f *Terminal) WriteErr(data []byte) error {
	if _, err := f.err.Write(data); err != nil {
		f.exit(err)
		return err
	}
	return nil
}

func (f *Terminal) exit(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	select {
	case <-f.done:
		return
	default:
		f.doneErr = err
		close(f.done)
	}
}

func (f *Terminal) Run(in io.Reader, out, err io.WriteCloser, resize <-chan remotecommand.TerminalSize) error {
	if in != nil {
		go func() {
			f.exit(session.CopyDetachable(f.done, &f.in, in))
		}()
	}
	if resize != nil {
		go func() {
			for {
				select {
				case <-f.done:
					return
				case r := <-resize:
					f.mu.Lock()
					f.resizes = append(f.resizes, r)
					f.mu.Unlock()
				}
			}
		}()
	}
	f.out = out
	f.err = err
	close(f.connected)
	<-f.done
	return f.doneErr
}

func (f *Terminal) Close() {
	f.exit(nil)
}

func (f *Terminal) Closed() bool {
	select {
	case <-f.done:
		return true
	default:
		return false
	}
}

func (f *Terminal) Connected() bool {
	select {
	case <-f.connected:
		return true
	default:
		return false
	}
}

func NewFakeTerm() *Terminal {
	return &Terminal{
		done:      make(chan struct{}),
		connected: make(chan struct{}),
	}
}

type ThreadSafeBuffer struct {
	mu  sync.RWMutex
	buf bytes.Buffer
}

func (b *ThreadSafeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *ThreadSafeBuffer) Bytes() []byte {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.buf.Bytes()
}

func (b *ThreadSafeBuffer) Close() error {
	return nil
}
