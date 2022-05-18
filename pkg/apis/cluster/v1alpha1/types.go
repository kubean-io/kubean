package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindCluster      = "Cluster"
	ResourcesSingularCluster = "cluster"
	ResourcesPluralCluster   = "clusters"

	HostCluster = "cluster-role.kpanda.io/host"
	// Description of which region the cluster been placed.
	ClusterRegion = "cluster.kpanda.io/region"
	// Name of the cluster group.
	ClusterGroup = "cluster.kpanda.io/group"

	Finalizer = "finalizer.cluster.kpanda.io"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope="Cluster"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=`.status.kubernetesVersion`,name="Version",type=string
// +kubebuilder:printcolumn:JSONPath=`.spec.syncMode`,name="Mode",type=string
// +kubebuilder:printcolumn:JSONPath=`.spec.provider`,name="Provider",type=string
// +kubebuilder:printcolumn:JSONPath=`.status.conditions[?(@.type=="Running")].status`,name="Running",type=string
// +kubebuilder:printcolumn:JSONPath=`.status.kubeSystemId`,name="kubeSystemId",type=string
// +kubebuilder:printcolumn:JSONPath=`.metadata.creationTimestamp`,name="Age",type=date

// Cluster represents the desire state and status of a member cluster.
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec represents the specification of the desired behavior of member cluster.
	Spec ClusterSpec `json:"spec"`

	// Status represents the status of member cluster.
	// +optional
	Status ClusterStatus `json:"status,omitempty"`
}

// ClusterSpec defines the desired state of a member cluster.
type ClusterSpec struct {
	test string `json:"test"`
	// SyncMode describes how a cluster sync resources from kubernetes control plane.
	// +kubebuilder:validation:Enum=Push;Pull
	// +required
	SyncMode ClusterSyncMode `json:"syncMode"`

	// Provider represents the cloud provider name of the member cluster.
	// +required
	Provider1 string `json:"provider1"`

	// The API endpoint of the member cluster. This can be a hostname,
	// hostname:port, IP or IP:port.
	// +optional
	APIEndpoint1 string `json:"apiEndpoint1,omitempty"`

	// SecretRef represents the secret contains mandatory credentials to access the member cluster.
	// The secret should hold credentials as follows:
	// - secret.data.token
	// - secret.data.caBundle
	// +optional
	SecretRef *LocalSecretReference `json:"secretRef,omitempty"`

	// InsecureSkipTLSVerification indicates that the kubernetes control plane should not confirm the validity of the serving
	// certificate of the cluster it is connecting to. This will make the HTTPS connection between the kubernetes control
	// plane and the member cluster insecure.
	// Defaults to false.
	// +optional
	InsecureSkipTLSVerification bool `json:"insecureSkipTlsVerification,omitempty"`

	// Region represents the region of the member cluster locate in.
	// +optional
	Region string `json:"region,omitempty"`

	// Zone represents the zone of the member cluster locate in.
	// +optional
	Zone string `json:"zone,omitempty"`

	// Taints attached to the member cluster.
	// Taints on the cluster have the "effect" on
	// any resource that does not tolerate the Taint.
	// +optional
	Taints []corev1.Taint `json:"taints,omitempty"`
}

// ClusterSyncMode describes the mode of synchronization between member cluster and kubernetes control plane.
type ClusterSyncMode string

const (
	// Push means that the controller on the kubernetes control plane will in charge of synchronization.
	// The controller watches resources change on kubernetes control plane then pushes them to member cluster.
	Push ClusterSyncMode = "Push"

	// Pull means that the controller running on the member cluster will in charge of synchronization.
	// The controller, as well known as 'agent', watches resources change on kubernetes control plane then fetches them
	// and applies locally on the member cluster.
	Pull ClusterSyncMode = "Pull"
)

// LocalSecretReference is a reference to a secret within the enclosing
// namespace.
type LocalSecretReference struct {
	// Namespace is the namespace for the resource being referenced.
	Namespace string `json:"namespace"`

	// Name is the name of resource being referenced.
	Name string `json:"name"`

	// ResourceVersion is the version of resource being referenced.
	// +optional
	ResourceVersion string `json:"resourceVersion"`
}

type ClusterConditionType string

// Define valid conditions of a member cluster.
const (
	// ClusterConditionCreating means the cluster is creating.
	ClusterConditionCreating ClusterConditionType = "Creating"

	// ClusterConditionRunning means the cluster is healthy and ready to accept workloads.
	ClusterConditionRunning ClusterConditionType = "Running"

	// ClusterConditionUpdating means the cluster is updating.
	ClusterConditionUpdating ClusterConditionType = "Updating"

	// ClusterConditionDeleting means the cluster is deleting.
	ClusterConditionDeleting ClusterConditionType = "Deleting"

	// ClusterConditionUnknown means the cluster is error status.
	ClusterConditionUnknown ClusterConditionType = "Unknown"
)

func (condition ClusterConditionType) String() string {
	return string(condition)
}

var ClusterConditions = []string{ClusterConditionCreating.String(), ClusterConditionRunning.String(), ClusterConditionUpdating.String(), ClusterConditionDeleting.String(), ClusterConditionUnknown.String()}

// ClusterCondition contains condition information for a cluster.
type ClusterCondition struct {
	// Type of node condition.
	Type ClusterConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=NodeConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status metav1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// Last time we got an update on a given condition.
	// +optional
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime,omitempty" protobuf:"bytes,3,opt,name=lastHeartbeatTime"`
	// Last time the condition transit from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// (brief) reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

// ClusterStatus contains information about the current status of a
// cluster updated periodically by cluster controller.
type ClusterStatus struct {
	// KubernetesVersion represents version of the member cluster.
	// +optional
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`

	// KubeSystemId represents the uuid of sub cluster kube-system namespace.
	// +optional
	KubeSystemID string `json:"kubeSystemId,omitempty"`

	// ProxyURL is the proxy URL for the cluster.
	// If not empty, the kubernetes control plane will use this proxy to talk to the cluster.
	// More details please refer to: https://github.com/kubernetes/client-go/issues/351
	// +optional
	ProxyURL string `json:"proxyUrl,omitempty"`

	// ClusterProxyMode represents the proxy mode of the member cluster
	// +optional
	ProxyMode string `json:"proxyMode,omitempty"`

	// ServiceCIDR represents the service's sub net of the member cluster
	// +optional
	ServiceCIDR string `json:"serviceCIDR,omitempty"`

	// PodCIDR represents the pod's sub net of the member cluster
	// +optional
	PodCIDR string `json:"podCIDR,omitempty"`

	// Conditions is an array of current cluster conditions.
	// +optional
	Conditions []ClusterCondition `json:"conditions,omitempty"`

	// NodeSummary represents the summary of nodes status in the member cluster.
	// +optional
	NodeSummary *NodeSummary `json:"nodeSummary,omitempty"`

	// ResourceSummary represents the summary of resources in the member cluster.
	// +optional
	ResourceSummary *ResourceSummary `json:"resourceSummary,omitempty"`

	// +optional
	DeploymentSummary *WorkloadSummary `json:"deploymentSummary,omitempty"`

	// +optional
	StatefulSetSummary *WorkloadSummary `json:"statefulSetSummary,omitempty"`

	// +optional
	DaemonSetSummary *WorkloadSummary `json:"daemonSetSummary,omitempty"`

	// +optional
	PodSetSummary *WorkloadSummary `json:"podSetSummary,omitempty"`
}

// APIResource specifies the name and kind names for the resource.
type APIResource struct {
	// Name is the plural name of the resource.
	// +required
	Name string `json:"name"`
	// Kind is the kind for the resource (e.g. 'Deployment' is the kind for resource 'deployments')
	// +required
	Kind string `json:"kind"`
}

// NodeSummary represents the summary of nodes status in a specific cluster.
type NodeSummary struct {
	// TotalNum is the total number of nodes in the cluster.
	// +optional
	TotalNum int32 `json:"totalNum,omitempty"`
	// ReadyNum is the number of ready nodes in the cluster.
	// +optional
	ReadyNum int32 `json:"readyNum,omitempty"`
}

// WorkloadSummary represents the summary of workload status in a specific cluster.
type WorkloadSummary struct {
	// TotalNum is the total number of workloads in the cluster.
	// +optional
	TotalNum int32 `json:"totalNum,omitempty"`
	// ReadyNum is the number of ready workloads in the cluster.
	// +optional
	ReadyNum int32 `json:"readyNum,omitempty"`
}

// ResourceSummary represents the summary of resources in the member cluster.
type ResourceSummary struct {
	// Allocatable represents the resources of a cluster that are available for scheduling.
	// Total amount of allocatable resources on all nodes.
	// +optional
	Allocatable corev1.ResourceList `json:"allocatable,omitempty"`
	// Allocating represents the resources of a cluster that are pending for scheduling.
	// Total amount of required resources of all Pods that are waiting for scheduling.
	// +optional
	Allocating corev1.ResourceList `json:"allocating,omitempty"`
	// Allocated represents the resources of a cluster that have been scheduled.
	// Total amount of required resources of all Pods that have been scheduled to nodes.
	// +optional
	Allocated corev1.ResourceList `json:"allocated,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterList contains a list of member cluster.
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items holds a list of Cluster.
	Items []Cluster `json:"items"`
}
