/*
Copyright 2023.

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

// http-go-operator/api/v1/kindcustomhttp_types.go
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KindCustomHttpSpec defines the desired state of KindCustomHttp
type KindCustomHttpSpec struct {
	ReplicaCount  int32             `json:"replicaCount"`
	Port          int32             `json:"port"`
	ConfigMapData map[string]string `json:"configMapData,omitempty"`
}

// KindCustomHttpStatus defines the observed state of KindCustomHttp
type KindCustomHttpStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KindCustomHttp is the Schema for the kindcustomhttps API
type KindCustomHttp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KindCustomHttpSpec   `json:"spec,omitempty"`
	Status KindCustomHttpStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KindCustomHttpList contains a list of KindCustomHttp
type KindCustomHttpList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KindCustomHttp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KindCustomHttp{}, &KindCustomHttpList{})
}
