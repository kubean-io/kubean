package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope="Cluster"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=`.metadata.creationTimestamp`,name="Age",type=date

type KuBeanComponentsVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec Spec `json:"spec"`

	// +optional
	Status Status `json:"status,omitempty"`
}

type Spec struct {
	// Kubespray , the tag of kubespray
	// +required
	Kubespray string `json:"kubespray,omitempty"`

	// Kubean , the tag of kubean-io
	// +optional
	Kubean string `json:"kubean,omitempty"`

	// +required
	Items []*SoftwareInfo `json:"items,omitempty"`

	// +optional
	Docker []*DockerInfo `json:"docker"`
}

type Status struct {
	Offline OfflineStatus `json:"offline,omitempty"`
}

type OfflineStatus struct {
	// +required
	Items []*SoftwareInfo `json:"items"`

	// +optional
	Docker []*DockerInfo `json:"docker"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type KuBeanComponentsVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items holds a list of KuBeanClusterOps.
	Items []KuBeanComponentsVersion `json:"items"`
}

type SoftwareInfo struct {
	Name           string   `json:"name"`
	DefaultVersion string   `json:"defaultVersion,omitempty"`
	VersionRange   []string `json:"versionRange,omitempty"`
}

type DockerInfo struct {
	OS              string   `json:"os"`
	VersionRange    []string `json:"versionRange,omitempty"`
	ContainerdRange []string `json:"containerdRange,omitempty"`
}
