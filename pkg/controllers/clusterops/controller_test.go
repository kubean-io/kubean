package clusterops

import (
	"context"
	"fmt"
	"testing"
	"time"

	clientsetfake "k8s.io/client-go/kubernetes/fake"
	kubeanclusterv1alpha1 "kubean.io/api/apis/kubeancluster/v1alpha1"
	kubeanclusteropsv1alpha1 "kubean.io/api/apis/kubeanclusterops/v1alpha1"
	kubeanclusterv1alpha1fake "kubean.io/api/generated/kubeancluster/clientset/versioned/fake"
	kubeanclusteropsv1alpha1fake "kubean.io/api/generated/kubeanclusterops/clientset/versioned/fake"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"kubean.io/api/apis"
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

func TestUpdateStatusLoop(t *testing.T) {
	controller := Controller{}
	controller.Client = newFakeClient()
	ops := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
	ops.ObjectMeta.Name = "clusteropsname"
	controller.Client.Create(context.Background(), &ops)
	ops.Spec.BackoffLimit = 12
	tests := []struct {
		name string
		args func(ops *kubeanclusteropsv1alpha1.KuBeanClusterOps) bool
		want bool
	}{
		{
			name: "the status is Succeeded",
			args: func(ops *kubeanclusteropsv1alpha1.KuBeanClusterOps) bool {
				ops.Status.Status = kubeanclusteropsv1alpha1.SucceededStatus
				needRequeue, err := controller.UpdateStatusLoop(ops, nil)
				return err == nil && !needRequeue
			},
			want: true,
		},
		{
			name: "the status is Failed",
			args: func(ops *kubeanclusteropsv1alpha1.KuBeanClusterOps) bool {
				ops.Status.Status = kubeanclusteropsv1alpha1.FailedStatus
				needRequeue, err := controller.UpdateStatusLoop(ops, nil)
				return err == nil && !needRequeue
			},
			want: true,
		},
		{
			name: "the status is Running and the result of fetchJobStatus is err",
			args: func(ops *kubeanclusteropsv1alpha1.KuBeanClusterOps) bool {
				ops.Status.Status = kubeanclusteropsv1alpha1.RunningStatus
				needRequeue, err := controller.UpdateStatusLoop(ops, func(ops *kubeanclusteropsv1alpha1.KuBeanClusterOps) (kubeanclusteropsv1alpha1.OpsStatus, error) {
					return "", fmt.Errorf("one error")
				})
				return err != nil && err.Error() == "one error" && !needRequeue
			},
			want: true,
		},
		{
			name: "the status is Running and the result of fetchJobStatus is still Running",
			args: func(ops *kubeanclusteropsv1alpha1.KuBeanClusterOps) bool {
				ops.Status.Status = kubeanclusteropsv1alpha1.RunningStatus
				needRequeue, err := controller.UpdateStatusLoop(ops, func(ops *kubeanclusteropsv1alpha1.KuBeanClusterOps) (kubeanclusteropsv1alpha1.OpsStatus, error) {
					return kubeanclusteropsv1alpha1.RunningStatus, nil
				})
				return err == nil && needRequeue
			},
			want: true,
		},
		{
			name: "the status is Running and the result of fetchJobStatus is still Succeed",
			args: func(ops *kubeanclusteropsv1alpha1.KuBeanClusterOps) bool {
				ops.Status.Status = kubeanclusteropsv1alpha1.RunningStatus
				ops.Status.EndTime = nil
				needRequeue, err := controller.UpdateStatusLoop(ops, func(ops *kubeanclusteropsv1alpha1.KuBeanClusterOps) (kubeanclusteropsv1alpha1.OpsStatus, error) {
					return kubeanclusteropsv1alpha1.SucceededStatus, nil
				})
				resultOps := &kubeanclusteropsv1alpha1.KuBeanClusterOps{}
				controller.Get(context.Background(), client.ObjectKey{Name: "clusteropsname"}, resultOps)
				if resultOps.Status.Status != kubeanclusteropsv1alpha1.SucceededStatus {
					return false
				}
				return err == nil && !needRequeue && ops.Status.Status == kubeanclusteropsv1alpha1.SucceededStatus && ops.Status.EndTime != nil
			},
			want: true,
		},
		{
			name: "the status is Running and the result of fetchJobStatus is still Failed",
			args: func(ops *kubeanclusteropsv1alpha1.KuBeanClusterOps) bool {
				controller.Get(context.Background(), client.ObjectKey{Name: "clusteropsname"}, ops)
				ops.Status.Status = kubeanclusteropsv1alpha1.RunningStatus
				ops.Status.EndTime = nil
				needRequeue, err := controller.UpdateStatusLoop(ops, func(ops *kubeanclusteropsv1alpha1.KuBeanClusterOps) (kubeanclusteropsv1alpha1.OpsStatus, error) {
					return kubeanclusteropsv1alpha1.FailedStatus, nil
				})
				return err == nil && !needRequeue && ops.Status.EndTime != nil && ops.Status.Status == kubeanclusteropsv1alpha1.FailedStatus
			},
			want: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opsCopy := ops
			if test.args(&opsCopy) != test.want {
				t.Fatal()
			}
		})
	}
}

func TestCompareSalt(t *testing.T) {
	controller := Controller{}
	controller.Client = newFakeClient()
	ops := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
	ops.ObjectMeta.Name = "clusteropsname"
	controller.Client.Create(context.Background(), &ops)
	ops.Spec.BackoffLimit = 12
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "same salt",
			args: func() bool {
				ops.Status.Digest = ""
				controller.UpdateClusterOpsStatusDigest(&ops)
				if len(ops.Status.Digest) == 0 {
					return false
				}
				return controller.compareDigest(&ops)
			},
			want: true,
		},
		{
			name: "not same salt with different spec value",
			args: func() bool {
				ops.Status.Digest = ""
				controller.UpdateClusterOpsStatusDigest(&ops)
				if len(ops.Status.Digest) == 0 {
					return false
				}
				ops.Spec.BackoffLimit = 12345
				return !controller.compareDigest(&ops)
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

func TestUpdateClusterOpsStatusSalt(t *testing.T) {
	controller := Controller{}
	controller.Client = newFakeClient()
	ops := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
	ops.ObjectMeta.Name = "clusteropsname"
	controller.Client.Create(context.Background(), &ops)
	ops.Spec.BackoffLimit = 12
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "salt empty value",
			args: func() bool {
				ops.Status.Digest = ""
				needRequeue, err := controller.UpdateClusterOpsStatusDigest(&ops)
				return needRequeue && err == nil && len(ops.Status.Digest) > 0
			},
			want: true,
		},
		{
			name: "nothing changed",
			args: func() bool {
				needRequeue, err := controller.UpdateClusterOpsStatusDigest(&ops)
				return !needRequeue && err == nil && len(ops.Status.Digest) > 0
			},
			want: true,
		},
		{
			name: "salt empty value again",
			args: func() bool {
				ops.Status.Digest = ""
				needRequeue, err := controller.UpdateClusterOpsStatusDigest(&ops)
				return needRequeue && err == nil && len(ops.Status.Digest) > 0
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

func TestUpdateStatusHasModified(t *testing.T) {
	controller := Controller{}
	controller.Client = newFakeClient()
	ops := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
	ops.ObjectMeta.Name = "clusteropsname"
	controller.Client.Create(context.Background(), &ops)
	ops.Spec.BackoffLimit = 12
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "empty salt value",
			args: func() bool {
				ops.Status.Digest = ""
				needRequeue, err := controller.UpdateStatusHasModified(&ops)
				return !needRequeue && err == nil && !ops.Status.HasModified
			},
			want: true,
		},
		{
			name: "already modified",
			args: func() bool {
				ops.Status.Digest = "123"
				ops.Status.HasModified = true
				needRequeue, err := controller.UpdateStatusHasModified(&ops)
				return !needRequeue && err == nil && ops.Status.HasModified
			},
			want: true,
		},
		{
			name: "nothing need update",
			args: func() bool {
				ops.Status.Digest = ""
				ops.Status.HasModified = false
				controller.UpdateClusterOpsStatusDigest(&ops)
				if len(ops.Status.Digest) == 0 {
					return false
				}
				needRequeue, err := controller.UpdateStatusHasModified(&ops)
				return len(ops.Status.Digest) != 0 && err == nil && !needRequeue && !ops.Status.HasModified
			},
			want: true,
		},
		{
			name: "something update",
			args: func() bool {
				ops.Status.Digest = ""
				ops.Status.HasModified = false
				controller.UpdateClusterOpsStatusDigest(&ops)
				if len(ops.Status.Digest) == 0 {
					return false
				}
				ops.Spec.BackoffLimit = 111
				needRequeue, err := controller.UpdateStatusHasModified(&ops)
				return len(ops.Status.Digest) != 0 && err == nil && needRequeue && ops.Status.HasModified
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
				return job.Namespace == "mynamespace" && job.Name == "kubean-myops-job" && len(job.Spec.Template.Spec.Containers) == 1 && len(job.Spec.Template.Spec.Containers[0].VolumeMounts) == 3 && len(job.Spec.Template.Spec.Volumes) == 3
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
				return job.Namespace == "mynamespace" && job.Name == "kubean-myops-job" && len(job.Spec.Template.Spec.Containers) == 1 && len(job.Spec.Template.Spec.Containers[0].VolumeMounts) == 4 && len(job.Spec.Template.Spec.Volumes) == 4
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

func TestCurrentJobNeedBlock(t *testing.T) {
	controller := Controller{}
	clusterOps := &kubeanclusteropsv1alpha1.KuBeanClusterOps{}
	clusterOps.Name = "the target one clusterOps"
	clusterOps.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000)}
	testCases := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "it occurs error",
			args: func() bool {
				_, err := controller.CurrentJobNeedBlock(clusterOps, func(clusterName string) ([]kubeanclusteropsv1alpha1.KuBeanClusterOps, error) {
					return nil, fmt.Errorf("error")
				})
				return err == nil
			},
			want: false,
		},
		{
			name: "it returns none jobs running before the target clusterOps",
			args: func() bool {
				needBlock, err := controller.CurrentJobNeedBlock(clusterOps, func(clusterName string) ([]kubeanclusteropsv1alpha1.KuBeanClusterOps, error) {
					return nil, nil
				})
				return err == nil && !needBlock
			},
			want: true,
		},
		{
			name: "it returns none jobs running before the target clusterOps again",
			args: func() bool {
				needBlock, err := controller.CurrentJobNeedBlock(clusterOps, func(clusterName string) ([]kubeanclusteropsv1alpha1.KuBeanClusterOps, error) {
					ops1 := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
					ops1.Name = "ops1"
					ops1.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000 + 10000)}
					ops2 := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
					ops2.Name = "ops2"
					ops2.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000 + 10000)}
					return []kubeanclusteropsv1alpha1.KuBeanClusterOps{ops1, ops2}, nil
				})
				return err == nil && !needBlock
			},
			want: true,
		},
		{
			name: "it returns some jobs running before the target clusterOps",
			args: func() bool {
				needBlock, err := controller.CurrentJobNeedBlock(clusterOps, func(clusterName string) ([]kubeanclusteropsv1alpha1.KuBeanClusterOps, error) {
					ops1 := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
					ops1.Name = "ops1"
					ops1.CreationTimestamp = metav1.Time{Time: time.UnixMilli(0)}
					ops2 := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
					ops2.Name = "ops2"
					ops2.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000 + 10000)}
					return []kubeanclusteropsv1alpha1.KuBeanClusterOps{ops1, ops2}, nil
				})
				return err == nil && needBlock
			},
			want: true,
		},
		{
			name: "it returns some jobs running before the target clusterOps with the same createTime",
			args: func() bool {
				needBlock, err := controller.CurrentJobNeedBlock(clusterOps, func(clusterName string) ([]kubeanclusteropsv1alpha1.KuBeanClusterOps, error) {
					ops1 := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
					ops1.Name = "ops1"
					ops1.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000)}
					ops1.Status.Status = kubeanclusteropsv1alpha1.RunningStatus
					ops2 := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
					ops2.Name = "ops2"
					ops2.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000 + 10000)}
					return []kubeanclusteropsv1alpha1.KuBeanClusterOps{ops1, ops2}, nil
				})
				return err == nil && !needBlock
			},
			want: true,
		},
		{
			name: "it returns some jobs completed before the target clusterOps",
			args: func() bool {
				needBlock, err := controller.CurrentJobNeedBlock(clusterOps, func(clusterName string) ([]kubeanclusteropsv1alpha1.KuBeanClusterOps, error) {
					ops1 := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
					ops1.Name = "ops1"
					ops1.CreationTimestamp = metav1.Time{Time: time.UnixMilli(0)}
					ops1.Status.Status = kubeanclusteropsv1alpha1.SucceededStatus
					ops1.Status.JobRef = &apis.JobRef{Name: "ok", NameSpace: "ok"}
					ops2 := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
					ops2.Name = "ops2"
					ops2.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000 + 10000)}
					return []kubeanclusteropsv1alpha1.KuBeanClusterOps{ops1, ops2}, nil
				})
				return err == nil && !needBlock
			},
			want: true,
		},
		{
			name: "it returns some jobs failed before the target clusterOps",
			args: func() bool {
				needBlock, err := controller.CurrentJobNeedBlock(clusterOps, func(clusterName string) ([]kubeanclusteropsv1alpha1.KuBeanClusterOps, error) {
					ops1 := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
					ops1.Name = "ops1"
					ops1.CreationTimestamp = metav1.Time{Time: time.UnixMilli(0)}
					ops1.Status.Status = kubeanclusteropsv1alpha1.FailedStatus
					ops1.Status.JobRef = &apis.JobRef{Name: "ok", NameSpace: "ok"}
					ops2 := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
					ops2.Name = "ops2"
					ops2.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000 + 10000)}
					return []kubeanclusteropsv1alpha1.KuBeanClusterOps{ops1, ops2}, nil
				})
				return err == nil && !needBlock
			},
			want: true,
		},
		{
			name: "it returns some jobs blocked before the target clusterOps",
			args: func() bool {
				needBlock, err := controller.CurrentJobNeedBlock(clusterOps, func(clusterName string) ([]kubeanclusteropsv1alpha1.KuBeanClusterOps, error) {
					ops1 := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
					ops1.Name = "ops1"
					ops1.CreationTimestamp = metav1.Time{Time: time.UnixMilli(0)}
					ops1.Status.Status = kubeanclusteropsv1alpha1.BlockedStatus
					ops2 := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
					ops2.Name = "ops2"
					ops2.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000 + 10000)}
					return []kubeanclusteropsv1alpha1.KuBeanClusterOps{ops1, ops2}, nil
				})
				return err == nil && needBlock
			},
			want: true,
		},
		{
			name: "it returns some jobs blocked at the same createTime",
			args: func() bool {
				needBlock, err := controller.CurrentJobNeedBlock(clusterOps, func(clusterName string) ([]kubeanclusteropsv1alpha1.KuBeanClusterOps, error) {
					ops1 := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
					ops1.Name = "ops1"
					ops1.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000)}
					ops1.Status.Status = kubeanclusteropsv1alpha1.BlockedStatus
					ops2 := kubeanclusteropsv1alpha1.KuBeanClusterOps{}
					ops2.Name = "ops2"
					ops2.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000 + 10000)}
					return []kubeanclusteropsv1alpha1.KuBeanClusterOps{ops1, ops2}, nil
				})
				return err == nil && !needBlock
			},
			want: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.args() != testCase.want {
				t.Fatal()
			}
		})
	}
}

func TestIsValidImageName(t *testing.T) {
	testCases := []struct {
		name string
		args string
		want bool
	}{
		{
			name: "empty string",
			args: "",
			want: false,
		},
		{
			name: "space string",
			args: " ",
			want: false,
		},
		{
			name: "underscore string",
			args: "_",
			want: false,
		},
		{
			name: "dot string",
			args: ".",
			want: false,
		},
		{
			name: "one letter",
			args: "a",
			want: true,
		},
		{
			name: "ubuntu",
			args: "ubuntu",
			want: true,
		},
		{
			name: "ubuntu with tag",
			args: "ubuntu:14.01",
			want: true,
		},
		{
			name: "valid image name",
			args: "ghcr.io/kubean-io/spray-job:latest",
			want: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if IsValidImageName(testCase.args) != testCase.want {
				t.Fatal()
			}
		})
	}
}

func TestCreateKubeSprayJob(t *testing.T) {
	controller := Controller{
		Client:              newFakeClient(),
		ClientSet:           clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:    kubeanclusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet: kubeanclusteropsv1alpha1fake.NewSimpleClientset(),
	}
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
	controller.CreateKubeSprayJob(clusterOps)
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "create job successfully",
			args: func() bool {
				needRequeue, err := controller.CreateKubeSprayJob(clusterOps)
				return needRequeue && err == nil
			},
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
