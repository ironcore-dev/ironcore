// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package remotecommand

import (
	"io"
	"net/http"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/httpstream/wsstream"
	"k8s.io/client-go/tools/remotecommand"
)

type Streams interface {
	io.Closer
	Stdin() io.ReadCloser
	Stdout() io.WriteCloser
	Stderr() io.WriteCloser
	Resize() remotecommand.TerminalSizeQueue
	WriteStatus(status *apierrors.StatusError) error
}

// streams contains the connection and streams used when
// forwarding an attach or execute session into a container.
type streams struct {
	conn         io.Closer
	stdinStream  io.ReadCloser
	stdoutStream io.WriteCloser
	stderrStream io.WriteCloser
	writeStatus  func(status *apierrors.StatusError) error
	resizeStream io.ReadCloser
	resizeChan   chan remotecommand.TerminalSize
	tty          bool
}

func (s *streams) Close() error {
	return s.conn.Close()
}

func (s *streams) Stdin() io.ReadCloser {
	return s.stdinStream
}

func (s *streams) Stdout() io.WriteCloser {
	return s.stdoutStream
}

func (s *streams) Stderr() io.WriteCloser {
	return s.stderrStream
}

func (s *streams) Resize() remotecommand.TerminalSizeQueue {
	return terminalSizeQueueChannel(s.resizeChan)
}

func (s *streams) WriteStatus(status *apierrors.StatusError) error {
	return s.writeStatus(status)
}

type StreamsOptions struct {
	Stdin  bool
	Stdout bool
	Stderr bool
	TTY    bool

	SupportedProtocols []string

	IdleTimeout     time.Duration
	CreationTimeout time.Duration
}

func NewStreams(
	req *http.Request,
	w http.ResponseWriter,
	opts StreamsOptions,
) (Streams, bool) {
	var s *streams
	var ok bool
	if wsstream.IsWebSocketRequest(req) {
		s, ok = createWebSocketStreams(req, w, opts)
	} else {
		s, ok = createHTTPStreamStreams(req, w, opts)
	}
	if !ok {
		return nil, false
	}

	if s.resizeStream != nil {
		s.resizeChan = make(chan remotecommand.TerminalSize)
		go handleResizeEvents(s.resizeStream, s.resizeChan)
	}

	return s, true
}

type terminalSizeQueueChannel <-chan remotecommand.TerminalSize

func (t terminalSizeQueueChannel) Next() *remotecommand.TerminalSize {
	s, ok := <-t
	if !ok {
		return nil
	}
	return &s
}
