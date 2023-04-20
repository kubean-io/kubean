package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubean.io/api/apis"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope="Cluster"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=`.metadata.creationTimestamp`,name="Age",type=date

// ClusterOperation represents the desire state and status of a member cluster.
type ClusterOperation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec Spec `json:"spec"`

	// +optional
	Status Status `json:"status,omitempty"`
}

type (
	ActionSource string
	ActionType   string
)

const (
	PlaybookActionType ActionType = "playbook"
	ShellActionType    ActionType = "shell"
)

const (
	BuiltinActionSource   ActionSource = "builtin"
	ConfigMapActionSource ActionSource = "configmap"
)

// Spec defines the desired state of a member cluster.
type Spec struct {
	// Cluster the name of Cluster.kubean.io.
	// +required
	Cluster string `json:"cluster"`
	// HostsConfRef will be filled by operator when it performs backup.
	// +optional
	HostsConfRef *apis.ConfigMapRef `json:"hostsConfRef"`
	// VarsConfRef will be filled by operator when it performs backup.
	// +optional
	VarsConfRef *apis.ConfigMapRef `json:"varsConfRef,omitempty"`
	// SSHAuthRef will be filled by operator when it performs backup.
	// +optional
	SSHAuthRef *apis.SecretRef `json:"sshAuthRef,omitempty"`
	// +optional
	// EntrypointSHRef will be filled by operator when it renders entrypoint.sh.
	EntrypointSHRef *apis.ConfigMapRef `json:"entrypointSHRef,omitempty"`
	// +required
	ActionType ActionType `json:"actionType"`
	// +required
	Action string `json:"action"`
	// +optional
	// +kubebuilder:default="builtin"
	ActionSource *ActionSource `json:"actionSource"`
	// +optional
	ActionSourceRef *apis.ConfigMapRef `json:"actionSourceRef,omitempty"`
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
	// +optional
	Resources corev1.ResourceRequirements `json:"resources"`
	// +optional
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty"`
}

type HookAction struct {
	// +required
	ActionType ActionType `json:"actionType"`
	// +required
	Action string `json:"action"`
	// +optional
	// +kubebuilder:default="builtin"
	ActionSource *ActionSource `json:"actionSource"`
	// +optional
	ActionSourceRef *apis.ConfigMapRef `json:"actionSourceRef,omitempty"`
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
	JobRef *apis.JobRef `json:"jobRef,omitempty"`
	// +optional
	Status OpsStatus `json:"status"`
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// +optional
	EndTime *metav1.Time `json:"endTime,omitempty"`
	// Digest is used to avoid the change of clusterOps by others. it will be filled by operator. Do Not change this value.
	// +optional
	Digest string `json:"digest,omitempty"`
	// HasModified indicates the spec has been modified by others after created.
	// +optional
	HasModified bool `json:"hasModified,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterOperationList contains a list of member cluster.
type ClusterOperationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items holds a list of ClusterOperation.
	Items []ClusterOperation `json:"items"`
}
