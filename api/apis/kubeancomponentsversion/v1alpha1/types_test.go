package v1alpha1

import (
	"testing"
)

func TestMerge(t *testing.T) {
	componentsVersion := &KuBeanComponentsVersion{}
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "add software air-gap etcd1 item",
			args: func() bool {
				updated := componentsVersion.Status.Offline.MergeSoftwareInfo("etcd1", []string{"1.1", "1.2"})
				return updated && len(componentsVersion.Status.Offline.Items) == 1
			},
			want: true,
		},
		{
			name: "update software air-gap etcd1 item",
			args: func() bool {
				updated := componentsVersion.Status.Offline.MergeSoftwareInfo("etcd1", []string{"1.2", "1.3"})
				return updated && len(componentsVersion.Status.Offline.Items) == 1 && componentsVersion.Status.Offline.Items[0].Name == "etcd1" && len(componentsVersion.Status.Offline.Items[0].VersionRange) == 3
			},
			want: true,
		},
		{
			name: "update software air-gap etcd1 item but nothing changed",
			args: func() bool {
				updated := componentsVersion.Status.Offline.MergeSoftwareInfo("etcd1", []string{"1.2", "1.3"})
				return !updated && len(componentsVersion.Status.Offline.Items) == 1 && componentsVersion.Status.Offline.Items[0].Name == "etcd1" && len(componentsVersion.Status.Offline.Items[0].VersionRange) == 3
			},
			want: true,
		},
		{
			name: "update software air-gap etcd2 item",
			args: func() bool {
				updated := componentsVersion.Status.Offline.MergeSoftwareInfo("etcd2", []string{"2.1", "2.2"})
				return updated && len(componentsVersion.Status.Offline.Items) == 2 && componentsVersion.Status.Offline.Items[1].Name == "etcd2" && len(componentsVersion.Status.Offline.Items[1].VersionRange) == 2
			},
			want: true,
		},
		{
			name: "update software air-gap etcd2 item but nothing changed",
			args: func() bool {
				updated := componentsVersion.Status.Offline.MergeSoftwareInfo("etcd2", []string{"2.1", "2.2"})
				return !updated && len(componentsVersion.Status.Offline.Items) == 2 && componentsVersion.Status.Offline.Items[1].Name == "etcd2" && len(componentsVersion.Status.Offline.Items[1].VersionRange) == 2
			},
			want: true,
		},
		{
			name: "add docker-ce info",
			args: func() bool {
				updated := componentsVersion.Status.Offline.MergeDockerInfo("centos7", []string{"20.01", "20.02"})
				return updated && len(componentsVersion.Status.Offline.Docker) == 1 && len(componentsVersion.Status.Offline.Docker[0].VersionRange) == 2
			},
			want: true,
		},
		{
			name: "update docker-ce info",
			args: func() bool {
				updated := componentsVersion.Status.Offline.MergeDockerInfo("centos7", []string{"20.01", "20.02"})
				return !updated && len(componentsVersion.Status.Offline.Docker) == 1 && len(componentsVersion.Status.Offline.Docker[0].VersionRange) == 2
			},
			want: true,
		},
		{
			name: "update docker-ce info with zero length",
			args: func() bool {
				updated := componentsVersion.Status.Offline.MergeDockerInfo("centos0008", nil)
				return updated && len(componentsVersion.Status.Offline.Docker) == 2 && len(componentsVersion.Status.Offline.Docker[1].VersionRange) == 0
			},
			want: true,
		},
		{
			name: "update docker-ce info",
			args: func() bool {
				updated := componentsVersion.Status.Offline.MergeDockerInfo("centos0008", []string{"1.1", "1.2", "1.3"})
				return updated && len(componentsVersion.Status.Offline.Docker) == 2 && len(componentsVersion.Status.Offline.Docker[1].VersionRange) == 3
			},
			want: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.args() != test.want {
				t.Fatal()
			}
		})
	}
}
