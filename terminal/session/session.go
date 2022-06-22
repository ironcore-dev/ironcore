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

// Package session allows sharing a single terminal.Terminal with multiple connections.
package session

import (
	"container/list"
	"errors"
	"io"
	"sync"

	"github.com/moby/term"
	"github.com/onmetal/onmetal-api/terminal"
	"k8s.io/client-go/tools/remotecommand"
)

// ErrStoppedAccepting is an error that gets returned upon calling Session.Run when the session is already
// shutting down / is shut down.
var ErrStoppedAccepting = errors.New("stopped accepting any new connections")

// Session allows multiplexing a terminal.Terminal to multiple connections.
//
// Initially, a Session starts without any connection. Once the Session received any connection via Run, if
// all connections exited, the session exits as well.
// If the underlying terminal exits, the Session stops all connections and exits.
type Session struct {
	mu sync.RWMutex

	exited chan struct{}

	term terminal.Terminal
	tty  bool

	done     chan struct{}
	doneOnce sync.Once

	conns chan conn
}

// Start starts a new Session with the given terminal.
//
// tty determines whether stderr can be active.
func Start(term terminal.Terminal, tty bool) *Session {
	s := &Session{
		done:   make(chan struct{}),
		exited: make(chan struct{}),
		term:   term,
		tty:    tty,
	}
	s.start()
	return s
}

// Exited returns a channel that gets closed as soon as the session exited completely, meaning
// connections don't get new reads / writes and the terminal.Terminal routine closed as well.
func (s *Session) Exited() <-chan struct{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.exited
}

type conn struct {
	In     io.Reader
	Out    io.WriteCloser
	Err    io.WriteCloser
	Resize <-chan remotecommand.TerminalSize

	Res chan error
}

// Run runs the given parameters as a new connection.
func (s *Session) Run(in io.Reader, out, err io.WriteCloser, resize <-chan remotecommand.TerminalSize) error {
	c := conn{
		In:     in,
		Out:    out,
		Err:    err,
		Resize: resize,
		Res:    make(chan error, 1),
	}

	select {
	case <-s.done:
		return ErrStoppedAccepting
	case s.conns <- c:
		return <-c.Res
	}
}

func (s *Session) start() {
	s.conns = make(chan conn)

	go func() {
		defer func() {
			s.mu.Lock()
			defer s.mu.Unlock()
			close(s.exited)
		}()
		s.loop()
	}()
}

// CopyDetachable copies src to dst while respecting both the done channel and an escape sequence (ctrl-p ctrl-q)
// to cancel streaming.
func CopyDetachable(done <-chan struct{}, dst io.Writer, src io.Reader) error {
	// Default detach keys: ctrl-p ctrl-q
	src = term.NewEscapeProxy(src, []byte{16, 17})
	p := make([]byte, 1024*32)
	for {
		select {
		case <-done:
			return nil
		default:
			n, err := src.Read(p)
			if err != nil {
				switch {
				case errors.Is(err, io.EOF):
					return nil
				case errors.As(err, new(term.EscapeError)):
					return nil
				default:
					return err
				}
			}

			select {
			case <-done:
				return nil
			default:
				w, err := dst.Write(p[:n])
				if err != nil {
					return err
				}

				if w < n {
					return io.ErrShortWrite
				}
			}
		}
	}
}

func forwardResize(done <-chan struct{}, dst chan<- remotecommand.TerminalSize, src <-chan remotecommand.TerminalSize) {
	for {
		select {
		case <-done:
			return
		case r, ok := <-src:
			if !ok {
				return
			}

			select {
			case <-done:
				return
			case dst <- r:
			}
		}
	}
}

func runConn(done <-chan struct{}, getTermErr func() error, in io.Reader, termIn io.Writer, outErr chan error, errErr chan error, resize <-chan remotecommand.TerminalSize, termResize chan<- remotecommand.TerminalSize) error {
	var (
		inErr chan error
	)

	if in != nil {
		inErr = make(chan error, 1)
		go func() {
			inErr <- CopyDetachable(done, termIn, in)
		}()
	}
	if resize != nil {
		go func() {
			forwardResize(done, termResize, resize)
		}()
	}

	select {
	case <-done:
		return getTermErr()
	case err := <-inErr:
		return err
	case err := <-outErr:
		return err
	case err := <-errErr:
		return err
	}
}

type connsWriters struct {
	mu   sync.RWMutex
	list *list.List
}

func (c *connsWriters) Add(w connWriter) *list.Element {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.list.PushFront(w)
}

func (c *connsWriters) Remove(elem *list.Element) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.list.Remove(elem)
}

func (c *connsWriters) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.list.Len()
}

func (c *connsWriters) eachItem(f func(w *connWriter) bool) {
	c.mu.RLock()
	ws := copyList(c.list)
	c.mu.RUnlock()

	for _, w := range ws {
		if !f(&w) {
			break
		}
	}
}

func (c *connsWriters) Out() io.WriteCloser {
	return &connsWritersOut{c}
}

func (c *connsWriters) Err() io.WriteCloser {
	return &connsWritersErr{c}
}

func newConnsWriters() *connsWriters {
	return &connsWriters{list: list.New()}
}

type connsWritersOut struct {
	*connsWriters
}

func copyList(list *list.List) []connWriter {
	res := make([]connWriter, 0, list.Len())
	for elem := list.Front(); elem != nil; elem = elem.Next() {
		res = append(res, elem.Value.(connWriter))
	}
	return res
}

func (c *connsWritersOut) Write(p []byte) (int, error) {
	c.eachItem(func(w *connWriter) bool {
		if _, err := w.out.Write(p); err != nil {
			select {
			case w.outErr <- err:
			default:
			}
		}
		return true
	})
	return len(p), nil
}

func (c *connsWritersOut) Close() error {
	c.eachItem(func(w *connWriter) bool {
		if err := w.out.Close(); err != nil {
			select {
			case w.outErr <- err:
			default:
			}
		}
		return true
	})
	return nil
}

type connsWritersErr struct {
	*connsWriters
}

func (c *connsWritersErr) Write(p []byte) (int, error) {
	c.eachItem(func(w *connWriter) bool {
		if _, err := w.err.Write(p); err != nil {
			select {
			case w.errErr <- err:
			default:
			}
		}
		return true
	})
	return len(p), nil
}

func (c *connsWritersErr) Close() error {
	c.eachItem(func(w *connWriter) bool {
		if err := w.err.Close(); err != nil {
			select {
			case w.errErr <- err:
			default:
			}
		}
		return true
	})
	return nil
}

type connWriter struct {
	out    io.WriteCloser
	outErr chan error

	err    io.WriteCloser
	errErr chan error
}

type onceError struct {
	sync.Mutex // guards following
	err        error
}

func (a *onceError) Store(err error) {
	a.Lock()
	defer a.Unlock()
	if a.err != nil {
		return
	}
	a.err = err
}

func (a *onceError) Load() error {
	a.Lock()
	defer a.Unlock()
	return a.err
}

func (s *Session) Stop() {
	s.doneOnce.Do(func() { close(s.done) })
}

func (s *Session) loop() {
	var (
		termDone        = make(chan struct{})
		termErr         onceError
		writeIn, termIn = io.Pipe()
		resize          = make(chan remotecommand.TerminalSize)
		writers         = newConnsWriters()
		removeConn      = make(chan *list.Element)
	)
	go func() {
		defer close(termDone)
		defer s.Stop()
		var err io.WriteCloser
		if !s.tty {
			err = writers.Err()
		}
		res := s.term.Run(writeIn, writers.Out(), err, resize)
		termErr.Store(res)
	}()

Loop:
	for {
		select {
		case <-s.done:
			_ = writeIn.Close()
			break Loop
		case c := <-s.conns:
			var (
				outErr, errErr chan error
			)
			if c.Out != nil {
				outErr = make(chan error, 1)
			}
			if !s.tty && c.Err != nil {
				errErr = make(chan error, 1)
			}

			elem := writers.Add(connWriter{out: c.Out, outErr: outErr, err: c.Err, errErr: errErr})
			go func() {
				defer func() { removeConn <- elem }()
				err := runConn(termDone, termErr.Load, c.In, termIn, outErr, errErr, c.Resize, resize)
				c.Res <- err
			}()
		case elem := <-removeConn:
			writers.Remove(elem)
		}
	}

	for writers.Len() > 0 {
		elem := <-removeConn
		writers.Remove(elem)
	}

	<-termDone
}
