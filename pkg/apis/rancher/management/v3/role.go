package v3

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GlobalRole is a ClusterRole approximation used within rancher ecosystem
// +kubebuilder:object:root=true

type GlobalRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Rules []rbacv1.PolicyRule `json:"rules,omitempty"`
}

// GlobalRoleList contains a list of GlobalRoles.
// +kubebuilder:object:root=true

type GlobalRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []GlobalRole `json:"items"`
}

// GlobalRoleBinding specifies a global role binging to the user.
// +kubebuilder:object:root=true

type GlobalRoleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	UserName       string `json:"userName,omitempty"`
	GlobalRoleName string `json:"globalRoleName,omitempty"`
}

// GlobalRoleBindingList contains a list of GlobalRoleBingins.
// +kubebuilder:object:root=true

type GlobalRoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []GlobalRoleBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GlobalRole{}, &GlobalRoleBinding{}, &GlobalRoleList{}, &GlobalRoleBindingList{})
}
