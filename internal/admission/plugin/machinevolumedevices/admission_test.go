// Copyright 2022 IronCore authors
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

package machinevolumedevices_test

import (
	"context"

	. "github.com/ironcore-dev/ironcore/internal/admission/plugin/machinevolumedevices"
	"github.com/ironcore-dev/ironcore/internal/apis/compute"
	"github.com/ironcore-dev/ironcore/internal/apis/storage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/admission"
)

var _ = Describe("Admission", func() {
	var (
		plugin *MachineVolumeDevices
	)
	BeforeEach(func() {
		plugin = NewMachineVolumeDevices()
	})

	It("should ignore non-machine objects", func() {
		volume := &storage.Volume{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "foo",
				Name:      "bar",
			},
		}
		origVolume := volume.DeepCopy()
		Expect(plugin.Admit(
			context.TODO(),
			admission.NewAttributesRecord(
				volume,
				nil,
				storage.Kind("Volume").WithVersion("version"),
				volume.Namespace,
				volume.Name,
				storage.Resource("volumes").WithVersion("version"),
				"",
				admission.Create,
				&metav1.CreateOptions{},
				false,
				nil,
			),
			nil,
		)).NotTo(HaveOccurred())
		Expect(volume).To(Equal(origVolume))
	})

	It("should add volume device names when unset", func() {
		machine := &compute.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "foo",
				Name:      "bar",
			},
			Spec: compute.MachineSpec{
				Volumes: []compute.Volume{
					{
						Device: "odb",
					},
					{},
					{},
				},
			},
		}
		Expect(plugin.Admit(
			context.TODO(),
			admission.NewAttributesRecord(
				machine,
				nil,
				compute.Kind("Machine").WithVersion("version"),
				machine.Namespace,
				machine.Name,
				compute.Resource("machines").WithVersion("version"),
				"",
				admission.Create,
				&metav1.CreateOptions{},
				false,
				nil,
			),
			nil,
		)).NotTo(HaveOccurred())

		Expect(machine.Spec.Volumes).To(Equal([]compute.Volume{
			{
				Device: "odb",
			},
			{
				Device: "oda",
			},
			{
				Device: "odc",
			},
		}))
	})
})
