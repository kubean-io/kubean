package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope="KuBeanCluster"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=`.metadata.creationTimestamp`,name="Age",type=date

// KuBeanCluster represents the desire state and status of a member cluster.
type KuBeanCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec represents the specification of the desired behavior of member cluster.
	Spec ClusterSpec `json:"spec"`

	// Status represents the status of member cluster.
	// +optional
	Status ClusterStatus `json:"status,omitempty"`
}

type ConfigMapRef struct {
	NameSpace string `json:"nameSpace"`
	Name      string `json:"name"`
}

type SecretRef struct {
	NameSpace string `json:"nameSpace"`
	Name      string `json:"name"`
}

// ClusterSpec defines the desired state of a member cluster.
type ClusterSpec struct {
	HostsConfRef *ConfigMapRef `json:"hostsConfRef"`
	VarsConfRef  *ConfigMapRef `json:"varsConfRef"`
	SSHAuthRef   *SecretRef    `json:"sshAuthRef"`
}

// ClusterStatus contains information about the current status of a
// cluster updated periodically by cluster controller.
type ClusterStatus struct {
	testResult string `json:"testResult"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KuBeanClusterList contains a list of member cluster.
type KuBeanClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items holds a list of KuBeanCluster.
	Items []KuBeanCluster `json:"items"`
}
