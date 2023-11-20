// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import "github.com/gogo/protobuf/proto"

type Object interface {
	proto.Message
	GetMetadata() *ObjectMetadata
}
