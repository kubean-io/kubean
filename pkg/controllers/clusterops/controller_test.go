package clusterops

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clientsetfake "k8s.io/client-go/kubernetes/fake"
	clusterv1alpha1 "kubean.io/api/apis/cluster/v1alpha1"
	clusteroperationv1alpha1 "kubean.io/api/apis/clusteroperation/v1alpha1"
	manifestv1alpha1 "kubean.io/api/apis/manifest/v1alpha1"
	"kubean.io/api/constants"
	clusterv1alpha1fake "kubean.io/api/generated/cluster/clientset/versioned/fake"
	clusteroperationv1alpha1fake "kubean.io/api/generated/clusteroperation/clientset/versioned/fake"
	manifestv1alpha1fake "kubean.io/api/generated/manifest/clientset/versioned/fake"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"kubean.io/api/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newFakeClient() client.Client {
	sch := scheme.Scheme
	if err := clusteroperationv1alpha1.AddToScheme(sch); err != nil {
		panic(err)
	}
	if err := clusterv1alpha1.AddToScheme(sch); err != nil {
		panic(err)
	}
	client := fake.NewClientBuilder().WithScheme(sch).WithRuntimeObjects(&clusteroperationv1alpha1.ClusterOperation{}).WithRuntimeObjects(&clusterv1alpha1.Cluster{}).Build()
	return client
}

func TestUpdateStatusLoop(t *testing.T) {
	controller := Controller{}
	controller.Client = newFakeClient()
	ops := clusteroperationv1alpha1.ClusterOperation{}
	ops.ObjectMeta.Name = "clusteropsname"
	controller.Client.Create(context.Background(), &ops)
	ops.Spec.BackoffLimit = 12
	tests := []struct {
		name string
		args func(ops *clusteroperationv1alpha1.ClusterOperation) bool
		want bool
	}{
		{
			name: "the status is Succeeded",
			args: func(ops *clusteroperationv1alpha1.ClusterOperation) bool {
				ops.Status.Status = clusteroperationv1alpha1.SucceededStatus
				needRequeue, err := controller.UpdateStatusLoop(ops, nil)
				return err == nil && !needRequeue
			},
			want: true,
		},
		{
			name: "the status is Failed",
			args: func(ops *clusteroperationv1alpha1.ClusterOperation) bool {
				ops.Status.Status = clusteroperationv1alpha1.FailedStatus
				needRequeue, err := controller.UpdateStatusLoop(ops, nil)
				return err == nil && !needRequeue
			},
			want: true,
		},
		{
			name: "the status is Running and the result of fetchJobStatus is err",
			args: func(ops *clusteroperationv1alpha1.ClusterOperation) bool {
				ops.Status.Status = clusteroperationv1alpha1.RunningStatus
				needRequeue, err := controller.UpdateStatusLoop(ops, func(ops *clusteroperationv1alpha1.ClusterOperation) (clusteroperationv1alpha1.OpsStatus, error) {
					return "", fmt.Errorf("one error")
				})
				return err != nil && err.Error() == "one error" && !needRequeue
			},
			want: true,
		},
		{
			name: "the status is Running and the result of fetchJobStatus is still Running",
			args: func(ops *clusteroperationv1alpha1.ClusterOperation) bool {
				ops.Status.Status = clusteroperationv1alpha1.RunningStatus
				needRequeue, err := controller.UpdateStatusLoop(ops, func(ops *clusteroperationv1alpha1.ClusterOperation) (clusteroperationv1alpha1.OpsStatus, error) {
					return clusteroperationv1alpha1.RunningStatus, nil
				})
				return err == nil && needRequeue
			},
			want: true,
		},
		{
			name: "the status is Running and the result of fetchJobStatus is still Succeed",
			args: func(ops *clusteroperationv1alpha1.ClusterOperation) bool {
				ops.Status.Status = clusteroperationv1alpha1.RunningStatus
				ops.Status.EndTime = nil
				needRequeue, err := controller.UpdateStatusLoop(ops, func(ops *clusteroperationv1alpha1.ClusterOperation) (clusteroperationv1alpha1.OpsStatus, error) {
					return clusteroperationv1alpha1.SucceededStatus, nil
				})
				resultOps := &clusteroperationv1alpha1.ClusterOperation{}
				controller.Get(context.Background(), client.ObjectKey{Name: "clusteropsname"}, resultOps)
				if resultOps.Status.Status != clusteroperationv1alpha1.SucceededStatus {
					return false
				}
				return err == nil && !needRequeue && ops.Status.Status == clusteroperationv1alpha1.SucceededStatus && ops.Status.EndTime != nil
			},
			want: true,
		},
		{
			name: "the status is Running and the result of fetchJobStatus is still Failed",
			args: func(ops *clusteroperationv1alpha1.ClusterOperation) bool {
				controller.Get(context.Background(), client.ObjectKey{Name: "clusteropsname"}, ops)
				ops.Status.Status = clusteroperationv1alpha1.RunningStatus
				ops.Status.EndTime = nil
				needRequeue, err := controller.UpdateStatusLoop(ops, func(ops *clusteroperationv1alpha1.ClusterOperation) (clusteroperationv1alpha1.OpsStatus, error) {
					return clusteroperationv1alpha1.FailedStatus, nil
				})
				return err == nil && !needRequeue && ops.Status.EndTime != nil && ops.Status.Status == clusteroperationv1alpha1.FailedStatus
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
	ops := clusteroperationv1alpha1.ClusterOperation{}
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
	ops := clusteroperationv1alpha1.ClusterOperation{}
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
	ops := clusteroperationv1alpha1.ClusterOperation{}
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
	ops := clusteroperationv1alpha1.ClusterOperation{}
	ops.Spec.Cluster = "123456789"
	ops.Spec.ActionType = "1"
	ops.Spec.Action = "2"
	ops.Spec.BackoffLimit = 3
	ops.Spec.Image = "4"
	ops.Spec.PreHook = []clusteroperationv1alpha1.HookAction{
		{
			ActionType: "11",
			Action:     "22",
		},
		{
			ActionType: "22",
			Action:     "33",
		},
	}
	ops.Spec.PostHook = []clusteroperationv1alpha1.HookAction{
		{
			ActionType: "55",
			Action:     "66",
		},
	}
	targetSaltValue := controller.CalSalt(&ops)
	tests := []struct {
		name string
		args func(clusteroperationv1alpha1.ClusterOperation) string
		want bool
	}{
		{
			name: "change clusterName",
			args: func(ops clusteroperationv1alpha1.ClusterOperation) string {
				ops.Spec.Cluster = "ok1"
				return controller.CalSalt(&ops)
			},
			want: false,
		},
		{
			name: "change actionType",
			args: func(ops clusteroperationv1alpha1.ClusterOperation) string {
				ops.Spec.ActionType = "luck"
				return controller.CalSalt(&ops)
			},
			want: false,
		},
		{
			name: "change action",
			args: func(ops clusteroperationv1alpha1.ClusterOperation) string {
				ops.Spec.Action = "ok123"
				return controller.CalSalt(&ops)
			},
			want: false,
		},
		{
			name: "change backoff",
			args: func(ops clusteroperationv1alpha1.ClusterOperation) string {
				ops.Spec.BackoffLimit = 100
				return controller.CalSalt(&ops)
			},
			want: false,
		},
		{
			name: "change image",
			args: func(ops clusteroperationv1alpha1.ClusterOperation) string {
				ops.Spec.Image = "image123"
				return controller.CalSalt(&ops)
			},
			want: false,
		},
		{
			name: "unchanged",
			args: func(ops clusteroperationv1alpha1.ClusterOperation) string {
				return controller.CalSalt(&ops)
			},
			want: true,
		},
		{
			name: "change postHook",
			args: func(ops clusteroperationv1alpha1.ClusterOperation) string {
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
	ops := &clusteroperationv1alpha1.ClusterOperation{}
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
	controller := Controller{
		Client:                newFakeClient(),
		ClientSet:             clientsetfake.NewSimpleClientset(),
		InfoManifestClientSet: manifestv1alpha1fake.NewSimpleClientset(),
	}
	clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
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
		{
			name: "activeDeadlineSeconds args",
			args: func() bool {
				ActiveDeadlineSeconds := int64(10)
				clusterOps.Spec.ActiveDeadlineSeconds = &ActiveDeadlineSeconds
				job := controller.NewKubesprayJob(clusterOps)
				return *job.Spec.ActiveDeadlineSeconds == 10
			},
			want: true,
		},
		{
			name: "nil activeDeadlineSeconds args",
			args: func() bool {
				clusterOps.Spec.ActiveDeadlineSeconds = nil
				job := controller.NewKubesprayJob(clusterOps)
				return job.Spec.ActiveDeadlineSeconds == nil
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
	clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
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
				_, err := controller.CurrentJobNeedBlock(clusterOps, func(clusterName string) ([]clusteroperationv1alpha1.ClusterOperation, error) {
					return nil, fmt.Errorf("error")
				})
				return err == nil
			},
			want: false,
		},
		{
			name: "it returns none jobs running before the target clusterOps",
			args: func() bool {
				needBlock, err := controller.CurrentJobNeedBlock(clusterOps, func(clusterName string) ([]clusteroperationv1alpha1.ClusterOperation, error) {
					return nil, nil
				})
				return err == nil && !needBlock
			},
			want: true,
		},
		{
			name: "it returns none jobs running before the target clusterOps again",
			args: func() bool {
				needBlock, err := controller.CurrentJobNeedBlock(clusterOps, func(clusterName string) ([]clusteroperationv1alpha1.ClusterOperation, error) {
					ops1 := clusteroperationv1alpha1.ClusterOperation{}
					ops1.Name = "ops1"
					ops1.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000 + 10000)}
					ops2 := clusteroperationv1alpha1.ClusterOperation{}
					ops2.Name = "ops2"
					ops2.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000 + 10000)}
					return []clusteroperationv1alpha1.ClusterOperation{ops1, ops2}, nil
				})
				return err == nil && !needBlock
			},
			want: true,
		},
		{
			name: "it returns some jobs running before the target clusterOps",
			args: func() bool {
				needBlock, err := controller.CurrentJobNeedBlock(clusterOps, func(clusterName string) ([]clusteroperationv1alpha1.ClusterOperation, error) {
					ops1 := clusteroperationv1alpha1.ClusterOperation{}
					ops1.Name = "ops1"
					ops1.CreationTimestamp = metav1.Time{Time: time.UnixMilli(0)}
					ops2 := clusteroperationv1alpha1.ClusterOperation{}
					ops2.Name = "ops2"
					ops2.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000 + 10000)}
					return []clusteroperationv1alpha1.ClusterOperation{ops1, ops2}, nil
				})
				return err == nil && needBlock
			},
			want: true,
		},
		{
			name: "it returns some jobs running before the target clusterOps with the same createTime",
			args: func() bool {
				needBlock, err := controller.CurrentJobNeedBlock(clusterOps, func(clusterName string) ([]clusteroperationv1alpha1.ClusterOperation, error) {
					ops1 := clusteroperationv1alpha1.ClusterOperation{}
					ops1.Name = "ops1"
					ops1.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000)}
					ops1.Status.Status = clusteroperationv1alpha1.RunningStatus
					ops2 := clusteroperationv1alpha1.ClusterOperation{}
					ops2.Name = "ops2"
					ops2.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000 + 10000)}
					return []clusteroperationv1alpha1.ClusterOperation{ops1, ops2}, nil
				})
				return err == nil && !needBlock
			},
			want: true,
		},
		{
			name: "it returns some jobs completed before the target clusterOps",
			args: func() bool {
				needBlock, err := controller.CurrentJobNeedBlock(clusterOps, func(clusterName string) ([]clusteroperationv1alpha1.ClusterOperation, error) {
					ops1 := clusteroperationv1alpha1.ClusterOperation{}
					ops1.Name = "ops1"
					ops1.CreationTimestamp = metav1.Time{Time: time.UnixMilli(0)}
					ops1.Status.Status = clusteroperationv1alpha1.SucceededStatus
					ops1.Status.JobRef = &apis.JobRef{Name: "ok", NameSpace: "ok"}
					ops2 := clusteroperationv1alpha1.ClusterOperation{}
					ops2.Name = "ops2"
					ops2.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000 + 10000)}
					return []clusteroperationv1alpha1.ClusterOperation{ops1, ops2}, nil
				})
				return err == nil && !needBlock
			},
			want: true,
		},
		{
			name: "it returns some jobs failed before the target clusterOps",
			args: func() bool {
				needBlock, err := controller.CurrentJobNeedBlock(clusterOps, func(clusterName string) ([]clusteroperationv1alpha1.ClusterOperation, error) {
					ops1 := clusteroperationv1alpha1.ClusterOperation{}
					ops1.Name = "ops1"
					ops1.CreationTimestamp = metav1.Time{Time: time.UnixMilli(0)}
					ops1.Status.Status = clusteroperationv1alpha1.FailedStatus
					ops1.Status.JobRef = &apis.JobRef{Name: "ok", NameSpace: "ok"}
					ops2 := clusteroperationv1alpha1.ClusterOperation{}
					ops2.Name = "ops2"
					ops2.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000 + 10000)}
					return []clusteroperationv1alpha1.ClusterOperation{ops1, ops2}, nil
				})
				return err == nil && !needBlock
			},
			want: true,
		},
		{
			name: "it returns some jobs blocked before the target clusterOps",
			args: func() bool {
				needBlock, err := controller.CurrentJobNeedBlock(clusterOps, func(clusterName string) ([]clusteroperationv1alpha1.ClusterOperation, error) {
					ops1 := clusteroperationv1alpha1.ClusterOperation{}
					ops1.Name = "ops1"
					ops1.CreationTimestamp = metav1.Time{Time: time.UnixMilli(0)}
					ops1.Status.Status = clusteroperationv1alpha1.BlockedStatus
					ops2 := clusteroperationv1alpha1.ClusterOperation{}
					ops2.Name = "ops2"
					ops2.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000 + 10000)}
					return []clusteroperationv1alpha1.ClusterOperation{ops1, ops2}, nil
				})
				return err == nil && needBlock
			},
			want: true,
		},
		{
			name: "it returns some jobs blocked at the same createTime",
			args: func() bool {
				needBlock, err := controller.CurrentJobNeedBlock(clusterOps, func(clusterName string) ([]clusteroperationv1alpha1.ClusterOperation, error) {
					ops1 := clusteroperationv1alpha1.ClusterOperation{}
					ops1.Name = "ops1"
					ops1.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000)}
					ops1.Status.Status = clusteroperationv1alpha1.BlockedStatus
					ops2 := clusteroperationv1alpha1.ClusterOperation{}
					ops2.Name = "ops2"
					ops2.CreationTimestamp = metav1.Time{Time: time.UnixMilli(1000 + 10000)}
					return []clusteroperationv1alpha1.ClusterOperation{ops1, ops2}, nil
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
		Client:                newFakeClient(),
		ClientSet:             clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:      clusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet:   clusteroperationv1alpha1fake.NewSimpleClientset(),
		InfoManifestClientSet: manifestv1alpha1fake.NewSimpleClientset(),
	}
	clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
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

func Test_CheckConfigMapExist(t *testing.T) {
	controller := Controller{
		Client:              newFakeClient(),
		ClientSet:           clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:    clusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet: clusteroperationv1alpha1fake.NewSimpleClientset(),
	}
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "not exist configMap data",
			args: func() bool {
				return controller.CheckConfigMapExist("my_namespace_cm", "my_name_cm")
			},
			want: false,
		},
		{
			name: "configMap data exists",
			args: func() bool {
				cm := &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "my_namespace_cm_1",
						Name:      "my_name_cm_1",
					},
					Data: map[string]string{"ok": "ok123"},
				}
				controller.ClientSet.CoreV1().ConfigMaps("my_namespace_cm_1").Create(context.Background(), cm, metav1.CreateOptions{})
				return controller.CheckConfigMapExist("my_namespace_cm_1", "my_name_cm_1")
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

func Test_CheckSecretExist(t *testing.T) {
	controller := Controller{
		Client:              newFakeClient(),
		ClientSet:           clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:    clusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet: clusteroperationv1alpha1fake.NewSimpleClientset(),
	}
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "secret data not exist",
			args: func() bool {
				return controller.CheckSecretExist("my_namespace_secret", "my_name_secret")
			},
			want: false,
		},
		{
			name: "secret data exist",
			args: func() bool {
				secret := &corev1.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "my_namespace_secret_1",
						Name:      "my_name_secret_1",
					},
					Data: map[string][]byte{"ok": []byte(base64.StdEncoding.EncodeToString([]byte("ok123")))},
				}
				controller.ClientSet.CoreV1().Secrets("my_namespace_secret_1").Create(context.Background(), secret, metav1.CreateOptions{})
				return controller.CheckSecretExist("my_namespace_secret_1", "my_name_secret_1")
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

func Test_CheckClusterDataRef(t *testing.T) {
	controller := Controller{
		Client:              newFakeClient(),
		ClientSet:           clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:    clusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet: clusteroperationv1alpha1fake.NewSimpleClientset(),
	}
	cluster := &clusterv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my_kubean_cluster",
		},
	}
	clusterOps := &clusteroperationv1alpha1.ClusterOperation{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my_kubean_ops_cluster",
		},
	}
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "my_namespace",
			Name:      "my_name_secret_1",
		},
		Data: map[string][]byte{"ok": []byte(base64.StdEncoding.EncodeToString([]byte("ok123")))},
	}
	cmHosts := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "my_namespace",
			Name:      "my_name_cm_1",
		},
		Data: map[string]string{"ok": "ok123"},
	}
	cmVars := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "my_namespace",
			Name:      "my_name_cm_2",
		},
		Data: map[string]string{"ok": "ok123"},
	}
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "cluster.hostConf empty",
			args: func() bool {
				err := controller.CheckClusterDataRef(cluster, clusterOps)
				return err != nil && strings.Contains(err.Error(), "hostsConfRef is empty")
			},
			want: true,
		},
		{
			name: "cluster.hostConf not exist",
			args: func() bool {
				cluster.Spec.HostsConfRef = &apis.ConfigMapRef{
					NameSpace: cmHosts.Namespace,
					Name:      cmHosts.Name,
				}
				err := controller.CheckClusterDataRef(cluster, clusterOps)
				return err != nil && strings.Contains(err.Error(), "hostsConfRef my_namespace,my_name_cm_1 not found")
			},
			want: true,
		},
		{
			name: "cluster.varsConf empty",
			args: func() bool {
				controller.ClientSet.CoreV1().ConfigMaps(cmHosts.Namespace).Create(context.Background(), cmHosts, metav1.CreateOptions{})
				err := controller.CheckClusterDataRef(cluster, clusterOps)
				return err != nil && strings.Contains(err.Error(), "varsConfRef is empty")
			},
			want: true,
		},
		{
			name: "cluster.varsConf not exist",
			args: func() bool {
				cluster.Spec.VarsConfRef = &apis.ConfigMapRef{
					NameSpace: cmVars.Namespace,
					Name:      cmVars.Name,
				}
				err := controller.CheckClusterDataRef(cluster, clusterOps)
				fmt.Println(err)
				return err != nil && strings.Contains(err.Error(), "varsConfRef my_namespace,my_name_cm_2 not found")
			},
			want: true,
		},
		{
			name: "cluster.SSHAuthRef not exist",
			args: func() bool {
				controller.ClientSet.CoreV1().ConfigMaps(cmVars.Namespace).Create(context.Background(), cmVars, metav1.CreateOptions{})
				cluster.Spec.SSHAuthRef = &apis.SecretRef{
					NameSpace: secret.Namespace,
					Name:      secret.Name,
				}
				err := controller.CheckClusterDataRef(cluster, clusterOps)
				fmt.Println(err)
				return err != nil && strings.Contains(err.Error(), "sshAuthRef my_namespace,my_name_secret_1 not found")
			},
			want: true,
		},
		{
			name: "ok",
			args: func() bool {
				controller.ClientSet.CoreV1().Secrets(secret.Namespace).Create(context.Background(), secret, metav1.CreateOptions{})
				err := controller.CheckClusterDataRef(cluster, clusterOps)
				return err == nil
			},
			want: true,
		},
		{
			name: "multi namespace",
			args: func() bool {
				secret.Namespace = "other_namespace"
				cluster.Spec.SSHAuthRef = &apis.SecretRef{
					NameSpace: secret.Namespace,
					Name:      secret.Name,
				}
				controller.ClientSet.CoreV1().Secrets(secret.Namespace).Create(context.Background(), secret, metav1.CreateOptions{})
				err := controller.CheckClusterDataRef(cluster, clusterOps)
				return err != nil && strings.Contains(err.Error(), "hostsConfRef varsConfRef or sshAuthRef not in the same namespace")
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

func Test_GetKuBeanCluster(t *testing.T) {
	controller := Controller{
		Client:              newFakeClient(),
		ClientSet:           clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:    clusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet: clusteroperationv1alpha1fake.NewSimpleClientset(),
	}
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "no data",
			args: func() bool {
				data, err := controller.GetKuBeanCluster(&clusteroperationv1alpha1.ClusterOperation{Spec: clusteroperationv1alpha1.Spec{Cluster: "cluster1"}})
				return data == nil && err != nil
			},
			want: true,
		},
		{
			name: "data exists",
			args: func() bool {
				cluster := &clusterv1alpha1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster2",
					},
				}
				controller.KubeanClusterSet.KubeanV1alpha1().Clusters().Create(context.Background(), cluster, metav1.CreateOptions{})
				data, _ := controller.GetKuBeanCluster(&clusteroperationv1alpha1.ClusterOperation{Spec: clusteroperationv1alpha1.Spec{Cluster: "cluster2"}})
				return data != nil && data.Name == "cluster2"
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

func Test_CreateEntryPointShellConfigMap(t *testing.T) {
	controller := Controller{
		Client:              newFakeClient(),
		ClientSet:           clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:    clusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet: clusteroperationv1alpha1fake.NewSimpleClientset(),
	}
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "EntrypointSHRef not empty",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
				clusterOps.Spec.EntrypointSHRef = &apis.ConfigMapRef{
					NameSpace: "a",
					Name:      "b",
				}
				result, err := controller.CreateEntryPointShellConfigMap(clusterOps)
				return !result && err == nil
			},
			want: true,
		},
		{
			name: "generate new EntrypointSHRef",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
				clusterOps.Name = "cluster1"
				clusterOps.Spec.Action = "ping.yml"
				clusterOps.Spec.ActionType = "playbook"
				clusterOps.Spec.PreHook = []clusteroperationv1alpha1.HookAction{
					{ActionType: "playbook", Action: "update-hosts.yml", ExtraArgs: "-e ip=\"111.221.33.441123\" -e host=\"my-1.host.jh.com\""},
				}
				clusterOps.Spec.HostsConfRef = &apis.ConfigMapRef{
					NameSpace: "a",
					Name:      "abc",
				}
				controller.Client.Create(context.Background(), clusterOps)
				controller.KubeanClusterOpsSet.KubeanV1alpha1().ClusterOperations().Create(context.Background(), clusterOps, metav1.CreateOptions{})
				result, err := controller.CreateEntryPointShellConfigMap(clusterOps)
				return result && err == nil
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

func Test_FetchJobStatus(t *testing.T) {
	controller := Controller{
		Client:              newFakeClient(),
		ClientSet:           clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:    clusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet: clusteroperationv1alpha1fake.NewSimpleClientset(),
	}
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "jobRef is empty",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
				_, err := controller.FetchJobStatus(clusterOps)
				return err != nil
			},
			want: true,
		},
		{
			name: "job is not found",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
				clusterOps.Status.JobRef = &apis.JobRef{NameSpace: "kubean-system", Name: "job123"}
				status, err := controller.FetchJobStatus(clusterOps)
				return err == nil && status == clusteroperationv1alpha1.FailedStatus
			},
			want: true,
		},
		{
			name: "job finished",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
				clusterOps.Status.JobRef = &apis.JobRef{NameSpace: "kubean-system", Name: "job-finished"}
				targetJob := &batchv1.Job{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "batch/v1",
						Kind:       "Job",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "job-finished",
					},
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type:   batchv1.JobComplete,
								Status: corev1.ConditionTrue,
							},
						},
					},
				}
				controller.ClientSet.BatchV1().Jobs("kubean-system").Create(context.Background(), targetJob, metav1.CreateOptions{})
				status, err := controller.FetchJobStatus(clusterOps)
				return err == nil && status == clusteroperationv1alpha1.SucceededStatus
			},
			want: true,
		},
		{
			name: "job failed",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
				clusterOps.Status.JobRef = &apis.JobRef{NameSpace: "kubean-system", Name: "job-failed"}
				targetJob := &batchv1.Job{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "batch/v1",
						Kind:       "Job",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "job-failed",
					},
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type:   batchv1.JobFailed,
								Status: corev1.ConditionTrue,
							},
						},
					},
				}
				controller.ClientSet.BatchV1().Jobs("kubean-system").Create(context.Background(), targetJob, metav1.CreateOptions{})
				status, err := controller.FetchJobStatus(clusterOps)
				return err == nil && status == clusteroperationv1alpha1.FailedStatus
			},
			want: true,
		},
		{
			name: "job running",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
				clusterOps.Status.JobRef = &apis.JobRef{NameSpace: "kubean-system", Name: "job-running"}
				targetJob := &batchv1.Job{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "batch/v1",
						Kind:       "Job",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "job-running",
					},
				}
				controller.ClientSet.BatchV1().Jobs("kubean-system").Create(context.Background(), targetJob, metav1.CreateOptions{})
				status, err := controller.FetchJobStatus(clusterOps)
				return err == nil && status == clusteroperationv1alpha1.RunningStatus
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

func Test_ListClusterOps(t *testing.T) {
	controller := Controller{
		Client:              newFakeClient(),
		ClientSet:           clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:    clusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet: clusteroperationv1alpha1fake.NewSimpleClientset(),
	}
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "list and get nothing",
			args: func() bool {
				result, err := controller.ListClusterOps("kubeanCluster-123")
				return len(result) == 0 && err == nil
			},
			want: true,
		},
		{
			name: "list and get something",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   "my_kubean_ops_cluster",
						Labels: map[string]string{constants.KubeanClusterLabelKey: "kubeanCluster-123"},
					},
				}
				controller.KubeanClusterOpsSet.KubeanV1alpha1().ClusterOperations().Create(context.Background(), clusterOps, metav1.CreateOptions{})
				result, err := controller.ListClusterOps("kubeanCluster-123")
				return len(result) > 0 && err == nil
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

func Test_CopyConfigMap(t *testing.T) {
	controller := Controller{
		Client:              newFakeClient(),
		ClientSet:           clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:    clusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet: clusteroperationv1alpha1fake.NewSimpleClientset(),
	}
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "origin configMap is not found",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   "my_kubean_ops_cluster",
						Labels: map[string]string{constants.KubeanClusterLabelKey: "kubeanCluster-123"},
					},
				}
				_, err := controller.CopyConfigMap(clusterOps, &apis.ConfigMapRef{}, "")
				return err != nil && apierrors.IsNotFound(err)
			},
			want: true,
		},
		{
			name: "origin configMap is copied successfully",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   "my_kubean_ops_cluster",
						Labels: map[string]string{constants.KubeanClusterLabelKey: "kubeanCluster-123"},
					},
				}
				configMap := &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "ConfigMap",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "a-configmap",
					},
				}
				controller.ClientSet.CoreV1().ConfigMaps("kubean-system").Create(context.Background(), configMap, metav1.CreateOptions{})
				result, _ := controller.CopyConfigMap(clusterOps, &apis.ConfigMapRef{NameSpace: "kubean-system", Name: "a-configmap"}, "b-configmap")
				return result.Name == "b-configmap"
			},
			want: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.want != test.args() {
				t.Fatal()
			}
		})
	}
}

func Test_CopySecret(t *testing.T) {
	controller := Controller{
		Client:              newFakeClient(),
		ClientSet:           clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:    clusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet: clusteroperationv1alpha1fake.NewSimpleClientset(),
	}
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "origin secret is not found",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   "my_kubean_ops_cluster",
						Labels: map[string]string{constants.KubeanClusterLabelKey: "kubeanCluster-123"},
					},
				}
				_, err := controller.CopySecret(clusterOps, &apis.SecretRef{}, "")
				return err != nil && apierrors.IsNotFound(err)
			},
			want: true,
		},
		{
			name: "origin secret is copied successfully",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   "my_kubean_ops_cluster",
						Labels: map[string]string{constants.KubeanClusterLabelKey: "kubeanCluster-123"},
					},
				}
				secret := &corev1.Secret{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Secret",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "a-secret",
					},
				}
				controller.ClientSet.CoreV1().Secrets("kubean-system").Create(context.Background(), secret, metav1.CreateOptions{})
				result, _ := controller.CopySecret(clusterOps, &apis.SecretRef{NameSpace: "kubean-system", Name: "a-secret"}, "b-secret")
				return result.Name == "b-secret"
			},
			want: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.want != test.args() {
				t.Fatal()
			}
		})
	}
}

func Test_BackUpDataRef(t *testing.T) {
	controller := Controller{
		Client:              newFakeClient(),
		ClientSet:           clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:    clusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet: clusteroperationv1alpha1fake.NewSimpleClientset(),
	}
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "hostsConf is empty",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   "cluster1-ops",
						Labels: map[string]string{constants.KubeanClusterLabelKey: "cluster1"},
					},
				}
				cluster := &clusterv1alpha1.Cluster{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Cluster",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster1",
					},
				}
				_, err := controller.BackUpDataRef(clusterOps, cluster)
				return err != nil && strings.Contains(err.Error(), "DataRef has empty value")
			},
			want: true,
		},
		{
			name: "hostsConf and varsConf are not empty",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   "cluster1-ops",
						Labels: map[string]string{constants.KubeanClusterLabelKey: "cluster1"},
					},
					Spec: clusteroperationv1alpha1.Spec{},
				}
				cluster := &clusterv1alpha1.Cluster{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Cluster",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster1",
					},
					Spec: clusterv1alpha1.Spec{
						HostsConfRef: &apis.ConfigMapRef{NameSpace: "kubean-system", Name: "hosts-a"},
						VarsConfRef:  &apis.ConfigMapRef{NameSpace: "kubean-system", Name: "vars-a"},
					},
				}
				hostsConfigMap := &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "ConfigMap",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "hosts-a",
					},
				}
				varsConfigMap := &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "ConfigMap",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "vars-a",
					},
				}
				controller.ClientSet.CoreV1().ConfigMaps("kubean-system").Create(context.Background(), hostsConfigMap, metav1.CreateOptions{})
				controller.ClientSet.CoreV1().ConfigMaps("kubean-system").Create(context.Background(), varsConfigMap, metav1.CreateOptions{})
				controller.Client.Create(context.Background(), clusterOps)
				_, err := controller.BackUpDataRef(clusterOps, cluster)
				return err == nil
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

func Test_ProcessKubeanOperationImage(t *testing.T) {
	controller := Controller{
		Client:                newFakeClient(),
		ClientSet:             clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:      clusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet:   clusteroperationv1alpha1fake.NewSimpleClientset(),
		InfoManifestClientSet: manifestv1alpha1fake.NewSimpleClientset(),
	}
	tests := []struct {
		name string
		args func() string
		want string
	}{
		{
			name: "nothing to update",
			args: func() string {
				return controller.ProcessKubeanOperationImage("a:c", "my_tag")
			},
			want: "a:c",
		},
		{
			name: "update image tag",
			args: func() string {
				return controller.ProcessKubeanOperationImage("abc", "my_tag")
			},
			want: "abc:my_tag",
		},
		{
			name: "update image tag with latest",
			args: func() string {
				return controller.ProcessKubeanOperationImage("abc", "")
			},
			want: "abc:latest",
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

func Test_FetchGlobalManifestImageTag(t *testing.T) {
	controller := Controller{
		Client:                newFakeClient(),
		ClientSet:             clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:      clusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet:   clusteroperationv1alpha1fake.NewSimpleClientset(),
		InfoManifestClientSet: manifestv1alpha1fake.NewSimpleClientset(),
	}

	tests := []struct {
		name string
		args func() string
		want string
	}{
		{
			name: "none globalManifest",
			args: func() string {
				return controller.FetchGlobalManifestImageTag()
			},
			want: "",
		},
		{
			name: "none globalManifest",
			args: func() string {
				global := &manifestv1alpha1.Manifest{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Manifest",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   constants.InfoManifestGlobal,
						Labels: map[string]string{"origin": "v2"},
					},
					Spec: manifestv1alpha1.Spec{
						KubeanVersion: "123",
					},
				}
				controller.InfoManifestClientSet.KubeanV1alpha1().Manifests().Create(context.Background(), global, metav1.CreateOptions{})
				return controller.FetchGlobalManifestImageTag()
			},
			want: "123",
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
