package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/strings/slices"
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
	// +optional
	Offline OfflineStatus `json:"offline,omitempty"`
}

type OfflineStatus struct {
	// +optional
	Items []*SoftwareInfoStatus `json:"items"`

	// +optional
	Docker []*DockerInfoStatus `json:"docker"`
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

type SoftwareInfoStatus struct {
	Name         string   `json:"name"`
	VersionRange []string `json:"versionRange,omitempty"`
}

func (status *SoftwareInfoStatus) Merge(versionRange []string) bool {
	updated := false
	for i := range versionRange {
		if !slices.Contains(status.VersionRange, versionRange[i]) {
			updated = true
			status.VersionRange = append(status.VersionRange, versionRange[i])
		}
	}
	return updated
}

type DockerInfo struct {
	OS             string   `json:"os"`
	DefaultVersion string   `json:"defaultVersion,omitempty"`
	VersionRange   []string `json:"versionRange,omitempty"`
}

type DockerInfoStatus struct {
	OS           string   `json:"os"`
	VersionRange []string `json:"versionRange,omitempty"`
}

func (info *DockerInfoStatus) Merge(versionRange []string) bool {
	updated := false
	for i := range versionRange {
		if !slices.Contains(info.VersionRange, versionRange[i]) {
			updated = true
			info.VersionRange = append(info.VersionRange, versionRange[i])
		}
	}
	return updated
}

func (status *OfflineStatus) MergeSoftwareInfo(name string, versionRange []string) bool {
	var targetNameItem *SoftwareInfoStatus
	for i := range status.Items {
		if status.Items[i].Name == name {
			targetNameItem = status.Items[i]
			break
		}
	}
	if targetNameItem == nil {
		targetNameItem = &SoftwareInfoStatus{Name: name, VersionRange: versionRange}
		status.Items = append(status.Items, targetNameItem)
		return true
	}
	return targetNameItem.Merge(versionRange)
}

func (status *OfflineStatus) MergeDockerInfo(osName string, versionRange []string) bool {
	var targetNameDockerInfo *DockerInfoStatus
	for i := range status.Docker {
		if status.Docker[i].OS == osName {
			targetNameDockerInfo = status.Docker[i]
			break
		}
	}
	if targetNameDockerInfo == nil {
		targetNameDockerInfo = &DockerInfoStatus{OS: osName, VersionRange: versionRange}
		status.Docker = append(status.Docker, targetNameDockerInfo)
		return true
	}
	return targetNameDockerInfo.Merge(versionRange)
}
