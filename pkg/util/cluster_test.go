package util

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	kubeancomponentsversionv1alpha1 "kubean.io/api/apis/kubeancomponentsversion/v1alpha1"
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
			name:    "KuBeanComponentsVersion gvk",
			obj:     &kubeancomponentsversionv1alpha1.KuBeanComponentsVersion{},
			wantGVK: schema.GroupVersionKind{Group: "kubean.io", Version: "v1alpha1", Kind: "KuBeanComponentsVersion"},
		},
		{
			name:    "KuBeanOfflineVersion gvk",
			obj:     &kubeanofflineversionv1alpha1.KuBeanOfflineVersion{},
			wantGVK: schema.GroupVersionKind{Group: "kubean.io", Version: "v1alpha1", Kind: "KuBeanOfflineVersion"},
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
