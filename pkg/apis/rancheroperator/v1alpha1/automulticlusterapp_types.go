package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AutoMultiClusterAppSpec defines the desired state of AutoMultiClusterApp
type AutoMultiClusterAppSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

	MultiClusterApp string `json:"multiClusterApp"`
	ProjectSelector string `json:"projectSelector"`
}

// AutoMultiClusterAppStatus defines the observed state of AutoMultiClusterApp
type AutoMultiClusterAppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AutoMultiClusterApp is the Schema for the automulticlusterapps API
// +k8s:openapi-gen=true
type AutoMultiClusterApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutoMultiClusterAppSpec   `json:"spec,omitempty"`
	Status AutoMultiClusterAppStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AutoMultiClusterAppList contains a list of AutoMultiClusterApp
type AutoMultiClusterAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AutoMultiClusterApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AutoMultiClusterApp{}, &AutoMultiClusterAppList{})
}
