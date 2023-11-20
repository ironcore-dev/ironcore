// Copyright 2023 IronCore authors
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

	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/utils/generic"
	utilslices "github.com/ironcore-dev/ironcore/utils/slices"
)

func FindNewIRINetworkInterfaces(desiredIRINics, existingIRINics []*iri.NetworkInterface) []*iri.NetworkInterface {
	var (
		existingIRINicNames = utilslices.ToSetFunc(existingIRINics, (*iri.NetworkInterface).GetName)
		newIRINics          []*iri.NetworkInterface
	)
	for _, desiredIRINic := range desiredIRINics {
		if existingIRINicNames.Has(desiredIRINic.Name) {
			continue
		}

		newIRINics = append(newIRINics, desiredIRINic)
	}
	return newIRINics
}

func FindNewIRIVolumes(desiredIRIVolumes, existingIRIVolumes []*iri.Volume) []*iri.Volume {
	var (
		existingIRIVolumeNames = utilslices.ToSetFunc(existingIRIVolumes, (*iri.Volume).GetName)
		newIRIVolumes          []*iri.Volume
	)
	for _, desiredIRIVolume := range desiredIRIVolumes {
		if existingIRIVolumeNames.Has(desiredIRIVolume.Name) {
			continue
		}

		newIRIVolumes = append(newIRIVolumes, desiredIRIVolume)
	}
	return newIRIVolumes
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
