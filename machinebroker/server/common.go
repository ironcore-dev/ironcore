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

package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/go-logr/logr"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (s *Server) loggerFrom(ctx context.Context, keysWithValues ...interface{}) logr.Logger {
	return ctrl.LoggerFrom(ctx, keysWithValues...)
}

const idLength = 63

func (s *Server) hashID(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		h.Write([]byte(part))
	}
	data := h.Sum(nil)
	encoded := hex.EncodeToString(data)
	return encoded[:idLength] // TODO: What about purely numerical IDs?
}

func (s *Server) generateID() string {
	data := make([]byte, 32)
	for {
		_, _ = rand.Read(data)
		id := hex.EncodeToString(data)

		// Truncated versions of the id should not be numerical.
		if _, err := strconv.ParseInt(id[:12], 10, 64); err != nil {
			continue
		}

		return id[:idLength]
	}
}

func (s *Server) convertOnmetalIPs(ips []commonv1alpha1.IP) []string {
	res := make([]string, len(ips))
	for i, ip := range ips {
		res[i] = ip.String()
	}
	return res
}

func (s *Server) getOnmetalIPsIPFamilies(ips []commonv1alpha1.IP) []corev1.IPFamily {
	res := make([]corev1.IPFamily, len(ips))
	for i, ip := range ips {
		res[i] = ip.Family()
	}
	return res
}

func (s *Server) onmetalIPsToOnmetalIPSources(ips []commonv1alpha1.IP) []networkingv1alpha1.IPSource {
	res := make([]networkingv1alpha1.IPSource, len(ips))
	for i := range ips {
		res[i] = networkingv1alpha1.IPSource{
			Value: &ips[i],
		}
	}
	return res
}

func (s *Server) parseIPs(ipStrings []string) ([]commonv1alpha1.IP, error) {
	var ips []commonv1alpha1.IP
	for _, ipString := range ipStrings {
		ip, err := commonv1alpha1.ParseIP(ipString)
		if err != nil {
			return nil, fmt.Errorf("error parsing ip %q: %w", ipString, err)
		}

		ips = append(ips, ip)
	}
	return ips, nil
}
