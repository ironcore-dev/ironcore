// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"encoding/json"
	"fmt"

	poolletutils "github.com/ironcore-dev/ironcore/poollet/common/utils"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectPtr[E any] interface {
	*E
	client.Object
}

func ObjectSliceToMapByName[S ~[]E, ObjPtr ObjectPtr[E], E any](objs S) map[string]ObjPtr {
	res := make(map[string]ObjPtr)
	for i := range objs {
		obj := ObjPtr(&objs[i])
		res[obj.GetName()] = obj
	}
	return res
}

func ObjectByNameGetter[Obj client.Object](resource schema.GroupResource, objectByName map[string]Obj) func(name string) (Obj, error) {
	return func(name string) (Obj, error) {
		object, ok := objectByName[name]
		if !ok {
			var zero Obj
			return zero, apierrors.NewNotFound(resource, name)
		}

		return object, nil
	}
}

func ObjectSliceToByNameGetter[S ~[]E, ObjPtr ObjectPtr[E], E any](resource schema.GroupResource, objs S) func(name string) (ObjPtr, error) {
	objectByName := ObjectSliceToMapByName[S, ObjPtr](objs)
	return ObjectByNameGetter(resource, objectByName)
}

func ClientObjectGetter[ObjPtr ObjectPtr[Obj], Obj any](
	ctx context.Context,
	c client.Client,
	namespace string,
) func(name string) (ObjPtr, error) {
	return func(name string) (ObjPtr, error) {
		objPtr := ObjPtr(new(Obj))
		if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, objPtr); err != nil {
			return nil, err
		}
		return objPtr, nil
	}
}

func PrepareDownwardAPILabels(
	labelSource map[string]string,
	downwardAPILabels map[string]string,
	prefix string,
) map[string]string {
	preparedLabels := make(map[string]string)
	for downwardAPILabelName, defaultLabelName := range downwardAPILabels {
		key := poolletutils.DownwardAPILabel(prefix, downwardAPILabelName)
		value := labelSource[key]
		if value == "" {
			value = labelSource[defaultLabelName]
		}
		if value != "" {
			preparedLabels[key] = value
		}
	}
	return preparedLabels
}

func EncodeAnnotationsAnnotation(annotations map[string]string) (string, error) {
	data, err := json.Marshal(annotations)
	if err != nil {
		return "", fmt.Errorf("error marshalling annotations: %w", err)
	}
	return string(data), nil
}

func DecodeAnnotationsAnnotation(data string) (map[string]string, error) {
	var annotations map[string]string
	if err := json.Unmarshal([]byte(data), &annotations); err != nil {
		return nil, fmt.Errorf("error unmarshalling annotations: %w", err)
	}
	return annotations, nil
}

func EncodeLabelsAnnotation(labels map[string]string) (string, error) {
	data, err := json.Marshal(labels)
	if err != nil {
		return "", fmt.Errorf("error mashalling labels: %w", err)
	}
	return string(data), nil
}

func DecodeLabelsAnnotations(data string) (map[string]string, error) {
	var labels map[string]string
	if err := json.Unmarshal([]byte(data), &labels); err != nil {
		return nil, fmt.Errorf("error unmarshalling labels: %w", err)
	}
	return labels, nil
}
