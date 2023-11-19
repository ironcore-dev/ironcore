// Copyright 2022 IronCore authors
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
