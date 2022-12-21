package v1alpha1

import (
	"fmt"
	"sort"
)

func (tgt *LoadBalancerTargetSpec) Key() string {
	portStrings := make([]string, len(tgt.Ports))
	for i, port := range tgt.Ports {
		portStrings[i] = fmt.Sprintf("%s:%d-%d", port.Protocol.String(), port.Port, port.EndPort)
	}
	sort.Strings(portStrings)
	return fmt.Sprintf("%s%v", tgt.Ip, portStrings)
}
