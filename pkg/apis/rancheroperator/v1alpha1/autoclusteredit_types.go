package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rancherv3 "github.com/rancher/types/apis/management.cattle.io/v3"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AutoClusterEditSpec defines the desired state of AutoClusterEdit
type AutoClusterEditSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	ClusterSelector string            `json:"clusterSelector"`
	ClusterTemplate rancherv3.Cluster `json:clusterTemplate`
}

// AutoClusterEditStatus defines the observed state of AutoClusterEdit
type AutoClusterEditStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AutoClusterEdit is the Schema for the autoclusteredits API
// +k8s:openapi-gen=true
type AutoClusterEdit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutoClusterEditSpec   `json:"spec,omitempty"`
	Status AutoClusterEditStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AutoClusterEditList contains a list of AutoClusterEdit
type AutoClusterEditList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AutoClusterEdit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AutoClusterEdit{}, &AutoClusterEditList{})
}
