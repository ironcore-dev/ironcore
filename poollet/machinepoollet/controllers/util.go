// Copyright 2023 OnMetal authors
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

package controllers

import (
	"fmt"
	"strconv"

	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	"github.com/onmetal/onmetal-api/utils/generic"
	utilslices "github.com/onmetal/onmetal-api/utils/slices"
)

func FindNewORINetworkInterfaces(desiredORINics, existingORINics []*ori.NetworkInterface) []*ori.NetworkInterface {
	var (
		existingORINicNames = utilslices.ToSetFunc(existingORINics, (*ori.NetworkInterface).GetName)
		newORINics          []*ori.NetworkInterface
	)
	for _, desiredORINic := range desiredORINics {
		if existingORINicNames.Has(desiredORINic.Name) {
			continue
		}

		newORINics = append(newORINics, desiredORINic)
	}
	return newORINics
}

func FindNewORIVolumes(desiredORIVolumes, existingORIVolumes []*ori.Volume) []*ori.Volume {
	var (
		existingORIVolumeNames = utilslices.ToSetFunc(existingORIVolumes, (*ori.Volume).GetName)
		newORIVolumes          []*ori.Volume
	)
	for _, desiredORIVolume := range desiredORIVolumes {
		if existingORIVolumeNames.Has(desiredORIVolume.Name) {
			continue
		}

		newORIVolumes = append(newORIVolumes, desiredORIVolume)
	}
	return newORIVolumes
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func getAndParseFromStringMap[E any](annotations map[string]string, key string, parse func(string) (E, error)) (E, error) {
	s, ok := annotations[key]
	if !ok {
		return generic.Zero[E](), fmt.Errorf("no value found at key %s", key)
	}

	e, err := parse(s)
	if err != nil {
		return e, fmt.Errorf("error parsing key %s data %s: %w", key, s, err)
	}

	return e, nil
}
