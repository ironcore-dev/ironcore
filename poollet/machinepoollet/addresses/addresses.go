// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package addresses

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	KubernetesServiceName         = "KUBERNETES_SERVICE_NAME"
	KubernetesPodNamespaceEnvVar  = "KUBERNETES_POD_NAMESPACE"
	KubernetesClusterDomainEnvVar = "KUBERNETES_CLUSTER_DOMAIN"
)

const (
	DefaultKubernetesClusterDomain = "cluster.local"
)

var (
	ErrNotInCluster = fmt.Errorf("unable to load in-cluster addresses, %s and %s must be defined",
		KubernetesServiceName, KubernetesPodNamespaceEnvVar)
)

type GetOptions struct {
	Filename         string
	IPOverride       string
	HostnameOverride string
}

func (o *GetOptions) ApplyOptions(opts []GetOption) {
	for _, opt := range opts {
		opt.ApplyToGet(o)
	}
}

func (o *GetOptions) ApplyToGet(o2 *GetOptions) {
	if o.Filename != "" {
		o2.Filename = o.Filename
	}
	if o.IPOverride != "" {
		o2.IPOverride = o.IPOverride
	}
	if o.HostnameOverride != "" {
		o2.HostnameOverride = o.HostnameOverride
	}
}

func (o *GetOptions) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Filename, "addresses-filename", "", "File pointing to endpoints address configuration.")
	fs.StringVar(&o.IPOverride, "addresses-ip-override", "", "Machine pool address IP to use.")
	fs.StringVar(&o.HostnameOverride, "addresses-hostname-override", "", "Machine pool address hostname to use.")
}

type GetOption interface {
	ApplyToGet(o *GetOptions)
}

func Load(data []byte) ([]computev1alpha1.MachinePoolAddress, error) {
	type Config struct {
		Addresses []computev1alpha1.MachinePoolAddress `json:"addresses"`
	}

	file := &Config{}
	if err := yaml.NewYAMLOrJSONDecoder(bytes.NewBuffer(data), 4096).Decode(file); err != nil {
		return nil, fmt.Errorf("error unmarshalling address file: %w", err)
	}
	return file.Addresses, nil
}

func LoadFromFile(filename string) ([]computev1alpha1.MachinePoolAddress, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file at %q: %w", filename, err)
	}

	return Load(data)
}

func LocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("error getting interfaces: %w", err)
	}

	for _, addr := range addrs {
		var (
			ip net.IP
			ok bool
		)
		switch v := addr.(type) {
		case *net.IPNet:
			ip, ok = v.IP, true
		case *net.IPAddr:
			ip, ok = v.IP, true
		}
		if ok && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() {
			return ip.String(), nil
		}
	}
	return "", nil
}

func IsInCluster() bool {
	podIP := os.Getenv(KubernetesServiceName)
	return podIP != ""
}

func Get(opts ...GetOption) ([]computev1alpha1.MachinePoolAddress, error) {
	o := &GetOptions{}
	o.ApplyOptions(opts)

	if o.Filename != "" {
		return LoadFromFile(o.Filename)
	}

	if !IsInCluster() {
		ip := o.IPOverride
		if ip == "" {
			localIP, err := LocalIP()
			if err != nil {
				return nil, fmt.Errorf("error getting local ip: %w", err)
			}

			ip = localIP
		}

		hostname := o.HostnameOverride
		if hostname == "" {
			h, err := os.Hostname()
			if err != nil {
				return nil, fmt.Errorf("error getting hostname: %w", err)
			}

			hostname = strings.TrimSpace(h)
		}

		var addresses []computev1alpha1.MachinePoolAddress
		if ip != "" {
			addresses = append(addresses, computev1alpha1.MachinePoolAddress{
				Type:    computev1alpha1.MachinePoolInternalIP,
				Address: ip,
			})
		}
		if hostname != "" {
			addresses = append(addresses, computev1alpha1.MachinePoolAddress{
				Type:    computev1alpha1.MachinePoolInternalDNS,
				Address: hostname,
			})
		}
		return addresses, nil
	}

	return InCluster()
}

func InCluster() ([]computev1alpha1.MachinePoolAddress, error) {
	serviceName := os.Getenv(KubernetesServiceName)
	namespace := os.Getenv(KubernetesPodNamespaceEnvVar)
	clusterDomain := os.Getenv(KubernetesClusterDomainEnvVar)

	if serviceName == "" || namespace == "" {
		return nil, ErrNotInCluster
	}

	if clusterDomain == "" {
		clusterDomain = DefaultKubernetesClusterDomain
	}

	internalDNS := fmt.Sprintf("%s.%s.svc.%s", serviceName, namespace, clusterDomain)

	return []computev1alpha1.MachinePoolAddress{
		{
			Type:    computev1alpha1.MachinePoolInternalIP,
			Address: serviceName,
		},
		{
			Type:    computev1alpha1.MachinePoolInternalDNS,
			Address: internalDNS,
		},
	}, nil
}
