package clusterops

import (
	"testing"

	"github.com/daocloud/kubean/pkg/apis"
	kubeanclusterv1alpha1 "github.com/daocloud/kubean/pkg/apis/kubeancluster/v1alpha1"
	kubeanclusteropsv1alpha1 "github.com/daocloud/kubean/pkg/apis/kubeanclusterops/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newFakeClient() client.Client {
	sch := scheme.Scheme
	if err := kubeanclusteropsv1alpha1.AddToScheme(sch); err != nil {
		panic(err)
	}
	if err := kubeanclusterv1alpha1.AddToScheme(sch); err != nil {
		panic(err)
	}
	client := fake.NewClientBuilder().WithScheme(sch).WithRuntimeObjects(&kubeanclusteropsv1alpha1.KuBeanClusterOps{}).WithRuntimeObjects(&kubeanclusterv1alpha1.KuBeanCluster{}).Build()
	return client
}

func TestUpdateClusterOpsStatusSalt(t *testing.T) {
	controller := Controller{}
	controller.Client = newFakeClient()
	ops := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
	ops.Spec.BackoffLimit = 12
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "salt empty value",
			args: func() bool {
				ops.Status.Salt = ""
				needRequeue, err := controller.UpdateClusterOpsStatusSalt(&ops)
				return needRequeue && err == nil && len(ops.Status.Salt) > 0
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

func TestCalSalt(t *testing.T) {
	controller := Controller{}
	ops := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
	ops.Spec.KuBeanCluster = "123456789"
	ops.Spec.ActionType = "1"
	ops.Spec.Action = "2"
	ops.Spec.BackoffLimit = 3
	ops.Spec.Image = "4"
	ops.Spec.PreHook = []kubeanclusteropsv1alpha1.HookAction{
		{
			ActionType: "11",
			Action:     "22",
		},
		{
			ActionType: "22",
			Action:     "33",
		},
	}
	ops.Spec.PostHook = []kubeanclusteropsv1alpha1.HookAction{
		{
			ActionType: "55",
			Action:     "66",
		},
	}
	targetSaltValue := controller.CalSalt(&ops)
	tests := []struct {
		name string
		args func(kubeanclusteropsv1alpha1.KuBeanClusterOps) string
		want bool
	}{
		{
			name: "change clusterName",
			args: func(ops kubeanclusteropsv1alpha1.KuBeanClusterOps) string {
				ops.Spec.KuBeanCluster = "ok1"
				return controller.CalSalt(&ops)
			},
			want: false,
		},
		{
			name: "change actionType",
			args: func(ops kubeanclusteropsv1alpha1.KuBeanClusterOps) string {
				ops.Spec.ActionType = "luck"
				return controller.CalSalt(&ops)
			},
			want: false,
		},
		{
			name: "change action",
			args: func(ops kubeanclusteropsv1alpha1.KuBeanClusterOps) string {
				ops.Spec.Action = "ok123"
				return controller.CalSalt(&ops)
			},
			want: false,
		},
		{
			name: "change backoff",
			args: func(ops kubeanclusteropsv1alpha1.KuBeanClusterOps) string {
				ops.Spec.BackoffLimit = 100
				return controller.CalSalt(&ops)
			},
			want: false,
		},
		{
			name: "change image",
			args: func(ops kubeanclusteropsv1alpha1.KuBeanClusterOps) string {
				ops.Spec.Image = "image123"
				return controller.CalSalt(&ops)
			},
			want: false,
		},
		{
			name: "unchanged",
			args: func(ops kubeanclusteropsv1alpha1.KuBeanClusterOps) string {
				return controller.CalSalt(&ops)
			},
			want: true,
		},
		{
			name: "change postHook",
			args: func(ops kubeanclusteropsv1alpha1.KuBeanClusterOps) string {
				ops.Spec.PostHook[0].ActionType = "ok12qaz"
				return controller.CalSalt(&ops)
			},
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.args(ops) == targetSaltValue
			if result != test.want {
				t.Fatal()
			}
		})
	}
}

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
