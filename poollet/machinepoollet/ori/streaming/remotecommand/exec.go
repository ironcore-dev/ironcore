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
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	remotecommandconsts "k8s.io/apimachinery/pkg/util/remotecommand"
	"k8s.io/client-go/tools/remotecommand"
	utilexec "k8s.io/utils/exec"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Exec is an implementation of exec.
type Exec interface {
	// Exec runs exec.
	Exec(ctx context.Context, in io.Reader, out io.WriteCloser, resize remotecommand.TerminalSizeQueue) error
}

type ExecFunc func(ctx context.Context, in io.Reader, out io.WriteCloser, resize remotecommand.TerminalSizeQueue) error

func (f ExecFunc) Exec(ctx context.Context, in io.Reader, out io.WriteCloser, resize remotecommand.TerminalSizeQueue) error {
	return f(ctx, in, out, resize)
}

type ExecHandlerOptions struct {
	SupportedStreamProtocols []string

	StreamIdleTimeout     time.Duration
	StreamCreationTimeout time.Duration
}

type ExecHandler struct {
	exec Exec

	supportedStreamProtocols []string
	streamIdleTimeout        time.Duration
	streamCreationTimeout    time.Duration
}

func setExecHandlerOptionsDefaults(o *ExecHandlerOptions) {
	if o.SupportedStreamProtocols == nil {
		o.SupportedStreamProtocols = remotecommandconsts.SupportedStreamingProtocols
	}
	if o.StreamIdleTimeout == 0 {
		o.StreamCreationTimeout = 4 * time.Hour
	}
	if o.StreamCreationTimeout == 0 {
		o.StreamCreationTimeout = remotecommandconsts.DefaultStreamCreationTimeout
	}
}

func NewExecHandler(exec Exec, opts ExecHandlerOptions) (*ExecHandler, error) {
	if exec == nil {
		return nil, fmt.Errorf("must specify exec")
	}

	setExecHandlerOptionsDefaults(&opts)

	return &ExecHandler{
		exec:                     exec,
		supportedStreamProtocols: opts.SupportedStreamProtocols,
		streamIdleTimeout:        opts.StreamIdleTimeout,
		streamCreationTimeout:    opts.StreamCreationTimeout,
	}, nil
}

type ExecOptions struct{}

func (h *ExecHandler) Handle(w http.ResponseWriter, req *http.Request, opts ExecOptions) {
	ctx := req.Context()
	log := ctrl.LoggerFrom(ctx)

	strms, ok := NewStreams(req, w, StreamsOptions{
		Stdin:              true, // TODO: Make these configurable via ExecOptions.
		Stdout:             true,
		Stderr:             false,
		TTY:                true,
		SupportedProtocols: h.supportedStreamProtocols,
		IdleTimeout:        h.streamIdleTimeout,
		CreationTimeout:    h.streamCreationTimeout,
	})
	if !ok {
		// error is handled by NewStreams
		return
	}
	defer func() {
		if err := strms.Close(); err != nil {
			log.Error(err, "Error closing streams")
		}
	}()

	err := h.exec.Exec(ctx, strms.Stdin(), strms.Stdout(), strms.Resize())
	if err != nil {
		exitErr, ok := err.(utilexec.ExitError)
		if !ok || !exitErr.Exited() {
			log.Error(err, "Error running exec")
			_ = strms.WriteStatus(apierrors.NewInternalError(fmt.Errorf("error running exec: %w", err)))
			return
		}

		rc := exitErr.ExitStatus()
		_ = strms.WriteStatus(&apierrors.StatusError{ErrStatus: metav1.Status{
			Status: metav1.StatusFailure,
			Reason: remotecommandconsts.NonZeroExitCodeReason,
			Details: &metav1.StatusDetails{
				Causes: []metav1.StatusCause{
					{
						Type:    remotecommandconsts.ExitCodeCauseType,
						Message: fmt.Sprintf("%d", rc),
					},
				},
			},
			Message: fmt.Sprintf("command terminated with non-zero exit code: %v", exitErr),
		}})
		return
	}

	_ = strms.WriteStatus(&apierrors.StatusError{ErrStatus: metav1.Status{
		Status: metav1.StatusSuccess,
	}})
}
