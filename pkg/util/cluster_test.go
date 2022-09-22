package util

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	kubeanclusterv1alpha1 "kubean.io/api/apis/kubeancluster/v1alpha1"
	kubeanclusteropsv1alpha1 "kubean.io/api/apis/kubeanclusterops/v1alpha1"
	kubeaninfomanifestv1alpha1 "kubean.io/api/apis/kubeaninfomanifest/v1alpha1"
	kubeanofflineversionv1alpha1 "kubean.io/api/apis/kubeanofflineversion/v1alpha1"
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
			obj:     &kubeaninfomanifestv1alpha1.KubeanInfoManifest{},
			wantGVK: schema.GroupVersionKind{Group: "kubean.io", Version: "v1alpha1", Kind: "KubeanInfoManifest"},
		},
		{
			name:    "KuBeanOfflineVersion gvk",
			obj:     &kubeanofflineversionv1alpha1.KuBeanOfflineVersion{},
			wantGVK: schema.GroupVersionKind{Group: "kubean.io", Version: "v1alpha1", Kind: "KuBeanOfflineVersion"},
		},
		{
			name:    "KuBeanCluster gvk",
			obj:     &kubeanclusterv1alpha1.KuBeanCluster{},
			wantGVK: schema.GroupVersionKind{Group: "kubean.io", Version: "v1alpha1", Kind: "KuBeanCluster"},
		},
		{
			name:    "KuBeanClusterOps gvk",
			obj:     &kubeanclusteropsv1alpha1.KuBeanClusterOps{},
			wantGVK: schema.GroupVersionKind{Group: "kubean.io", Version: "v1alpha1", Kind: "KuBeanClusterOps"},
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
