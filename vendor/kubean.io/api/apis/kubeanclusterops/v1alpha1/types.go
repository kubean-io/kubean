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

// KuBeanClusterOps represents the desire state and status of a member cluster.
type KuBeanClusterOps struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec Spec `json:"spec"`

	// +optional
	Status Status `json:"status,omitempty"`
}

type ActionType string

const (
	PlaybookActionType ActionType = "playbook"
	ShellActionType    ActionType = "shell"
)

// Spec defines the desired state of a member cluster.
type Spec struct {
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
	// +optional
	ExtraArgs string `json:"extraArgs"`
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
	// +required
	ActionType ActionType `json:"actionType"`
	// +required
	Action string `json:"action"`
	// +optional
	ExtraArgs string `json:"extraArgs"`
}

type OpsStatus string

const (
	RunningStatus   OpsStatus = "Running"
	SucceededStatus OpsStatus = "Succeeded"
	FailedStatus    OpsStatus = "Failed"
	BlockedStatus   OpsStatus = "Blocked"
)

// Status contains information about the current status of a
// cluster operation job updated periodically by cluster controller.
type Status struct {
	// +optional
	Action string `json:"action"`
	// +optional
	JobRef *apis.JobRef `json:"jobRef"`
	// +optional
	Status OpsStatus `json:"status"`
	// +optional
	StartTime *metav1.Time `json:"startTime"`
	// +optional
	EndTime *metav1.Time `json:"endTime"`
	// Digest is used to avoid the change of clusterOps by others. it will be filled by operator. Do Not change this value.
	// +optional
	Digest string `json:"digest,omitempty"`
	// HasModified indicates the spec has been modified by others after created.
	// +optional
	HasModified bool `json:"hasModified,omitempty"`
	// KubeConfig will be modified by the job completed successfully and operator fetch the kubeconfig of the new k8s cluster.
	// +optional
	KubeConfig string `json:"kubeConfig,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KuBeanClusterOpsList contains a list of member cluster.
type KuBeanClusterOpsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items holds a list of KuBeanClusterOps.
	Items []KuBeanClusterOps `json:"items"`
}
