package v1alpha1

import (
	"github.com/daocloud/kubean/pkg/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	Spec ClusterSpec `json:"spec"`

	// +optional
	Status ClusterStatus `json:"status,omitempty"`
}

// ClusterSpec defines the desired state of a member cluster.
type ClusterSpec struct {
	// +required
	HostsConfRef *apis.ConfigMapRef `json:"hostsConfRef"`
	// +required
	VarsConfRef *apis.ConfigMapRef `json:"varsConfRef"`
	// +required
	SSHAuthRef *apis.SecretRef `json:"sshAuthRef"`
}

type ClusterConditionType string

const (
	ClusterConditionCreating ClusterConditionType = "Running"

	ClusterConditionRunning ClusterConditionType = "Succeeded"

	ClusterConditionUpdating ClusterConditionType = "Failed"
)

type ClusterCondition struct {
	// +required
	ClusterOps string `json:"clusterOps"`
	// +required
	Status ClusterConditionType `json:"status"`
	// +optional
	StartTime *metav1.Time `json:"startTime"`
	// +optional
	EndTime *metav1.Time `json:"endTime"`
}

// ClusterStatus contains information about the current status of a
// cluster updated periodically by cluster controller.
type ClusterStatus struct {
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
