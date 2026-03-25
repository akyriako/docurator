/*
Copyright 2026.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SpaceSpec defines the desired state of Space.
type SpaceSpec struct {
	Name string `json:"name"`

	// +kubebuilder:validation:Optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="akyriako78/docusaurus-gitea-bootstrap:0.0.1"
	BootstrapImage string `json:"bootstrapImage,omitempty"`

	GiteaSecretRef corev1.SecretReference `json:"giteaSecretRef"`
}

// SpaceStatus defines the observed state of Space.
type SpaceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Ready   *bool  `json:"ready"`
	RepoURL string `json:"repoUrl"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Space is the Schema for the spaces API.
// +kubebuilder:printcolumn:name="Repo URL",type=string,JSONPath=`.status.repoUrl`
// +kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=".metadata.creationTimestamp"
type Space struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SpaceSpec   `json:"spec,omitempty"`
	Status SpaceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SpaceList contains a list of Space.
type SpaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Space `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Space{}, &SpaceList{})
}
