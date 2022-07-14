package cmd

import (
	"time"

	"github.com/onmetal/onmetal-api/machinepoollet/server"
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

// AddFlags adds the flags to the pflag.FlagSet.
func (o *ServingOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.DisableAuth, "insecure-disable-auth", o.DisableAuth, "whether to completely disable authN/Z. Insecure and discouraged.")
	fs.StringVar(&o.HostnameOverride, "hostname-override", o.HostnameOverride, "Override for the hostname.")
	fs.StringVar(&o.Address, ":20250", o.Address, "Address to listen / serve on.")
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

// AddFlags adds the flags to the pflag.FlagSet.
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	o.Serving.AddFlags(fs)
	o.Auth.AddFlags(fs)
}

// ServerOptions produces server.Options.
func (o *ServerOptions) ServerOptions(machinePoolName string, exec server.MachineExec) server.Options {
	return o.Serving.ServerOptions(exec, o.Auth.AuthOptions(machinePoolName))
}
