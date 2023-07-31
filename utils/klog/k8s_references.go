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

package klog

import (
	"bytes"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type KUIDMetadata interface {
	klog.KMetadata
	GetUID() types.UID
}

type ObjectUIDRef struct {
	Namespace string    `json:"namespace,omitempty"`
	Name      string    `json:"name"`
	UID       types.UID `json:"uid,omitempty"`
}

func (r ObjectUIDRef) String() string {
	if r.Namespace == "" && r.UID == "" {
		return r.Name
	}

	var sb strings.Builder
	n := len(r.Name)
	if r.Namespace != "" {
		// <namespace>/
		n += len(r.Namespace) + 1
	}
	if r.UID != "" {
		// (<uid>)
		n += 2 + len(r.UID) + 1
	}

	sb.Grow(n)
	if r.Namespace != "" {
		sb.WriteString(r.Namespace)
		sb.WriteRune('/')
	}

	sb.WriteString(r.Name)

	if r.UID != "" {
		sb.WriteRune('(')
		sb.WriteString(string(r.UID))
		sb.WriteRune(')')
	}

	return sb.String()
}

// KObjUID returns a ObjectUIDRef from the given KUIDMetadata.
// If the given KUIDMetadata is nil, a zero ObjectUIDRef is returned.
func KObjUID(obj KUIDMetadata) ObjectUIDRef {
	if obj == nil {
		return ObjectUIDRef{}
	}
	return ObjectUIDRef{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		UID:       obj.GetUID(),
	}
}

type kObjSliceLike struct {
	len int
	idx func(int) klog.KMetadata
}

func (ks kObjSliceLike) String() string {
	objectRefs := ks.process()
	return fmt.Sprintf("%v", objectRefs)
}

func (ks kObjSliceLike) MarshalLog() any {
	return ks.process()
}

func (ks kObjSliceLike) process() []any {
	refs := make([]any, 0, ks.len)
	for i := 0; i < ks.len; i++ {
		item := ks.idx(i)
		if item == nil {
			refs = append(refs, nil)
		} else {
			refs = append(refs, klog.KObj(item))
		}
	}
	return refs
}

var nilToken = []byte("null")

func (ks kObjSliceLike) WriteText(out *bytes.Buffer) {
	out.WriteRune('[')
	defer out.WriteRune(']')
	for i := 0; i < ks.len; i++ {
		if i > 0 {
			out.WriteRune(',')
		}
		item := ks.idx(i)
		if item == nil {
			out.Write(nilToken)
		} else {
			klog.KObj(item).WriteText(out)
		}
	}
}

func KObjStructSlice[S ~[]OStruct, KMetadata interface {
	klog.KMetadata
	*OStruct
}, OStruct any](s S) any {
	return kObjSliceLike{
		len: len(s),
		idx: func(i int) klog.KMetadata {
			return KMetadata(&s[i])
		},
	}
}
