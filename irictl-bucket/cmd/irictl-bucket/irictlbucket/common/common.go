// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"
	"time"

	iri "github.com/ironcore-dev/ironcore/iri/apis/bucket/v1alpha1"
	iriremotebucket "github.com/ironcore-dev/ironcore/iri/remote/bucket"
	"github.com/ironcore-dev/ironcore/irictl-bucket/renderers"
	irictlcmd "github.com/ironcore-dev/ironcore/irictl/cmd"
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
	New() (iri.BucketRuntimeClient, func() error, error)
}

type ClientOptions struct {
	Address string
}

func (o *ClientOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Address, "address", "", "Address to the iri server.")
}

func (o *ClientOptions) New() (iri.BucketRuntimeClient, func() error, error) {
	address, err := iriremotebucket.GetAddressWithTimeout(3*time.Second, o.Address)
	if err != nil {
		return nil, nil, err
	}

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("error dialing: %w", err)
	}

	return iri.NewBucketRuntimeClient(conn), conn.Close, nil
}

func NewOutputOptions() *irictlcmd.OutputOptions {
	return &irictlcmd.OutputOptions{
		Registry: Renderer,
	}
}

var (
	BucketAliases      = []string{"buckets"}
	BucketClassAliases = []string{"bucketchineclasses"}
)
