/*
 * Copyright (c) 2021 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ipamrange

import (
	"fmt"
	"github.com/onmetal/onmetal-api/pkg/ipam"
)

type RequestSpec struct {
	Request string
	Spec    ipam.RequestSpec
	Error   string
}

func (r *RequestSpec) String() string {
	return r.Request
}

func (r *RequestSpec) IsValid() bool {
	return r.Spec != nil
}

type RequestSpecList []*RequestSpec

func (r *RequestSpecList) Add(element ...*RequestSpec) {
	(*r) = append(*r, element...)
}

func (r RequestSpecList) PendingSpecs(list AllocationStatusList) RequestSpecList {
	newList := list.Copy()
	var result RequestSpecList
	for _, spec := range r {
		if spec.IsValid() {
			if index := list.LookUp(spec.Request); index >= 0 {
				newList = append(newList[0:index], newList[index+1:]...)
				if list[index].CIDR == nil {
					result = append(result, spec)
				}
			} else {
				result = append(result, spec)
			}
		}
	}
	return result
}

func (r RequestSpecList) ValidSpecs() RequestSpecList {
	var specs RequestSpecList
	for _, s := range r {
		if s.Spec != nil {
			specs = append(specs, s)
		}
	}
	return specs
}

func (r RequestSpecList) InValidSpecs() RequestSpecList {
	var specs RequestSpecList
	for _, s := range r {
		if s.Spec == nil {
			specs = append(specs, s)
		}
	}
	return specs
}

func (r RequestSpecList) String() string {
	s := "["
	sep := ""
	for _, spec := range r {
		s = fmt.Sprintf("%s%s%s", s, sep, spec)
		sep = ","
	}
	return s + "]"
}

func (r RequestSpecList) Error() error {
	msg := ""
	for i, s := range r {
		if s.Error != "" {
			msg = fmt.Sprintf("%s, [%d]%s: %s", msg, i, s.Request, s.Error)
		}
	}
	if msg == "" {
		return fmt.Errorf(msg)
	}
	return fmt.Errorf(msg[2:])
}

func ParseRequestSpec(request string) *RequestSpec {
	spec, err := ipam.ParseRequestSpec(request)
	return &RequestSpec{
		Request: request,
		Spec:    spec,
		Error:   OptionalErrorToString(err),
	}
}

func OptionalErrorToString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
