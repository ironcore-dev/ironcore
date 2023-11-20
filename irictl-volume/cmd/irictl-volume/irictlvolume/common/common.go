// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"
	"time"

	iri "github.com/ironcore-dev/ironcore/iri/apis/volume/v1alpha1"
	iriremotevolume "github.com/ironcore-dev/ironcore/iri/remote/volume"
	"github.com/ironcore-dev/ironcore/irictl-volume/renderers"
	clicommon "github.com/ironcore-dev/ironcore/irictl/cmd"
	"github.com/ironcore-dev/ironcore/irictl/renderer"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var Renderer = renderer.NewRegistry()

func init() {
	if err := renderers.AddToRegistry(Renderer); err != nil {
		panic(err)
	}
}

type ClientFactory interface {
	New() (iri.VolumeRuntimeClient, func() error, error)
}

type ClientOptions struct {
	Address string
}

func (o *ClientOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Address, "address", "", "Address to the iri server.")
}

func (o *ClientOptions) New() (iri.VolumeRuntimeClient, func() error, error) {
	address, err := iriremotevolume.GetAddressWithTimeout(3*time.Second, o.Address)
	if err != nil {
		return nil, nil, err
	}

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("error dialing: %w", err)
	}

	return iri.NewVolumeRuntimeClient(conn), conn.Close, nil
}

func NewOutputOptions() *clicommon.OutputOptions {
	return &clicommon.OutputOptions{
		Registry: Renderer,
	}
}

var (
	VolumeAliases = []string{"volumes", "vol", "vols"}
)
