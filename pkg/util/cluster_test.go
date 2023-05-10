package util

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	clusterv1alpha1 "github.com/kubean-io/kubean-api/apis/cluster/v1alpha1"
	clusteroperationv1alpha1 "github.com/kubean-io/kubean-api/apis/clusteroperation/v1alpha1"
	localartifactsetv1alpha1 "github.com/kubean-io/kubean-api/apis/localartifactset/v1alpha1"
	manifestv1alpha1 "github.com/kubean-io/kubean-api/apis/manifest/v1alpha1"
)

func TestNewSchema(t *testing.T) {
	aggregatedScheme := NewSchema()
	tests := []struct {
		name    string
		obj     runtime.Object
		wantGVK schema.GroupVersionKind
	}{
		{
			name:    "KuBeanInfoManifest gvk",
			obj:     &manifestv1alpha1.Manifest{},
			wantGVK: schema.GroupVersionKind{Group: "kubean.io", Version: "v1alpha1", Kind: "Manifest"},
		},
		{
			name:    "LocalArtifactSet gvk",
			obj:     &localartifactsetv1alpha1.LocalArtifactSet{},
			wantGVK: schema.GroupVersionKind{Group: "kubean.io", Version: "v1alpha1", Kind: "LocalArtifactSet"},
		},
		{
			name:    "Cluster gvk",
			obj:     &clusterv1alpha1.Cluster{},
			wantGVK: schema.GroupVersionKind{Group: "kubean.io", Version: "v1alpha1", Kind: "Cluster"},
		},
		{
			name:    "ClusterOperation gvk",
			obj:     &clusteroperationv1alpha1.ClusterOperation{},
			wantGVK: schema.GroupVersionKind{Group: "kubean.io", Version: "v1alpha1", Kind: "ClusterOperation"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if gvkArray, _, _ := aggregatedScheme.ObjectKinds(test.obj); gvkArray[0] != test.wantGVK {
				t.Fatal()
			}
		})
	}
}

func TestGetCurrentNSOrDefault(t *testing.T) {
	if namespace := GetCurrentNSOrDefault(); namespace == "" {
		t.Fatal()
	}
}
