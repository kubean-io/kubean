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

type Manifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec Spec `json:"spec"`

	// +optional
	Status Status `json:"status,omitempty"`
}

type Spec struct {
	// +optional
	LocalService LocalService `json:"localService,omitempty"`

	// +required
	KubesprayVersion string `json:"kubesprayVersion,omitempty"`

	// KubeanVersion , the tag of kubean-io
	// +required
	KubeanVersion string `json:"kubeanVersion,omitempty"`

	// +optional
	Components []*SoftwareInfo `json:"components,omitempty"`

	// +optional
	Docker []*DockerInfo `json:"docker,omitempty"`
}

type LocalService struct {
	// +optional
	ImageRepo map[ImageRepoType]string `json:"imageRepo" yaml:"imageRepo"`
	// +optional
	FilesRepo string `json:"filesRepo,omitempty" yaml:"filesRepo,omitempty"`
	// +optional
	YumRepos map[string][]string `json:"yumRepos,omitempty" yaml:"yumRepos,omitempty"`
	// +optional
	HostsMap []*HostsMap `json:"hostsMap,omitempty" yaml:"hostsMap,omitempty"`
}

func (localService *LocalService) GetGHCRImageRepo() string {
	if localService.ImageRepo == nil {
		return ""
	}
	return localService.ImageRepo[GithubImageRepo]
}

type ImageRepoType string

const KubeImageRepo ImageRepoType = "kubeImageRepo"
const GCRImageRepo ImageRepoType = "gcrImageRepo"
const GithubImageRepo ImageRepoType = "githubImageRepo"
const DockerImageRepo ImageRepoType = "dockerImageRepo"
const QuayImageRepo ImageRepoType = "quayImageRepo"

type HostsMap struct {
	// +required
	Domain string `json:"domain,omitempty" yaml:"domain,omitempty"`
	// +required
	Address string `json:"address,omitempty" yaml:"address,omitempty"`
}

type Status struct {
	// +optional
	LocalAvailable LocalAvailable `json:"localAvailable,omitempty"`
}

type LocalAvailable struct {
	// +optional
	KubesprayImage string `json:"kubesprayImage,omitempty"`

	// +optional
	Components []*SoftwareInfoStatus `json:"components,omitempty"`

	// +optional
	Docker []*DockerInfoStatus `json:"docker,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ManifestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items holds a list of KubeanClusterConfig.
	Items []Manifest `json:"items"`
}

type SoftwareInfo struct {
	Name string `json:"name"`
	// +optional
	DefaultVersion string `json:"defaultVersion,omitempty"`
	// +optional
	VersionRange []string `json:"versionRange,omitempty"`
}

type SoftwareInfoStatus struct {
	Name string `json:"name"`
	// +optional
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
	OS string `json:"os"`
	// +optional
	DefaultVersion string `json:"defaultVersion,omitempty"`
	// +optional
	VersionRange []string `json:"versionRange,omitempty"`
}

type DockerInfoStatus struct {
	OS string `json:"os"`
	// +optional
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

func (status *LocalAvailable) MergeSoftwareInfo(name string, versionRange []string) bool {
	var targetNameItem *SoftwareInfoStatus
	for i := range status.Components {
		if status.Components[i].Name == name {
			targetNameItem = status.Components[i]
			break
		}
	}
	if targetNameItem == nil {
		targetNameItem = &SoftwareInfoStatus{Name: name, VersionRange: versionRange}
		status.Components = append(status.Components, targetNameItem)
		return true
	}
	return targetNameItem.Merge(versionRange)
}

func (status *LocalAvailable) MergeDockerInfo(osName string, versionRange []string) bool {
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
