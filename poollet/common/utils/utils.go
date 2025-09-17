// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"strings"

	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"

	"k8s.io/kubectl/pkg/util/fieldpath"
)

type ID struct {
	Type string
	ID   string
}

func MakeID(typ, id string) ID {
	return ID{
		Type: typ,
		ID:   id,
	}
}

func (i *ID) UnmarshalText(data []byte) error {
	s := string(data)

	// Trim the quotes and split the type and ID.
	parts := strings.Split(strings.Trim(s, "\""), "://")
	if len(parts) != 2 {
		return fmt.Errorf("invalid ID: %q", s)
	}
	i.Type, i.ID = parts[0], parts[1]
	return nil
}

func (i *ID) String() string {
	return fmt.Sprintf("%s://%s", i.Type, i.ID)
}

func ParseID(s string) (ID, error) {
	var id ID
	return id, id.UnmarshalText([]byte(s))
}

// DownwardAPILabel makes a downward api label name from the given name.
func DownwardAPILabel(label_prefix, name string) string {
	return label_prefix + name
}

// DownwardAPIAnnotation makes a downward api annotation name from the given name.
func DownwardAPIAnnotation(annotation_prefix, name string) string {
	return annotation_prefix + name
}

func PrepareDownwardAPILabels(
	obj interface{},
	downwardAPILabels map[string]string,
	prefix string,
) (map[string]string, error) {
	labels := make(map[string]string)

	// Use reflection or type assertion to get fields by path
	for name, fieldPath := range downwardAPILabels {
		var value string
		var err error

		switch o := obj.(type) {
		case *computev1alpha1.Machine:
			value, err = fieldpath.ExtractFieldPathAsString(o, fieldPath)
		case *computev1alpha1.Reservation:
			value, err = fieldpath.ExtractFieldPathAsString(o, fieldPath)
		case *networkingv1alpha1.Network:
			value, err = fieldpath.ExtractFieldPathAsString(o, fieldPath)
		case *networkingv1alpha1.NetworkInterface:
			value, err = fieldpath.ExtractFieldPathAsString(o, fieldPath)
		case *storagev1alpha1.Volume:
			value, err = fieldpath.ExtractFieldPathAsString(o, fieldPath)
		case *storagev1alpha1.Bucket:
			value, err = fieldpath.ExtractFieldPathAsString(o, fieldPath)
		default:
			return nil, fmt.Errorf("unsupported type for downward API label extraction")
		}

		if err != nil {
			return nil, fmt.Errorf("error extracting downward api label %q: %w", name, err)
		}
		labels[DownwardAPILabel(prefix, name)] = value
	}
	return labels, nil
}
