package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/daocloud/kubean/pkg/apis"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope="Cluster"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=`.metadata.creationTimestamp`,name="Age",type=date

// KuBeanClusterOps represents the desire state and status of a member cluster.
type KuBeanClusterOps struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec ClusterSpec `json:"spec"`

	// +optional
	Status ClusterStatus `json:"status,omitempty"`
}

type ActionType string

const (
	PlaybookActionType ActionType = "playbook"
	ShellActionType    ActionType = "shell"
)

// ClusterSpec defines the desired state of a member cluster.
type ClusterSpec struct {
	// KuBeanCluster the name of KuBeanCluster.
	// +required
	KuBeanCluster string `json:"kuBeanCluster"`
	// HostsConfRef will be filled by operator when it performs backup.
	// +optional
	HostsConfRef *apis.ConfigMapRef `json:"hostsConfRef"`
	// VarsConfRef will be filled by operator when it performs backup.
	// +optional
	VarsConfRef *apis.ConfigMapRef `json:"varsConfRef"`
	// SSHAuthRef will be filled by operator when it performs backup.
	// +optional
	SSHAuthRef *apis.SecretRef `json:"sshAuthRef"`
	// +optional
	// EntrypointSHRef will be filled by operator when it renders entrypoint.sh.
	EntrypointSHRef *apis.ConfigMapRef `json:"entrypointSHRef"`
	// +required
	ActionType ActionType `json:"actionType"`
	// +required
	Action string `json:"action"`
	// +required
	BackoffLimit int `json:"backoffLimit"`
	// +required
	Image string `json:"image"`
	// +optional
	PreHook []HookAction `json:"preHook"`
	// +optional
	PostHook []HookAction `json:"postHook"`
}

type HookAction struct {
	ActionType ActionType `json:"actionType"` // todo 小写 修改example里的yaml
	Action     string     `json:"action"`
}

type ClusterOpsStatus string

const (
	RunningStatus   ClusterOpsStatus = "Running"
	SucceededStatus ClusterOpsStatus = "Succeeded"
	FailedStatus    ClusterOpsStatus = "Failed"
)

// ClusterStatus contains information about the current status of a
// cluster updated periodically by cluster controller.
type ClusterStatus struct {
	// +optional
	Action string `json:"action"`
	// +optional
	JobRef *apis.JobRef `json:"jobRef"`
	// +optional
	PodRef *apis.PodRef `json:"podRef"`
	// +optional
	Status ClusterOpsStatus `json:"status"`
	// +optional
	StartTime *metav1.Time `json:"startTime"`
	// +optional
	EndTime *metav1.Time `json:"endTime"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KuBeanClusterOpsList contains a list of member cluster.
type KuBeanClusterOpsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items holds a list of KuBeanClusterOps.
	Items []KuBeanClusterOps `json:"items"`
}
