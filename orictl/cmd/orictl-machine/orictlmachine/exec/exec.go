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

package exec

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	clicommon "github.com/onmetal/onmetal-api/orictl/cli/common"
	"github.com/onmetal/onmetal-api/orictl/cmd/orictl-machine/orictlmachine/common"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/httpstream/spdy"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/util/term"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Command(streams clicommon.Streams, clientFactory common.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "exec machine-id",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := ctrl.LoggerFrom(ctx)

			client, cleanup, err := clientFactory.New()
			if err != nil {
				return err
			}
			defer func() {
				if err := cleanup(); err != nil {
					log.Error(err, "Error cleaning up")
				}
			}()

			machineID := args[0]

			return Run(ctx, streams, client, machineID)
		},
	}

	return cmd
}

func Run(ctx context.Context, streams clicommon.Streams, client ori.MachineRuntimeClient, machineID string) error {
	log := ctrl.LoggerFrom(ctx)
	res, err := client.Exec(ctx, &ori.ExecRequest{
		MachineId: machineID,
	})
	if err != nil {
		return fmt.Errorf("error running exec: %w", err)
	}

	u, err := url.ParseRequestURI(res.Url)
	if err != nil {
		return fmt.Errorf("error parsing request url %q: %w", res.Url, err)
	}

	log.V(1).Info("Got exec url", "URL", res.Url)

	var sizeQueue remotecommand.TerminalSizeQueue
	tty := term.TTY{
		In:     streams.In,
		Out:    streams.Out,
		Raw:    true,
		TryDev: true,
	}
	if size := tty.GetSize(); size != nil {
		// fake resizing +1 and then back to normal so that attach-detach-reattach will result in the
		// screen being redrawn
		sizePlusOne := *size
		sizePlusOne.Width++
		sizePlusOne.Height++

		// this call spawns a goroutine to monitor/update the terminal size
		sizeQueue = tty.MonitorSize(&sizePlusOne, size)
	}

	roundTripper := spdy.NewRoundTripperWithConfig(spdy.RoundTripperConfig{
		TLS:        http.DefaultTransport.(*http.Transport).TLSClientConfig,
		Proxier:    http.ProxyFromEnvironment,
		PingPeriod: 5 * time.Second,
	})
	exec, err := remotecommand.NewSPDYExecutorForTransports(roundTripper, roundTripper, http.MethodGet, u)
	if err != nil {
		return fmt.Errorf("error creating remote command executor: %w", err)
	}

	_, _ = fmt.Fprintln(os.Stderr, "If you don't see a command prompt, try pressing enter.")
	return tty.Safe(func() error {
		return exec.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdin:             tty.In,
			Stdout:            tty.Out,
			Tty:               true,
			TerminalSizeQueue: sizeQueue,
		})
	})
}
