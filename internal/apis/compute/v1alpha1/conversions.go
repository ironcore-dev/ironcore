// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"fmt"

	"github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

func addConversionFuncs(scheme *runtime.Scheme) error {
	if err := scheme.AddFieldLabelConversionFunc(
		SchemeGroupVersion.WithKind("Machine"),
		func(label, value string) (internalLabel, internalValue string, err error) {
			switch label {
			case "metadata.name", "metadata.namespace",
				v1alpha1.MachineMachinePoolRefNameField,
				v1alpha1.MachineMachineClassRefNameField:
				return label, value, nil
			default:
				return "", "", fmt.Errorf("field label not supported: %s", label)
			}
		},
	); err != nil {
		return err
	}
	return nil
}
