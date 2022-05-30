package clusterops

import (
	"testing"

	kubeanclusteropsv1alpha1 "github.com/daocloud/kubean/pkg/apis/kubeanclusterops/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

func TestController_SetOwnerReferences(t *testing.T) {
	cm := corev1.ConfigMap{}
	ops := &kubeanclusteropsv1alpha1.KuBeanClusterOps{}
	ops.UID = "this is uid"
	controller := Controller{}
	controller.SetOwnerReferences(&cm.ObjectMeta, ops)
	if len(cm.OwnerReferences) == 0 {
		t.Fatal()
	}
	if cm.OwnerReferences[0].UID != "this is uid" {
		t.Fatal()
	}
}
