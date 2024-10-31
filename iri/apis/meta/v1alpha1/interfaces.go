// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"google.golang.org/protobuf/proto"
)

type Object interface {
	proto.Message
	GetMetadata() *ObjectMetadata
	Reset()
	String() string
	ProtoMessage()
}
