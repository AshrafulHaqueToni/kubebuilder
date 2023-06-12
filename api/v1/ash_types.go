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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ContainerSpec struct {
	Image string `json:"image"`
	Port  int32  `json:"port"`
}

type ServiceSpec struct {
	//+optional
	ServiceName string `json:"serviceName,omitempty"`
	ServiceType string `json:"serviceType"`
	//+optional
	ServiceNodePort int32 `json:"serviceNodePort,omitempty"`
	ServicePort     int32 `json:"servicePort"`
}

// AshSpec defines the desired state of Ash
type AshSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	//+optional
	DeploymentName string `json:"deploymentName,omitempty"`
	//Replicas defines number of pods will be running in the deployment
	Replicas  *int32        `json:"replicas"`
	Container ContainerSpec `json:"container"`
	// Service contains ServiceName, ServiceType, ServiceNodePort
	//+optional
	Service ServiceSpec `json:"service,omitempty"`
}

// AshStatus defines the observed state of Ash
type AshStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Ash is the Schema for the ashes API
type Ash struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AshSpec   `json:"spec,omitempty"`
	Status AshStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AshList contains a list of Ash
type AshList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ash `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ash{}, &AshList{})
}
