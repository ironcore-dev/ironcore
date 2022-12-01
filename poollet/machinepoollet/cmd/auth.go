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

package cmd

import (
	"time"

	"github.com/onmetal/onmetal-api/poollet/machinepoollet/server"
	"github.com/spf13/pflag"
)

// AuthOptions are options for configuring server.Options.AuthOptions.
type AuthOptions struct {
	ClientCAFile string
	Anonymous    bool
}

// AddFlags adds the flags to the pflag.FlagSet.
func (o *AuthOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ClientCAFile, "client-ca-file", o.ClientCAFile, "Path pointing to a PEM-encoded CA file for verifying client requests.")
	fs.BoolVar(&o.Anonymous, "anonymous", o.Anonymous, "Whether to authenticate unknown users as 'system:anonymous' or not.")
}

// AuthOptions produces the server.AuthOptions.
func (o *AuthOptions) AuthOptions(machinePoolName string) server.AuthOptions {
	return server.AuthOptions{
		MachinePoolName: machinePoolName,
		Authentication: server.AuthenticationOptions{
			ClientCAFile: o.ClientCAFile,
		},
		Authorization: server.AuthorizationOptions{
			Anonymous: o.Anonymous,
		},
	}
}

// ServingOptions are options for configuring the serving part of server.Options.
type ServingOptions struct {
	DisableAuth      bool
	HostnameOverride string
	Address          string
	CertDir          string

	StreamCreationTimeout time.Duration
	StreamIdleTimeout     time.Duration
	ShutdownTimeout       time.Duration
}

func NewServingOptions() *ServingOptions {
	return &ServingOptions{
		Address: ":20250",
	}
}

// AddFlags adds the flags to the pflag.FlagSet.
func (o *ServingOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.DisableAuth, "insecure-disable-auth", o.DisableAuth, "whether to completely disable authN/Z. Insecure and discouraged.")
	fs.StringVar(&o.HostnameOverride, "hostname-override", o.HostnameOverride, "Override for the hostname.")
	fs.StringVar(&o.Address, "address", o.Address, "Address to listen / serve on.")
	fs.StringVar(&o.CertDir, "cert-dir", o.CertDir, "The directory that contains the server key and certificate."+
		"The files have to be named 'tls.crt' and 'tls.key'. If unset, "+
		"{TempDir}/onmetal-api-machinepool-server/serving-certs will be looked up for the certificates.")

	fs.DurationVar(&o.StreamCreationTimeout, "stream-creation-timeout", o.StreamCreationTimeout, "Timeout for creating streams.")
	fs.DurationVar(&o.StreamIdleTimeout, "stream-idle-timeout", o.StreamIdleTimeout, "Timeout for idle streams to be considered closed.")
	fs.DurationVar(&o.ShutdownTimeout, "shutdown-timeout", o.ShutdownTimeout, "Timeout for shutting down the http server.")
}

// ServerOptions produces server.Options.
func (o *ServingOptions) ServerOptions(exec server.MachineExec, authOpts server.AuthOptions) server.Options {
	return server.Options{
		MachineExec:           exec,
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

// ServerOptions couples together all options required to create server.Options.
type ServerOptions struct {
	Serving ServingOptions
	Auth    AuthOptions
}

func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		Serving: *NewServingOptions(),
		Auth:    AuthOptions{},
	}
}

// AddFlags adds the flags to the pflag.FlagSet.
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	o.Serving.AddFlags(fs)
	o.Auth.AddFlags(fs)
}

// ServerOptions produces server.Options.
func (o *ServerOptions) ServerOptions(machinePoolName string, exec server.MachineExec) server.Options {
	return o.Serving.ServerOptions(exec, o.Auth.AuthOptions(machinePoolName))
}
