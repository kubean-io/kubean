package clusterops

import (
	"testing"

	"github.com/daocloud/kubean/pkg/apis"
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

func TestNewKubesprayJob(t *testing.T) {
	controller := Controller{}
	clusterOps := &kubeanclusteropsv1alpha1.KuBeanClusterOps{}
	clusterOps.Spec.BackoffLimit = 10
	clusterOps.Name = "myops"
	clusterOps.Spec.Image = "myimage"
	clusterOps.Spec.HostsConfRef = &apis.ConfigMapRef{
		NameSpace: "mynamespace",
		Name:      "hostsconf",
	}
	clusterOps.Spec.VarsConfRef = &apis.ConfigMapRef{
		NameSpace: "mynamespace",
		Name:      "varsconf",
	}
	clusterOps.Spec.EntrypointSHRef = &apis.ConfigMapRef{
		NameSpace: "mynamespace",
		Name:      "entrypoint",
	}
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "no ssh args",
			args: func() bool {
				job := controller.NewKubesprayJob(clusterOps)
				return job.Namespace == "mynamespace" && job.Name == "myops-job" && len(job.Spec.Template.Spec.Containers) == 1 && len(job.Spec.Template.Spec.Containers[0].VolumeMounts) == 3 && len(job.Spec.Template.Spec.Volumes) == 3
			},
			want: true,
		},
		{
			name: "ssh args",
			args: func() bool {
				clusterOps.Spec.SSHAuthRef = &apis.SecretRef{
					NameSpace: "mynamespace",
					Name:      "secret",
				}
				job := controller.NewKubesprayJob(clusterOps)
				return job.Namespace == "mynamespace" && job.Name == "myops-job" && len(job.Spec.Template.Spec.Containers) == 1 && len(job.Spec.Template.Spec.Containers[0].VolumeMounts) == 4 && len(job.Spec.Template.Spec.Volumes) == 4
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
