package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rancherv3 "github.com/rancher/types/apis/management.cattle.io/v3"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AutoProjectSpec defines the desired state of AutoProject
type AutoProjectSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

	ProjectSpec rancherv3.ProjectSpec `json:"projectSpec`
}

// AutoProjectStatus defines the observed state of AutoProject
type AutoProjectStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AutoProject is the Schema for the autoprojects API
// +k8s:openapi-gen=true
type AutoProject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutoProjectSpec   `json:"spec,omitempty"`
	Status AutoProjectStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AutoProjectList contains a list of AutoProject
type AutoProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AutoProject `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AutoProject{}, &AutoProjectList{})
}
