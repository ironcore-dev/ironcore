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

package client

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apiserver/pkg/server/egressselector"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
)

// ConnectionInfo provides the information needed to connect to a machine poollet.
type ConnectionInfo struct {
	Scheme                         string
	Hostname                       string
	Port                           string
	Transport                      http.RoundTripper
	InsecureSkipTLSVerifyTransport http.RoundTripper
}

type ConnectionInfoGetter interface {
	GetConnectionInfo(ctx context.Context, machinePoolName string) (*ConnectionInfo, error)
}

// MachinePoolGetter defines an interface for looking up a node by name
type MachinePoolGetter interface {
	Get(ctx context.Context, name string, options metav1.GetOptions) (*computev1alpha1.MachinePool, error)
}

// MachinePoolGetterFunc allows implementing MachinePoolGetter with a function
type MachinePoolGetterFunc func(ctx context.Context, name string, options metav1.GetOptions) (*computev1alpha1.MachinePool, error)

// Get fetches information via MachinePoolGetterFunc.
func (f MachinePoolGetterFunc) Get(ctx context.Context, name string, options metav1.GetOptions) (*computev1alpha1.MachinePool, error) {
	return f(ctx, name, options)
}

// MachinePoolConnectionInfoGetter obtains connection info from the status of a MachinePool API object
type MachinePoolConnectionInfoGetter struct {
	// machinePools is used to look up MachinePool objects
	machinePools MachinePoolGetter
	// scheme is the scheme to use to connect to all machinepoollets.
	scheme string
	// defaultPort is the port to use if no machinepoollet endpoint port is recorded in the node status
	defaultPort int
	// transport is the transport to use to send a request to all machinepoollets.
	transport http.RoundTripper
	// insecureSkipTLSVerifyTransport is the transport to use if the kube-apiserver wants to skip verifying the TLS certificate of the machinepoolllet
	insecureSkipTLSVerifyTransport http.RoundTripper
	// preferredAddressTypes specifies the preferred order to use to find a node address
	preferredAddressTypes []computev1alpha1.MachinePoolAddressType
}

type MachinePoolletClientConfig struct {
	// Port specifies the default port - used if no information about Machinepoollet port can be found in Node.NodeStatus.DaemonEndpoints.
	Port uint

	// ReadOnlyPort specifies the Port for ReadOnly communications.
	ReadOnlyPort uint

	// PreferredAddressTypes - used to select an address from Node.NodeStatus.Addresses
	PreferredAddressTypes []string

	// TLSClientConfig contains settings to enable transport layer security
	rest.TLSClientConfig

	// Server requires Bearer authentication
	BearerToken string `datapolicy:"token"`

	// HTTPTimeout is used by the client to timeout http requests to MachinePoollet.
	HTTPTimeout time.Duration

	// Dial is a custom dialer used for the client
	Dial utilnet.DialFunc

	// Lookup will give us a dialer if the egress selector is configured for it
	Lookup egressselector.Lookup
}

func (c *MachinePoolletClientConfig) transportConfig() *transport.Config {
	cfg := &transport.Config{
		TLS: transport.TLSConfig{
			CAFile:     c.CAFile,
			CAData:     c.CAData,
			CertFile:   c.CertFile,
			CertData:   c.CertData,
			KeyFile:    c.KeyFile,
			KeyData:    c.KeyData,
			NextProtos: c.NextProtos,
		},
		BearerToken: c.BearerToken,
	}
	if !cfg.HasCA() {
		cfg.TLS.Insecure = true
	}
	return cfg
}

// MakeTransport creates a secure RoundTripper for HTTP Transport.
func MakeTransport(config *MachinePoolletClientConfig) (http.RoundTripper, error) {
	return makeTransport(config, false)
}

// MakeInsecureTransport creates an insecure RoundTripper for HTTP Transport.
func MakeInsecureTransport(config *MachinePoolletClientConfig) (http.RoundTripper, error) {
	return makeTransport(config, true)
}

// makeTransport creates a RoundTripper for HTTP Transport.
func makeTransport(config *MachinePoolletClientConfig, insecureSkipTLSVerify bool) (http.RoundTripper, error) {
	// do the insecureSkipTLSVerify on the pre-transport *before* we go get a potentially cached connection.
	// transportConfig always produces a new struct pointer.
	preTLSConfig := config.transportConfig()
	if insecureSkipTLSVerify && preTLSConfig != nil {
		preTLSConfig.TLS.Insecure = true
		preTLSConfig.TLS.CAData = nil
		preTLSConfig.TLS.CAFile = ""
	}

	tlsConfig, err := transport.TLSConfigFor(preTLSConfig)
	if err != nil {
		return nil, err
	}

	rt := http.DefaultTransport
	dialer := config.Dial
	if dialer == nil && config.Lookup != nil {
		// Assuming EgressSelector if SSHTunnel is not turned on.
		// We will not get a dialer if egress selector is disabled.
		networkContext := egressselector.Cluster.AsNetworkContext()
		dialer, err = config.Lookup(networkContext)
		if err != nil {
			return nil, fmt.Errorf("failed to get context dialer for 'cluster': got %v", err)
		}
	}
	if dialer != nil || tlsConfig != nil {
		// If SSH Tunnel is turned on
		rt = utilnet.SetOldTransportDefaults(&http.Transport{
			DialContext:     dialer,
			TLSClientConfig: tlsConfig,
		})
	}

	return transport.HTTPWrappersForConfig(config.transportConfig(), rt)
}

// NoMatchError is a typed implementation of the error interface. It indicates a failure to get a matching Node.
type NoMatchError struct {
	addresses []computev1alpha1.MachinePoolAddress
}

// Error is the implementation of the conventional interface for
// representing an error condition, with the nil value representing no error.
func (e *NoMatchError) Error() string {
	return fmt.Sprintf("no preferred addresses found; known addresses: %v", e.addresses)
}

// GetPreferredMachinePoolAddress returns the address of the provided node, using the provided preference order.
// If none of the preferred address types are found, an error is returned.
func GetPreferredMachinePoolAddress(machinePool *computev1alpha1.MachinePool, preferredAddressTypes []computev1alpha1.MachinePoolAddressType) (string, error) {
	for _, addressType := range preferredAddressTypes {
		for _, address := range machinePool.Status.Addresses {
			if address.Type == addressType {
				return address.Address, nil
			}
		}
	}
	return "", &NoMatchError{addresses: machinePool.Status.Addresses}
}

// GetConnectionInfo retrieves connection info from the status of a Node API object.
func (k *MachinePoolConnectionInfoGetter) GetConnectionInfo(ctx context.Context, machinePoolName string) (*ConnectionInfo, error) {
	machinePool, err := k.machinePools.Get(ctx, machinePoolName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Find a machinepoollet-reported address, using preferred address type
	host, err := GetPreferredMachinePoolAddress(machinePool, k.preferredAddressTypes)
	if err != nil {
		return nil, err
	}

	// Use the machinepoollet-reported port, if present
	port := int(machinePool.Status.DaemonEndpoints.MachinepoolletEndpoint.Port)
	if port <= 0 {
		port = k.defaultPort
	}

	return &ConnectionInfo{
		Scheme:                         k.scheme,
		Hostname:                       host,
		Port:                           strconv.Itoa(port),
		Transport:                      k.transport,
		InsecureSkipTLSVerifyTransport: k.insecureSkipTLSVerifyTransport,
	}, nil
}

func NewMachinePoolConnectionInfoGetter(machinePools MachinePoolGetter, config MachinePoolletClientConfig) (ConnectionInfoGetter, error) {
	transport, err := MakeTransport(&config)
	if err != nil {
		return nil, err
	}
	insecureSkipTLSVerifyTransport, err := MakeInsecureTransport(&config)
	if err != nil {
		return nil, err
	}

	var types []computev1alpha1.MachinePoolAddressType
	for _, t := range config.PreferredAddressTypes {
		types = append(types, computev1alpha1.MachinePoolAddressType(t))
	}

	return &MachinePoolConnectionInfoGetter{
		machinePools:                   machinePools,
		scheme:                         "https",
		defaultPort:                    int(config.Port),
		transport:                      transport,
		insecureSkipTLSVerifyTransport: insecureSkipTLSVerifyTransport,

		preferredAddressTypes: types,
	}, nil
}
