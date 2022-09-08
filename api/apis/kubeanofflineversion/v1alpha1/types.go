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

type KuBeanOfflineVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec Spec `json:"spec"`
}

type Spec struct {
	// Arch for x86_64  aarch64... , represent for the arch of this offline package
	// +required
	Arch []string `json:"arch,omitempty""`

	// Kubespray , the tag of kubespray
	// +required
	Kubespray string `json:"kubespray,omitempty"`

	// Items cni containerd kubeadm kube etcd cilium calico
	// +required
	Items []*SoftwareInfo `json:"items"`

	// +optional
	Docker []*DockerInfo `json:"docker"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type KuBeanOfflineVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items holds a list of KuBeanClusterOps.
	Items []KuBeanOfflineVersion `json:"items"`
}

type SoftwareInfo struct {
	Name string `json:"name"`
	// +optional
	VersionRange []string `json:"versionRange,omitempty"`
}

type DockerInfo struct {
	OS string `json:"os"`
	// +optional
	VersionRange []string `json:"versionRange,omitempty"`
}
