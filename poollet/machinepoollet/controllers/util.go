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
	ori "github.com/onmetal/onmetal-api/ori/apis/machine/v1alpha1"
	utilslices "github.com/onmetal/onmetal-api/utils/slices"
	"golang.org/x/exp/constraints"
)

func Max[T constraints.Ordered](x, y T) T {
	if x < y {
		return y
	}
	return x
}

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
