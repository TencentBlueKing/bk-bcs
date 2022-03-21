/*
Copyright 2022.

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

// ArgocdProjectSpec defines the desired state of ArgocdProject
type ArgocdProjectSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ArgocdProject. Edit argocdproject_types.go to remove/update
	Foo string `json:"foo,omitempty" protobuf:"bytes,1,opt,name=foo"`
}

// ArgocdProjectStatus defines the observed state of ArgocdProject
type ArgocdProjectStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ArgocdProject is the Schema for the argocdprojects API
type ArgocdProject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   ArgocdProjectSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status ArgocdProjectStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

//+kubebuilder:object:root=true

// ArgocdProjectList contains a list of ArgocdProject
type ArgocdProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []ArgocdProject `json:"items" protobuf:"bytes,2,rep,name=items"`
}
