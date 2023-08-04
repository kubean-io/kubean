package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubean-io/kubean-api/apis"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope="Cluster"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=`.metadata.creationTimestamp`,name="Age",type=date

// Cluster represents the desire state and status of a member cluster.
type Cluster struct {
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
	// +optional
	PreCheckRef *apis.ConfigMapRef `json:"preCheckRef"`
}

func (spec *Spec) ConfigDataList() []*apis.ConfigMapRef {
	return []*apis.ConfigMapRef{spec.HostsConfRef, spec.VarsConfRef, spec.KubeConfRef, spec.PreCheckRef}
}

func (spec *Spec) SecretDataList() []*apis.SecretRef {
	return []*apis.SecretRef{spec.SSHAuthRef}
}

type ClusterConditionType string

const (
	ClusterConditionCreating ClusterConditionType = "Running"

	ClusterConditionRunning ClusterConditionType = "Succeeded"

	ClusterConditionUpdating ClusterConditionType = "Failed"

	BlockedStatus ClusterConditionType = "Blocked"
)

type ClusterCondition struct {
	// ClusterOps refers to the name of ClusterOperation.
	// +required
	ClusterOps string `json:"clusterOps"`
	// +optional
	Status ClusterConditionType `json:"status"`
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// +optional
	EndTime *metav1.Time `json:"endTime,omitempty"`
}

// Status contains information about the current status of a
// cluster updated periodically by cluster controller.
type Status struct {
	Conditions []ClusterCondition `json:"conditions"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterList contains a list of member cluster.
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items holds a list of Cluster.
	Items []Cluster `json:"items"`
}
