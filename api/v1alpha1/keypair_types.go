/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KeyPairSpec defines the desired state of KeyPair
type KeyPairSpec struct {
	// Foo is an example field of KeyPair. Edit KeyPair_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// KeyPairStatus defines the observed state of KeyPair
type KeyPairStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// KeyPair is the Schema for the keypairs API
type KeyPair struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   KeyPairSpec   `json:"spec,omitempty"`
	Status KeyPairStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KeyPairList contains a list of KeyPair
type KeyPairList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KeyPair `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KeyPair{}, &KeyPairList{})
}
