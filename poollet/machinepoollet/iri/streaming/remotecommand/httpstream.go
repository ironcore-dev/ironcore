/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package remotecommand

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	api "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/apimachinery/pkg/util/httpstream/spdy"
	remotecommandconsts "k8s.io/apimachinery/pkg/util/remotecommand"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/remotecommand"

	"k8s.io/klog/v2"
)

// streamAndReply holds both a Stream and a channel that is closed when the stream's reply frame is
// enqueued. Consumers can wait for replySent to be closed prior to proceeding, to ensure that the
// replyFrame is enqueued before the connection's goaway frame is sent (e.g. if a stream was
// received and right after, the connection gets closed).
type streamAndReply struct {
	httpstream.Stream
	replySent <-chan struct{}
}

// waitStreamReply waits until either replySent or stop is closed. If replySent is closed, it sends
// an empty struct to the notify channel.
func waitStreamReply(replySent <-chan struct{}, notify chan<- struct{}, stop <-chan struct{}) {
	select {
	case <-replySent:
		notify <- struct{}{}
	case <-stop:
	}
}

func createHTTPStreamStreams(req *http.Request, w http.ResponseWriter, opts StreamsOptions) (*streams, bool) {
	protocol, err := httpstream.Handshake(req, w, opts.SupportedProtocols)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, false
	}

	streamCh := make(chan streamAndReply)

	upgrader := spdy.NewResponseUpgrader()
	conn := upgrader.UpgradeResponse(w, req, func(stream httpstream.Stream, replySent <-chan struct{}) error {
		streamCh <- streamAndReply{Stream: stream, replySent: replySent}
		return nil
	})
	// from this point on, we can no longer call methods on response
	if conn == nil {
		// The upgrader is responsible for notifying the client of any errors that
		// occurred during upgrading. All we can do is return here at this point
		// if we weren't successful in upgrading.
		return nil, false
	}

	conn.SetIdleTimeout(opts.IdleTimeout)

	var handler protocolHandler
	switch protocol {
	case remotecommandconsts.StreamProtocolV4Name:
		handler = &v4ProtocolHandler{}
	case remotecommandconsts.StreamProtocolV3Name:
		handler = &v3ProtocolHandler{}
	case remotecommandconsts.StreamProtocolV2Name:
		handler = &v2ProtocolHandler{}
	case "":
		klog.V(4).InfoS("Client did not request protocol negotiation. Falling back", "protocol", remotecommandconsts.StreamProtocolV1Name)
		fallthrough
	case remotecommandconsts.StreamProtocolV1Name:
		handler = &v1ProtocolHandler{}
	}

	// count the streams client asked for, starting with 1
	expectedStreams := 1
	if opts.Stdin {
		expectedStreams++
	}
	if opts.Stdout {
		expectedStreams++
	}
	if opts.Stderr {
		expectedStreams++
	}
	if opts.TTY && handler.supportsTerminalResizing() {
		expectedStreams++
	}

	expired := time.NewTimer(opts.CreationTimeout)
	defer expired.Stop()

	ctx, err := handler.waitForStreams(streamCh, expectedStreams, expired.C)
	if err != nil {
		runtime.HandleError(err)
		return nil, false
	}

	ctx.conn = conn
	ctx.tty = opts.TTY

	return ctx, true
}

type protocolHandler interface {
	// waitForStreams waits for the expected streams or a timeout, returning a
	// remoteCommandContext if all the streams were received, or an error if not.
	waitForStreams(streams <-chan streamAndReply, expectedStreams int, expired <-chan time.Time) (*streams, error)
	// supportsTerminalResizing returns true if the protocol handler supports terminal resizing
	supportsTerminalResizing() bool
}

// v4ProtocolHandler implements the V4 protocol version for streaming command execution. It only differs
// in from v3 in the error stream format using a json-marshaled metav1.Status which carries
// the process' exit code.
type v4ProtocolHandler struct{}

func (*v4ProtocolHandler) waitForStreams(streamCh <-chan streamAndReply, expectedStreams int, expired <-chan time.Time) (*streams, error) {
	ctx := &streams{}
	receivedStreams := 0
	replyChan := make(chan struct{})
	stop := make(chan struct{})
	defer close(stop)
WaitForStreams:
	for {
		select {
		case stream := <-streamCh:
			streamType := stream.Headers().Get(api.StreamType)
			switch streamType {
			case api.StreamTypeError:
				ctx.writeStatus = v4WriteStatusFunc(stream) // write json errors
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStdin:
				ctx.stdinStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStdout:
				ctx.stdoutStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStderr:
				ctx.stderrStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeResize:
				ctx.resizeStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			default:
				runtime.HandleError(fmt.Errorf("unexpected stream type: %q", streamType))
			}
		case <-replyChan:
			receivedStreams++
			if receivedStreams == expectedStreams {
				break WaitForStreams
			}
		case <-expired:
			// TODO find a way to return the error to the user. Maybe use a separate
			// stream to report errors?
			return nil, errors.New("timed out waiting for client to create streams")
		}
	}

	return ctx, nil
}

// supportsTerminalResizing returns true because v4ProtocolHandler supports it
func (*v4ProtocolHandler) supportsTerminalResizing() bool { return true }

// v3ProtocolHandler implements the V3 protocol version for streaming command execution.
type v3ProtocolHandler struct{}

func (*v3ProtocolHandler) waitForStreams(streamCh <-chan streamAndReply, expectedStreams int, expired <-chan time.Time) (*streams, error) {
	ctx := &streams{}
	receivedStreams := 0
	replyChan := make(chan struct{})
	stop := make(chan struct{})
	defer close(stop)
WaitForStreams:
	for {
		select {
		case stream := <-streamCh:
			streamType := stream.Headers().Get(api.StreamType)
			switch streamType {
			case api.StreamTypeError:
				ctx.writeStatus = v1WriteStatusFunc(stream)
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStdin:
				ctx.stdinStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStdout:
				ctx.stdoutStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStderr:
				ctx.stderrStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeResize:
				ctx.resizeStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			default:
				runtime.HandleError(fmt.Errorf("unexpected stream type: %q", streamType))
			}
		case <-replyChan:
			receivedStreams++
			if receivedStreams == expectedStreams {
				break WaitForStreams
			}
		case <-expired:
			// TODO find a way to return the error to the user. Maybe use a separate
			// stream to report errors?
			return nil, errors.New("timed out waiting for client to create streams")
		}
	}

	return ctx, nil
}

// supportsTerminalResizing returns true because v3ProtocolHandler supports it
func (*v3ProtocolHandler) supportsTerminalResizing() bool { return true }

// v2ProtocolHandler implements the V2 protocol version for streaming command execution.
type v2ProtocolHandler struct{}

func (*v2ProtocolHandler) waitForStreams(streamCh <-chan streamAndReply, expectedStreams int, expired <-chan time.Time) (*streams, error) {
	ctx := &streams{}
	receivedStreams := 0
	replyChan := make(chan struct{})
	stop := make(chan struct{})
	defer close(stop)
WaitForStreams:
	for {
		select {
		case stream := <-streamCh:
			streamType := stream.Headers().Get(api.StreamType)
			switch streamType {
			case api.StreamTypeError:
				ctx.writeStatus = v1WriteStatusFunc(stream)
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStdin:
				ctx.stdinStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStdout:
				ctx.stdoutStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStderr:
				ctx.stderrStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			default:
				runtime.HandleError(fmt.Errorf("unexpected stream type: %q", streamType))
			}
		case <-replyChan:
			receivedStreams++
			if receivedStreams == expectedStreams {
				break WaitForStreams
			}
		case <-expired:
			// TODO find a way to return the error to the user. Maybe use a separate
			// stream to report errors?
			return nil, errors.New("timed out waiting for client to create streams")
		}
	}

	return ctx, nil
}

// supportsTerminalResizing returns false because v2ProtocolHandler doesn't support it.
func (*v2ProtocolHandler) supportsTerminalResizing() bool { return false }

// v1ProtocolHandler implements the V1 protocol version for streaming command execution.
type v1ProtocolHandler struct{}

func (*v1ProtocolHandler) waitForStreams(streamCh <-chan streamAndReply, expectedStreams int, expired <-chan time.Time) (*streams, error) {
	ctx := &streams{}
	receivedStreams := 0
	replyChan := make(chan struct{})
	stop := make(chan struct{})
	defer close(stop)
WaitForStreams:
	for {
		select {
		case stream := <-streamCh:
			streamType := stream.Headers().Get(api.StreamType)
			switch streamType {
			case api.StreamTypeError:
				ctx.writeStatus = v1WriteStatusFunc(stream)

				// This defer statement shouldn't be here, but due to previous refactiring, it ended up in
				// here. This is what 1.0.x kubelets do, so we're retaining that behavior. This is fixed in
				// the v2ProtocolHandler.
				defer func() { _ = stream.Reset() }()

				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStdin:
				ctx.stdinStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStdout:
				ctx.stdoutStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			case api.StreamTypeStderr:
				ctx.stderrStream = stream
				go waitStreamReply(stream.replySent, replyChan, stop)
			default:
				runtime.HandleError(fmt.Errorf("unexpected stream type: %q", streamType))
			}
		case <-replyChan:
			receivedStreams++
			if receivedStreams == expectedStreams {
				break WaitForStreams
			}
		case <-expired:
			// TODO find a way to return the error to the user. Maybe use a separate
			// stream to report errors?
			return nil, errors.New("timed out waiting for client to create streams")
		}
	}

	if ctx.stdinStream != nil {
		_ = ctx.stdinStream.Close()
	}

	return ctx, nil
}

// supportsTerminalResizing returns false because v1ProtocolHandler doesn't support it.
func (*v1ProtocolHandler) supportsTerminalResizing() bool { return false }

func handleResizeEvents(stream io.Reader, channel chan<- remotecommand.TerminalSize) {
	defer runtime.HandleCrash()
	defer close(channel)

	decoder := json.NewDecoder(stream)
	for {
		size := remotecommand.TerminalSize{}
		if err := decoder.Decode(&size); err != nil {
			break
		}
		channel <- size
	}
}

func v1WriteStatusFunc(stream io.Writer) func(status *apierrors.StatusError) error {
	return func(status *apierrors.StatusError) error {
		if status.Status().Status == metav1.StatusSuccess {
			return nil // send error messages
		}
		_, err := stream.Write([]byte(status.Error()))
		return err
	}
}

// v4WriteStatusFunc returns a WriteStatusFunc that marshals a given api Status
// as json in the error channel.
func v4WriteStatusFunc(stream io.Writer) func(status *apierrors.StatusError) error {
	return func(status *apierrors.StatusError) error {
		bs, err := json.Marshal(status.Status())
		if err != nil {
			return err
		}
		_, err = stream.Write(bs)
		return err
	}
}
