package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Done = "Done"
	Sync = "Sync"
	Install = "Install"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AndroidSDKList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []AndroidSDK `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AndroidSDK struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              AndroidSDKSpec   `json:"spec"`
	Status            AndroidSDKStatus `json:"status,omitempty"`
}

type AndroidSDKSpec struct {
	Data string
}

type AndroidSDKStatus struct {
	Status string
}
