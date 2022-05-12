// Copyright 2022 OnMetal authors
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

package serializer

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

type subsetNegotiatedSerializer struct {
	accepts []func(info runtime.SerializerInfo) bool
	runtime.NegotiatedSerializer
}

func (s subsetNegotiatedSerializer) SupportedMediaTypes() []runtime.SerializerInfo {
	base := s.NegotiatedSerializer.SupportedMediaTypes()
	var filtered []runtime.SerializerInfo
	for _, info := range base {
		for _, accept := range s.accepts {
			if accept(info) {
				filtered = append(filtered, info)
				break
			}
		}
	}
	return filtered
}

// NoProtobuf is a function that disallows the use of protobuf.
func NoProtobuf(info runtime.SerializerInfo) bool {
	return info.MediaType != runtime.ContentTypeProtobuf
}

// SubsetNegotiatedSerializer is a runtime.NegotiatedSerializer that provides a whitelisted subset of the
// media types the provided serializer.CodecFactory provides.
func SubsetNegotiatedSerializer(codecs serializer.CodecFactory, accepts ...func(info runtime.SerializerInfo) bool) runtime.NegotiatedSerializer {
	return subsetNegotiatedSerializer{accepts, codecs}
}

// DefaultSubsetNegotiatedSerializer is the default onmetal serializer that does not use protobuf.
// Since our types *don't* implement protobuf encoding, and without removing the protobuf support,
// namespace deletion would fail (see issue https://github.com/kubernetes/kubernetes/issues/86666). As such,
// until we either enhance content type negotiation or implement protobuf for our types, we have to make
// use of this for our api group negotiated serializers.
func DefaultSubsetNegotiatedSerializer(codecs serializer.CodecFactory) runtime.NegotiatedSerializer {
	return SubsetNegotiatedSerializer(codecs, NoProtobuf)
}
