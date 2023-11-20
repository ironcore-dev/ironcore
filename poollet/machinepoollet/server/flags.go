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

package server

import (
	"time"

	"github.com/go-logr/logr"
	irimachine "github.com/ironcore-dev/ironcore/iri/apis/machine"
	"github.com/spf13/pflag"
)

// AuthFlags are options for configuring server.Options.AuthFlags.
type AuthFlags struct {
	ClientCAFile string
	Anonymous    bool
}

// BindFlags adds the flags to the pflag.FlagSet.
func (o *AuthFlags) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ClientCAFile, "client-ca-file", o.ClientCAFile, "Path pointing to a PEM-encoded CA file for verifying client requests.")
	fs.BoolVar(&o.Anonymous, "anonymous", o.Anonymous, "Whether to authenticate unknown users as 'system:anonymous' or not.")
}

// AuthOptions produces the server.AuthOptions.
func (o *AuthFlags) AuthOptions(machinePoolName string) AuthOptions {
	return AuthOptions{
		MachinePoolName: machinePoolName,
		Authentication: AuthenticationOptions{
			ClientCAFile: o.ClientCAFile,
		},
		Authorization: AuthorizationOptions{
			Anonymous: o.Anonymous,
		},
	}
}

// ServingFlags are options for configuring the serving part of server.Options.
type ServingFlags struct {
	DisableAuth      bool
	HostnameOverride string
	Address          string
	CertDir          string

	StreamCreationTimeout time.Duration
	StreamIdleTimeout     time.Duration
	ShutdownTimeout       time.Duration
}

func NewServingOptions() *ServingFlags {
	return &ServingFlags{
		Address: ":20250",
	}
}

// BindFlags adds the flags to the pflag.FlagSet.
func (o *ServingFlags) BindFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.DisableAuth, "insecure-disable-auth", o.DisableAuth, "whether to completely disable authN/Z. Insecure and discouraged.")
	fs.StringVar(&o.HostnameOverride, "hostname-override", o.HostnameOverride, "Override for the hostname.")
	fs.StringVar(&o.Address, "address", o.Address, "Address to listen / serve on.")
	fs.StringVar(&o.CertDir, "cert-dir", o.CertDir, "The directory that contains the server key and certificate."+
		"The files have to be named 'tls.crt' and 'tls.key'. If unset, "+
		"{TempDir}/machinepoollet-server/serving-certs will be looked up for the certificates.")

	fs.DurationVar(&o.StreamCreationTimeout, "stream-creation-timeout", o.StreamCreationTimeout, "Timeout for creating streams.")
	fs.DurationVar(&o.StreamIdleTimeout, "stream-idle-timeout", o.StreamIdleTimeout, "Timeout for idle streams to be considered closed.")
	fs.DurationVar(&o.ShutdownTimeout, "shutdown-timeout", o.ShutdownTimeout, "Timeout for shutting down the http server.")
}

// ServerOptions produces server.Options.
func (o *ServingFlags) ServerOptions(machineRuntime irimachine.RuntimeService, log logr.Logger, authOpts AuthOptions) Options {
	return Options{
		MachineRuntime:        machineRuntime,
		Log:                   log,
		HostnameOverride:      o.HostnameOverride,
		Address:               o.Address,
		CertDir:               o.CertDir,
		Auth:                  authOpts,
		DisableAuth:           o.DisableAuth,
		StreamCreationTimeout: o.StreamCreationTimeout,
		StreamIdleTimeout:     o.StreamIdleTimeout,
		ShutdownTimeout:       o.ShutdownTimeout,
	}
}

// Flags couples together all options required to create server.Options.
type Flags struct {
	Serving ServingFlags
	Auth    AuthFlags
}

func NewServerFlags() *Flags {
	return &Flags{
		Serving: *NewServingOptions(),
		Auth:    AuthFlags{},
	}
}

// BindFlags adds the flags to the pflag.FlagSet.
func (o *Flags) BindFlags(fs *pflag.FlagSet) {
	o.Serving.BindFlags(fs)
	o.Auth.BindFlags(fs)
}

// ServerOptions produces server.Options.
func (o *Flags) ServerOptions(machinePoolName string, machineRuntime irimachine.RuntimeService, log logr.Logger) Options {
	return o.Serving.ServerOptions(machineRuntime, log, o.Auth.AuthOptions(machinePoolName))
}
