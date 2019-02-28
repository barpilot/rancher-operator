package v3

import (
	"github.com/rancher/norman/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterUserAttribute struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Groups       []string `json:"groups,omitempty"`
	LastRefresh  string   `json:"lastRefresh,omitempty"`
	NeedsRefresh bool     `json:"needsRefresh"`
	Enabled      bool     `json:"enabled"`
}

type ClusterAuthToken struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	UserName      string `json:"userName"`
	ExpiresAt     string `json:"expiresAt,omitempty"`
	SecretKeyHash string `json:"hash"`
	Enabled       bool   `json:"enabled"`
}
