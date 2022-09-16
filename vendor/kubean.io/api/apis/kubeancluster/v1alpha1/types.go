package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubean.io/api/apis"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope="Cluster"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=`.metadata.creationTimestamp`,name="Age",type=date

// KuBeanCluster represents the desire state and status of a member cluster.
type KuBeanCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec Spec `json:"spec"`

	// +optional
	Status Status `json:"status,omitempty"`
}

// Spec defines the desired state of a member cluster.
type Spec struct {
	// HostsConfRef stores hosts.yml.
	// +required
	HostsConfRef *apis.ConfigMapRef `json:"hostsConfRef"`
	// VarsConfRef stores group_vars.yml.
	// +required
	VarsConfRef *apis.ConfigMapRef `json:"varsConfRef"`
	// KubeConfRef stores cluster kubeconfig.
	// +optional
	KubeConfRef *apis.ConfigMapRef `json:"kubeconfRef"`
	// SSHAuthRef stores ssh key and if it is empty ,then use sshpass.
	// +optional
	SSHAuthRef *apis.SecretRef `json:"sshAuthRef"`
}

type ClusterConditionType string

const (
	ClusterConditionCreating ClusterConditionType = "Running"

	ClusterConditionRunning ClusterConditionType = "Succeeded"

	ClusterConditionUpdating ClusterConditionType = "Failed"

	BlockedStatus ClusterConditionType = "Blocked"
)

type ClusterCondition struct {
	// ClusterOps refers to the name of KuBeanClusterOps.
	// +required
	ClusterOps string `json:"clusterOps"`
	// +optional
	Status ClusterConditionType `json:"status"`
	// +optional
	StartTime *metav1.Time `json:"startTime"`
	// +optional
	EndTime *metav1.Time `json:"endTime"`
}

// Status contains information about the current status of a
// cluster updated periodically by cluster controller.
type Status struct {
	Conditions []ClusterCondition `json:"conditions"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KuBeanClusterList contains a list of member cluster.
type KuBeanClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items holds a list of KuBeanCluster.
	Items []KuBeanCluster `json:"items"`
}
