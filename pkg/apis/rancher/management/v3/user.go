package v3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty" norman:"writeOnly,noupdate"`
}

// UserList contains a list of Users.
// +kubebuilder:object:root=true

type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []User `json:"items"`
}

func init() {
	SchemeBuilder.Register(&User{}, &UserList{})
}
