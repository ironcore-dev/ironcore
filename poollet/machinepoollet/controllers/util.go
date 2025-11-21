// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"strconv"

	iri "github.com/ironcore-dev/ironcore/iri/apis/machine/v1alpha1"
	"github.com/ironcore-dev/ironcore/utils/generic"
	utilslices "github.com/ironcore-dev/ironcore/utils/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func setLabel(om *metav1.ObjectMeta, lblKey string, lblVal string) {
	if len(lblVal) < 1 {
		return
	}

	if om.Labels == nil {
		om.Labels = make(map[string]string)
	}

	om.Labels[lblKey] = lblVal
}
