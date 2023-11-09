package v3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Setting is a rancher setting value
// +kubebuilder:object:root=true

type Setting struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Value string `json:"value"`
}

// SettingList contains a list of Settings.
// +kubebuilder:object:root=true

type SettingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Setting `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Setting{}, &SettingList{})
}
