/*
Copyright 2025.

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

type PetType string

const (
	DogPetType  PetType = "dog"
	CatPetType  PetType = "cat"
	BirdPetType PetType = "bird"
)

type PetPhase string

const (
	PetPending  PetPhase = "Pending"
	PetReady    PetPhase = "Ready"
	PetDeleting PetPhase = "Deleting"
	PetError    PetPhase = "Error"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PetSpec defines the desired state of Pet.
type PetSpec struct {
	// Name of the pet
	Name string `json:"name"`

	// Type of the pet
	Type PetType `json:"type"`
}

type PetStatus struct {
	// State represents controller state
	Phase      PetPhase           `json:"state,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Pet is the Schema for the pets API.
type Pet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PetSpec   `json:"spec,omitempty"`
	Status PetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PetList contains a list of Pet.
type PetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Pet{}, &PetList{})
}
