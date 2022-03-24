/*
Copyright 2022 Lance Yuan.

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

package v1

import (
	networkingv1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StaticSpec defines the desired state of Static
type StaticSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Static. Edit static_types.go to remove/update

	IngressClassName *string  `json:"ingressClassName"`
	Path             string   `json:"path"`
	Content          string   `json:"content"`
	Hosts            []string `json:"hosts"`
}

// StaticStatus defines the observed state of Static
type StaticStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	networkingv1.IngressStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Static is the Schema for the statics API
//+kubebuilder:printcolumn:name="Host",type=string,JSONPath=`.spec.hosts`
//+kubebuilder:printcolumn:name="Path",type=string,JSONPath=`.spec.path`
//+kubebuilder:printcolumn:name="Content",type=string,JSONPath=`.spec.content`
type Static struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StaticSpec   `json:"spec,omitempty"`
	Status StaticStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StaticList contains a list of Static
type StaticList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Static `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Static{}, &StaticList{})
}
