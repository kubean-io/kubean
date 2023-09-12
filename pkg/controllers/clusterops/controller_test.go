// Copyright 2023 Authors of kubean-io
// SPDX-License-Identifier: Apache-2.0

package clusterops

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"

	"github.com/kubean-io/kubean-api/apis"
	clusterv1alpha1 "github.com/kubean-io/kubean-api/apis/cluster/v1alpha1"
	clusteroperationv1alpha1 "github.com/kubean-io/kubean-api/apis/clusteroperation/v1alpha1"
	localartifactsetv1alpha1 "github.com/kubean-io/kubean-api/apis/localartifactset/v1alpha1"
	manifestv1alpha1 "github.com/kubean-io/kubean-api/apis/manifest/v1alpha1"
	"github.com/kubean-io/kubean-api/constants"
	clusterv1alpha1fake "github.com/kubean-io/kubean-api/generated/cluster/clientset/versioned/fake"
	clusteroperationv1alpha1fake "github.com/kubean-io/kubean-api/generated/clusteroperation/clientset/versioned/fake"
	manifestv1alpha1fake "github.com/kubean-io/kubean-api/generated/manifest/clientset/versioned/fake"
	"github.com/kubean-io/kubean/pkg/util"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientsetfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	k8stesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
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

func fetchTestingFake(obj interface{ RESTClient() rest.Interface }) *k8stesting.Fake {
	// https://stackoverflow.com/questions/69740891/mocking-errors-with-client-go-fake-client
	return reflect.Indirect(reflect.ValueOf(obj)).FieldByName("Fake").Interface().(*k8stesting.Fake)
}

func removeReactorFromTestingTake(obj interface{ RESTClient() rest.Interface }, verb, resource string) {
	if fakeObj := fetchTestingFake(obj); fakeObj != nil {
		newReactionChain := make([]k8stesting.Reactor, 0)
		fakeObj.Lock()
		defer fakeObj.Unlock()
		for i := range fakeObj.ReactionChain {
			reaction := fakeObj.ReactionChain[i]
			if simpleReaction, ok := reaction.(*k8stesting.SimpleReactor); ok && simpleReaction.Verb == verb && simpleReaction.Resource == resource {
				continue // ignore
			}
			newReactionChain = append(newReactionChain, reaction)
		}
		fakeObj.ReactionChain = newReactionChain
	}
}

func TestUpdateStatusLoop(t *testing.T) {
	controller := Controller{}
	controller.Client = newFakeClient()
	ops := clusteroperationv1alpha1.ClusterOperation{}
	ops.ObjectMeta.Name = "clusteropsname"
	controller.Client.Create(context.Background(), &ops)
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
				needRequeue, err := controller.UpdateStatusLoop(ops, func(ops *clusteroperationv1alpha1.ClusterOperation) (clusteroperationv1alpha1.OpsStatus, *metav1.Time, error) {
					return "", nil, fmt.Errorf("one error")
				})
				return err != nil && err.Error() == "one error" && !needRequeue
			},
			want: true,
		},
		{
			name: "the status is Running and the result of fetchJobStatus is still Running",
			args: func(ops *clusteroperationv1alpha1.ClusterOperation) bool {
				ops.Status.Status = clusteroperationv1alpha1.RunningStatus
				needRequeue, err := controller.UpdateStatusLoop(ops, func(ops *clusteroperationv1alpha1.ClusterOperation) (clusteroperationv1alpha1.OpsStatus, *metav1.Time, error) {
					return clusteroperationv1alpha1.RunningStatus, nil, nil
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
				needRequeue, err := controller.UpdateStatusLoop(ops, func(ops *clusteroperationv1alpha1.ClusterOperation) (clusteroperationv1alpha1.OpsStatus, *metav1.Time, error) {
					return clusteroperationv1alpha1.SucceededStatus, &metav1.Time{}, nil
				})
				resultOps := &clusteroperationv1alpha1.ClusterOperation{}
				controller.Client.Get(context.Background(), client.ObjectKey{Name: "clusteropsname"}, resultOps)
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
				controller.Client.Get(context.Background(), client.ObjectKey{Name: "clusteropsname"}, ops)
				ops.Status.Status = clusteroperationv1alpha1.RunningStatus
				ops.Status.EndTime = nil
				needRequeue, err := controller.UpdateStatusLoop(ops, func(ops *clusteroperationv1alpha1.ClusterOperation) (clusteroperationv1alpha1.OpsStatus, *metav1.Time, error) {
					return clusteroperationv1alpha1.FailedStatus, nil, nil
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
				ops.Spec.Image = "abc"
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
	os.Setenv("POD_NAMESPACE", "mynamespace")
	controller := Controller{
		Client:                newFakeClient(),
		ClientSet:             clientsetfake.NewSimpleClientset(),
		InfoManifestClientSet: manifestv1alpha1fake.NewSimpleClientset(),
	}
	clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
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
				job := controller.NewKubesprayJob(clusterOps, "kubean")
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
				job := controller.NewKubesprayJob(clusterOps, "kubean")
				return job.Namespace == "mynamespace" && job.Name == "kubean-myops-job" && len(job.Spec.Template.Spec.Containers) == 1 && len(job.Spec.Template.Spec.Containers[0].VolumeMounts) == 4 && len(job.Spec.Template.Spec.Volumes) == 4
			},
			want: true,
		},
		{
			name: "activeDeadlineSeconds args",
			args: func() bool {
				ActiveDeadlineSeconds := int64(10)
				clusterOps.Spec.ActiveDeadlineSeconds = &ActiveDeadlineSeconds
				job := controller.NewKubesprayJob(clusterOps, "kubean")
				return *job.Spec.ActiveDeadlineSeconds == 10
			},
			want: true,
		},
		{
			name: "nil activeDeadlineSeconds args",
			args: func() bool {
				clusterOps.Spec.ActiveDeadlineSeconds = nil
				job := controller.NewKubesprayJob(clusterOps, "kubean")
				return job.Spec.ActiveDeadlineSeconds == nil
			},
			want: true,
		},
		{
			name: "not empty Resources",
			args: func() bool {
				clusterOps.Spec.Resources = corev1.ResourceRequirements{Limits: map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("1m")}}
				job := controller.NewKubesprayJob(clusterOps, "kubean")
				return len(job.Spec.Template.Spec.Containers[0].Resources.Limits) != 0
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

func TestController_HookCustomAction(t *testing.T) {
	os.Setenv("POD_NAMESPACE", "mynamespace")
	builtinActionSource := clusteroperationv1alpha1.BuiltinActionSource
	configmapActionSource := clusteroperationv1alpha1.ConfigMapActionSource
	controller := Controller{
		Client:                newFakeClient(),
		ClientSet:             clientsetfake.NewSimpleClientset(),
		InfoManifestClientSet: manifestv1alpha1fake.NewSimpleClientset(),
	}
	clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
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
		name      string
		setAction func()
		wantErr   bool
	}{
		{
			name: "specified configmap as actionSource, but actionSourceRef empty",
			setAction: func() {
				clusterOps.Spec.PreHook = nil
				clusterOps.Spec.PostHook = nil

				clusterOps.Spec.ActionSource = &configmapActionSource
			},
			wantErr: true,
		},
		{
			name: "specified configmap as actionSource with not exist actionSourceRef",
			setAction: func() {
				clusterOps.Spec.PreHook = nil
				clusterOps.Spec.PostHook = nil

				clusterOps.Spec.ActionSource = &configmapActionSource
				clusterOps.Spec.ActionSourceRef = &apis.ConfigMapRef{
					Name:      "myplaybook",
					NameSpace: "mynamespace-not-exist",
				}
			},
			wantErr: true,
		},
		{
			name: "specified configmap as actionSource with actionSourceRef",
			setAction: func() {
				clusterOps.Spec.PreHook = nil
				clusterOps.Spec.PostHook = nil

				clusterOps.Spec.ActionSource = &configmapActionSource
				clusterOps.Spec.ActionSourceRef = &apis.ConfigMapRef{
					Name:      "myplaybook",
					NameSpace: "mynamespace",
				}
			},
			wantErr: false,
		},
		{
			name: "run with preHook and postHook in built-in",
			setAction: func() {
				clusterOps.Spec.PreHook = nil
				clusterOps.Spec.PostHook = nil
				clusterOps.Spec.PreHook = []clusteroperationv1alpha1.HookAction{{ActionSource: &builtinActionSource, ActionType: clusteroperationv1alpha1.PlaybookActionType, Action: "ping.yml"}}
				clusterOps.Spec.PostHook = []clusteroperationv1alpha1.HookAction{{ActionSource: &builtinActionSource, ActionType: clusteroperationv1alpha1.PlaybookActionType, Action: "ping.yml"}}
			},
			wantErr: false,
		},
		{
			name: "wrong empty preHook ActionSourceRef",
			setAction: func() {
				clusterOps.Spec.PreHook = nil
				clusterOps.Spec.PostHook = nil
				clusterOps.Spec.PreHook = []clusteroperationv1alpha1.HookAction{{ActionSource: &configmapActionSource, ActionSourceRef: nil}}
			},
			wantErr: true,
		},
		{
			name: "wrong preHook with not exist ActionSourceRef",
			setAction: func() {
				clusterOps.Spec.PreHook = nil
				clusterOps.Spec.PostHook = nil
				clusterOps.Spec.PreHook = []clusteroperationv1alpha1.HookAction{{ActionSource: &configmapActionSource, ActionSourceRef: &apis.ConfigMapRef{NameSpace: "abc", Name: "ok"}}}
			},
			wantErr: true,
		},
		{
			name: "wrong empty postHook ActionSourceRef",
			setAction: func() {
				clusterOps.Spec.PreHook = nil
				clusterOps.Spec.PostHook = nil
				clusterOps.Spec.PostHook = []clusteroperationv1alpha1.HookAction{{ActionSource: &configmapActionSource, ActionSourceRef: nil}}
			},
			wantErr: true,
		},
		{
			name: "wrong postHook with not exist ActionSourceRef",
			setAction: func() {
				clusterOps.Spec.PreHook = nil
				clusterOps.Spec.PostHook = nil
				clusterOps.Spec.PostHook = []clusteroperationv1alpha1.HookAction{{ActionSource: &configmapActionSource, ActionSourceRef: &apis.ConfigMapRef{NameSpace: "abc", Name: "ok"}}}
			},
			wantErr: true,
		},
		{
			name: "run with preHook and postHook in customer-action",
			setAction: func() {
				clusterOps.Spec.PreHook = nil
				clusterOps.Spec.PostHook = nil
				clusterOps.Spec.PreHook = []clusteroperationv1alpha1.HookAction{
					{
						ActionSource: &configmapActionSource,
						ActionSourceRef: &apis.ConfigMapRef{
							Name:      "myplaybook",
							NameSpace: "mynamespace",
						},
					},
				}
				clusterOps.Spec.PostHook = []clusteroperationv1alpha1.HookAction{
					{
						ActionSource: &configmapActionSource,
						ActionSourceRef: &apis.ConfigMapRef{
							Name:      "myplaybook",
							NameSpace: "mynamespace",
						},
					},
				}
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setAction()
			err := controller.HookCustomAction(clusterOps, controller.NewKubesprayJob(clusterOps, "kubean"))
			if (err != nil) != test.wantErr {
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

func TestGetServiceAccountName(t *testing.T) {
	controller := Controller{
		Client:                newFakeClient(),
		ClientSet:             clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:      clusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet:   clusteroperationv1alpha1fake.NewSimpleClientset(),
		InfoManifestClientSet: manifestv1alpha1fake.NewSimpleClientset(),
	}

	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "occurs error",
			args: func() bool {
				fetchTestingFake(controller.ClientSet.CoreV1()).PrependReactor("list", "serviceaccounts", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, fmt.Errorf("this is error")
				})
				_, err := controller.GetServiceAccountName("kubean-system", ServiceAccount)
				removeReactorFromTestingTake(controller.ClientSet.CoreV1(), "list", "serviceaccounts")
				return err != nil
			},
			want: true,
		},
		{
			name: "nothing to get",
			args: func() bool {
				result, _ := controller.GetServiceAccountName("kubean-system", ServiceAccount)
				return result == ""
			},
			want: true,
		},
		{
			name: "good result",
			args: func() bool {
				controller.ClientSet.CoreV1().ServiceAccounts("kubean-system").Create(context.Background(), &corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "sa1",
						Labels:    map[string]string{"kubean.io/kubean-operator": "sa"},
					},
				}, metav1.CreateOptions{})
				result, _ := controller.GetServiceAccountName("kubean-system", ServiceAccount)
				return result == "sa1"
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

func TestReconcile(t *testing.T) {
	os.Setenv("POD_NAMESPACE", "")
	genController := func() *Controller {
		return &Controller{
			Client:                newFakeClient(),
			ClientSet:             clientsetfake.NewSimpleClientset(),
			KubeanClusterSet:      clusterv1alpha1fake.NewSimpleClientset(),
			KubeanClusterOpsSet:   clusteroperationv1alpha1fake.NewSimpleClientset(),
			InfoManifestClientSet: manifestv1alpha1fake.NewSimpleClientset(),
		}
	}
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "clusterOps not found",
			args: func() bool {
				controller := genController()
				result, _ := controller.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: types.NamespacedName{Name: "cluster1"}})
				return result.Requeue
			},
			want: false,
		},
		{
			name: "clusterOps and cluster found successfully but not ValidImageName",
			args: func() bool {
				controller := genController()
				controller.ClientSet.CoreV1().ServiceAccounts("kubean-system").Create(context.Background(), &corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "sa1",
						Labels:    map[string]string{"kubean.io/kubean-operator": "sa"},
					},
				}, metav1.CreateOptions{})

				cluster := &clusterv1alpha1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my_kubean_cluster",
					},
					Spec: clusterv1alpha1.Spec{
						HostsConfRef: &apis.ConfigMapRef{
							Name:      "hosts-conf",
							NameSpace: "kubean-system",
						},
						VarsConfRef: &apis.ConfigMapRef{
							Name:      "vars-conf",
							NameSpace: "kubean-system",
						},
					},
				}
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   "my_kubean_ops_cluster",
						Labels: map[string]string{constants.KubeanClusterLabelKey: "my_kubean_cluster"},
					},
					Spec: clusteroperationv1alpha1.Spec{
						Cluster: "my_kubean_cluster",
						Image:   "myimagename:",
					},
				}
				controller.Client.Create(context.Background(), clusterOps)
				controller.KubeanClusterSet.KubeanV1alpha1().Clusters().Create(context.Background(), cluster, metav1.CreateOptions{})
				controller.KubeanClusterOpsSet.KubeanV1alpha1().ClusterOperations().Create(context.Background(), clusterOps, metav1.CreateOptions{})
				result, _ := controller.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: types.NamespacedName{Name: "my_kubean_ops_cluster"}})
				opsResult := &clusteroperationv1alpha1.ClusterOperation{}
				controller.Client.Get(context.Background(), types.NamespacedName{Name: "my_kubean_ops_cluster"}, opsResult)
				return result.Requeue == false && opsResult.Status.Status == clusteroperationv1alpha1.FailedStatus
			},
			want: true,
		},
		{
			name: "clusterOps and cluster found successfully but not hosts-conf and vars-conf data",
			args: func() bool {
				controller := genController()
				controller.ClientSet.CoreV1().ServiceAccounts("kubean-system").Create(context.Background(), &corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "sa1",
						Labels:    map[string]string{"kubean.io/kubean-operator": "sa"},
					},
				}, metav1.CreateOptions{})
				cluster := &clusterv1alpha1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my_kubean_cluster",
					},
					Spec: clusterv1alpha1.Spec{
						HostsConfRef: &apis.ConfigMapRef{
							Name:      "hosts-conf",
							NameSpace: "kubean-system",
						},
						VarsConfRef: &apis.ConfigMapRef{
							Name:      "vars-conf",
							NameSpace: "kubean-system",
						},
					},
				}
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   "my_kubean_ops_cluster",
						Labels: map[string]string{constants.KubeanClusterLabelKey: "my_kubean_cluster"},
					},
					Spec: clusteroperationv1alpha1.Spec{
						Cluster: "my_kubean_cluster",
						Image:   "myimagename",
					},
				}
				controller.Client.Create(context.Background(), clusterOps)
				controller.KubeanClusterSet.KubeanV1alpha1().Clusters().Create(context.Background(), cluster, metav1.CreateOptions{})
				controller.KubeanClusterOpsSet.KubeanV1alpha1().ClusterOperations().Create(context.Background(), clusterOps, metav1.CreateOptions{})
				result, _ := controller.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: types.NamespacedName{Name: "my_kubean_ops_cluster"}})
				opsResult := &clusteroperationv1alpha1.ClusterOperation{}
				controller.Client.Get(context.Background(), types.NamespacedName{Name: "my_kubean_ops_cluster"}, opsResult)
				return result.Requeue == false && opsResult.Status.Status == clusteroperationv1alpha1.FailedStatus
			},
			want: true,
		},
		{
			name: "clusterOps and cluster found with digest-updating",
			args: func() bool {
				controller := genController()
				controller.ClientSet.CoreV1().ServiceAccounts("kubean-system").Create(context.Background(), &corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "sa1",
						Labels:    map[string]string{"kubean.io/kubean-operator": "sa"},
					},
				}, metav1.CreateOptions{})
				controller.ClientSet.CoreV1().ConfigMaps("kubean-system").Create(context.Background(), &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "vars-conf",
					},
					Data: map[string]string{"ok": "ok123"},
				}, metav1.CreateOptions{})
				controller.ClientSet.CoreV1().ConfigMaps("kubean-system").Create(context.Background(), &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "hosts-conf",
					},
					Data: map[string]string{"ok": "ok123"},
				}, metav1.CreateOptions{})
				cluster := &clusterv1alpha1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my_kubean_cluster",
					},
					Spec: clusterv1alpha1.Spec{
						HostsConfRef: &apis.ConfigMapRef{
							Name:      "hosts-conf",
							NameSpace: "kubean-system",
						},
						VarsConfRef: &apis.ConfigMapRef{
							Name:      "vars-conf",
							NameSpace: "kubean-system",
						},
					},
				}
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   "my_kubean_ops_cluster",
						Labels: map[string]string{constants.KubeanClusterLabelKey: "my_kubean_cluster"},
					},
					Spec: clusteroperationv1alpha1.Spec{
						Cluster: "my_kubean_cluster",
						Image:   "myimagename",
					},
				}
				controller.Client.Create(context.Background(), clusterOps)
				controller.KubeanClusterSet.KubeanV1alpha1().Clusters().Create(context.Background(), cluster, metav1.CreateOptions{})
				controller.KubeanClusterOpsSet.KubeanV1alpha1().ClusterOperations().Create(context.Background(), clusterOps, metav1.CreateOptions{})
				result, _ := controller.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: types.NamespacedName{Name: "my_kubean_ops_cluster"}})
				opsResult := &clusteroperationv1alpha1.ClusterOperation{}
				controller.Client.Get(context.Background(), types.NamespacedName{Name: "my_kubean_ops_cluster"}, opsResult)
				return result.RequeueAfter > 0 && opsResult.Status.Digest != ""
			},
			want: true,
		},
		{
			name: "clusterOps and cluster found with wrong-actionType",
			args: func() bool {
				controller := genController()
				controller.ClientSet.CoreV1().ServiceAccounts("kubean-system").Create(context.Background(), &corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "sa1",
						Labels:    map[string]string{"kubean.io/kubean-operator": "sa"},
					},
				}, metav1.CreateOptions{})
				controller.ClientSet.CoreV1().ConfigMaps("kubean-system").Create(context.Background(), &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "vars-conf",
					},
					Data: map[string]string{"ok": "ok123"},
				}, metav1.CreateOptions{})
				controller.ClientSet.CoreV1().ConfigMaps("kubean-system").Create(context.Background(), &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "hosts-conf",
					},
					Data: map[string]string{"ok": "ok123"},
				}, metav1.CreateOptions{})
				cluster := &clusterv1alpha1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my_kubean_cluster",
					},
					Spec: clusterv1alpha1.Spec{
						HostsConfRef: &apis.ConfigMapRef{
							Name:      "hosts-conf",
							NameSpace: "kubean-system",
						},
						VarsConfRef: &apis.ConfigMapRef{
							Name:      "vars-conf",
							NameSpace: "kubean-system",
						},
					},
				}
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   "my_kubean_ops_cluster",
						Labels: map[string]string{constants.KubeanClusterLabelKey: "my_kubean_cluster"},
					},
					Spec: clusteroperationv1alpha1.Spec{
						Cluster:    "my_kubean_cluster",
						Image:      "myimagename",
						Action:     "ping.yml",
						ActionType: "error-type",
					},
				}
				controller.Client.Create(context.Background(), clusterOps)
				controller.KubeanClusterSet.KubeanV1alpha1().Clusters().Create(context.Background(), cluster, metav1.CreateOptions{})
				controller.KubeanClusterOpsSet.KubeanV1alpha1().ClusterOperations().Create(context.Background(), clusterOps, metav1.CreateOptions{})
				_, _ = controller.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: types.NamespacedName{Name: "my_kubean_ops_cluster"}})
				_, _ = controller.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: types.NamespacedName{Name: "my_kubean_ops_cluster"}})
				_, _ = controller.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: types.NamespacedName{Name: "my_kubean_ops_cluster"}})
				_, err := controller.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: types.NamespacedName{Name: "my_kubean_ops_cluster"}})
				opsResult := &clusteroperationv1alpha1.ClusterOperation{}
				controller.Client.Get(context.Background(), types.NamespacedName{Name: "my_kubean_ops_cluster"}, opsResult)
				return err != nil && opsResult.Status.Status == clusteroperationv1alpha1.FailedStatus
			},
			want: true,
		},
		{
			name: "clusterOps and cluster found and create job successfully",
			args: func() bool {
				controller := genController()
				_, err := controller.ClientSet.CoreV1().ServiceAccounts("default").Create(context.Background(), &corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "sa1",
						Labels:    map[string]string{"kubean.io/kubean-operator": "sa"},
					},
				}, metav1.CreateOptions{})
				if err != nil {
					t.Fatal(err)
				}
				_, err = controller.ClientSet.CoreV1().ServiceAccounts("kubean-system").Create(context.Background(), &corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "sa1",
						Labels:    map[string]string{"kubean.io/kubean-operator": "sa"},
					},
				}, metav1.CreateOptions{})
				if err != nil {
					t.Fatal(err)
				}
				time.Sleep(time.Second)
				controller.ClientSet.CoreV1().ConfigMaps("kubean-system").Create(context.Background(), &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "vars-conf",
					},
					Data: map[string]string{"ok": "ok123"},
				}, metav1.CreateOptions{})
				controller.ClientSet.CoreV1().ConfigMaps("kubean-system").Create(context.Background(), &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "hosts-conf",
					},
					Data: map[string]string{"ok": "ok123"},
				}, metav1.CreateOptions{})
				cluster := &clusterv1alpha1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my_kubean_cluster",
					},
					Spec: clusterv1alpha1.Spec{
						HostsConfRef: &apis.ConfigMapRef{
							Name:      "hosts-conf",
							NameSpace: "kubean-system",
						},
						VarsConfRef: &apis.ConfigMapRef{
							Name:      "vars-conf",
							NameSpace: "kubean-system",
						},
					},
				}
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   "my_kubean_ops_cluster",
						Labels: map[string]string{constants.KubeanClusterLabelKey: "my_kubean_cluster"},
					},
					Spec: clusteroperationv1alpha1.Spec{
						Cluster:    "my_kubean_cluster",
						Image:      "myimagename",
						Action:     "ping.yml",
						ActionType: "playbook",
					},
				}
				controller.Client.Create(context.Background(), clusterOps)
				controller.KubeanClusterSet.KubeanV1alpha1().Clusters().Create(context.Background(), cluster, metav1.CreateOptions{})
				controller.KubeanClusterOpsSet.KubeanV1alpha1().ClusterOperations().Create(context.Background(), clusterOps, metav1.CreateOptions{})
				_, _ = controller.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: types.NamespacedName{Name: "my_kubean_ops_cluster"}})
				_, _ = controller.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: types.NamespacedName{Name: "my_kubean_ops_cluster"}})
				_, _ = controller.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: types.NamespacedName{Name: "my_kubean_ops_cluster"}})
				_, _ = controller.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: types.NamespacedName{Name: "my_kubean_ops_cluster"}})
				_, _ = controller.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: types.NamespacedName{Name: "my_kubean_ops_cluster"}})
				_, _ = controller.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: types.NamespacedName{Name: "my_kubean_ops_cluster"}})
				result, err := controller.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: types.NamespacedName{Name: "my_kubean_ops_cluster"}})
				opsResult := &clusteroperationv1alpha1.ClusterOperation{}
				controller.Client.Get(context.Background(), types.NamespacedName{Name: "my_kubean_ops_cluster"}, opsResult)
				return err == nil && result.RequeueAfter == LoopForJobStatus
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

func TestCreateKubeSprayJob(t *testing.T) {
	controller := &Controller{
		Client:                newFakeClient(),
		ClientSet:             clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:      clusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet:   clusteroperationv1alpha1fake.NewSimpleClientset(),
		InfoManifestClientSet: manifestv1alpha1fake.NewSimpleClientset(),
	}
	clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
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
	clusterOps.Spec.SSHAuthRef = &apis.SecretRef{
		NameSpace: "mynamespace",
		Name:      "ssh-auth-ref",
	}
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "get sa but error",
			args: func() bool {
				controller.ClientSet.CoreV1().ServiceAccounts("default").Create(context.Background(), &corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "sa1",
						Labels:    map[string]string{"kubean.io/kubean-operator": "sa"},
					},
				}, metav1.CreateOptions{})
				clusterOps1 := *clusterOps
				fetchTestingFake(controller.ClientSet.CoreV1()).PrependReactor("list", "serviceaccounts", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, fmt.Errorf("this is error")
				})
				needRequeue, err := controller.CreateKubeSprayJob(&clusterOps1)
				removeReactorFromTestingTake(controller.ClientSet.CoreV1(), "list", "serviceaccounts")
				return !needRequeue && err != nil && err.Error() == "this is error"
			},
			want: true,
		},
		{
			name: "create job but error",
			args: func() bool {
				controller.ClientSet.CoreV1().ServiceAccounts("default").Create(context.Background(), &corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "sa1",
						Labels:    map[string]string{"kubean.io/kubean-operator": "sa"},
					},
				}, metav1.CreateOptions{})
				clusterOps1 := *clusterOps
				fetchTestingFake(controller.ClientSet.CoreV1()).PrependReactor("create", "jobs", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, fmt.Errorf("this is error when create job")
				})
				needRequeue, err := controller.CreateKubeSprayJob(&clusterOps1)
				removeReactorFromTestingTake(controller.ClientSet.CoreV1(), "create", "jobs")
				return !needRequeue && err != nil && err.Error() == "this is error when create job"
			},
			want: true,
		},
		{
			name: "try to get job but other error",
			args: func() bool {
				controller.ClientSet.CoreV1().ServiceAccounts("default").Create(context.Background(), &corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "sa1",
						Labels:    map[string]string{"kubean.io/kubean-operator": "sa"},
					},
				}, metav1.CreateOptions{})

				fetchTestingFake(controller.ClientSet.CoreV1()).PrependReactor("get", "jobs", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, fmt.Errorf("this is error")
				})
				clusterOps1 := *clusterOps
				needRequeue, err := controller.CreateKubeSprayJob(&clusterOps1)
				removeReactorFromTestingTake(controller.ClientSet.CoreV1(), "get", "jobs")
				return !needRequeue && err != nil && err.Error() == "this is error"
			},
			want: true,
		},
		{
			name: "create job successfully",
			args: func() bool {
				controller.ClientSet.CoreV1().ServiceAccounts("default").Create(context.Background(), &corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "sa1",
						Labels:    map[string]string{"kubean.io/kubean-operator": "sa"},
					},
				}, metav1.CreateOptions{})

				needRequeue, err := controller.CreateKubeSprayJob(clusterOps)
				return needRequeue && err == nil
			},
		},
		{
			name: "JobRef not empty",
			args: func() bool {
				controller.ClientSet.CoreV1().ServiceAccounts("default").Create(context.Background(), &corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "sa1",
						Labels:    map[string]string{"kubean.io/kubean-operator": "sa"},
					},
				}, metav1.CreateOptions{})
				clusterOps1 := *clusterOps
				clusterOps1.Status.JobRef = &apis.JobRef{NameSpace: "abc", Name: "abc"}
				needRequeue, err := controller.CreateKubeSprayJob(&clusterOps1)
				return !needRequeue && err == nil
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

func Test_TrySuspendPod(t *testing.T) {
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
			name: "job finished",
			args: func() bool {
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
				result, err := controller.TrySuspendPod(&clusteroperationv1alpha1.ClusterOperation{
					Spec: clusteroperationv1alpha1.Spec{
						Cluster: "cluster",
					},
					Status: clusteroperationv1alpha1.Status{
						JobRef: &apis.JobRef{
							NameSpace: "kubean-system",
							Name:      "job-finished",
						},
					},
				})
				return result == false && err == nil
			},
			want: true,
		},
		{
			name: "job not found",
			args: func() bool {
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
				result, err := controller.TrySuspendPod(&clusteroperationv1alpha1.ClusterOperation{
					Spec: clusteroperationv1alpha1.Spec{
						Cluster: "cluster",
					},
					Status: clusteroperationv1alpha1.Status{
						JobRef: &apis.JobRef{
							NameSpace: "kubean-system",
							Name:      "job-not-found",
						},
					},
				})
				return result == false && err == nil
			},
			want: true,
		},
		{
			name: "cluster ops no annotations",
			args: func() bool {
				controller.ClientSet.CoreV1().Pods("kubean-system").Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "kubean-system", Labels: map[string]string{"a": "b"}},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				}, metav1.CreateOptions{})
				targetJob := &batchv1.Job{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "batch/v1",
						Kind:       "Job",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "job-running",
					},
					Spec: batchv1.JobSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"a": "b"},
						},
					},
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type:   batchv1.JobComplete,
								Status: corev1.ConditionFalse,
							},
						},
					},
				}
				controller.ClientSet.BatchV1().Jobs("kubean-system").Create(context.Background(), targetJob, metav1.CreateOptions{})
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my_kubean_ops_cluster",
					},
					Spec: clusteroperationv1alpha1.Spec{
						Cluster: "cluster",
					},
					Status: clusteroperationv1alpha1.Status{
						JobRef: &apis.JobRef{
							NameSpace: "kubean-system",
							Name:      "job-running",
						},
					},
				}
				controller.Client.Create(context.Background(), clusterOps)
				result, err := controller.TrySuspendPod(clusterOps)
				return result == true && err == nil && clusterOps.Annotations[JobActorPodAnnoKey] == "pod1"
			},
			want: true,
		},
		{
			name: "suspend job success",
			args: func() bool {
				controller.ClientSet.CoreV1().Pods("kubean-system").Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "kubean-system", Labels: map[string]string{"a": "b"}},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				}, metav1.CreateOptions{})
				targetJob := &batchv1.Job{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "batch/v1",
						Kind:       "Job",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "job-running",
					},
					Spec: batchv1.JobSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"a": "b"},
						},
					},
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type:   batchv1.JobComplete,
								Status: corev1.ConditionFalse,
							},
						},
					},
				}
				controller.ClientSet.BatchV1().Jobs("kubean-system").Create(context.Background(), targetJob, metav1.CreateOptions{})

				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "my_kubean_ops_cluster",
						Annotations: map[string]string{JobActorPodAnnoKey: "key"},
					},
					Spec: clusteroperationv1alpha1.Spec{
						Cluster: "cluster",
					},
					Status: clusteroperationv1alpha1.Status{
						JobRef: &apis.JobRef{
							NameSpace: "kubean-system",
							Name:      "job-running",
						},
					},
				}
				controller.Client.Create(context.Background(), clusterOps)
				result, err := controller.TrySuspendPod(clusterOps)
				return result == true && err == nil
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

func Test_GetRunningPodFromJob(t *testing.T) {
	genController := func() Controller {
		return Controller{
			Client:              newFakeClient(),
			ClientSet:           clientsetfake.NewSimpleClientset(),
			KubeanClusterSet:    clusterv1alpha1fake.NewSimpleClientset(),
			KubeanClusterOpsSet: clusteroperationv1alpha1fake.NewSimpleClientset(),
		}
	}
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "",
			args: func() bool {
				controller := genController()
				controller.ClientSet.CoreV1().Pods("job1-namespace").Create(context.Background(), &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "job1-namespace", Labels: map[string]string{"a": "b"}},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				}, metav1.CreateOptions{})
				targetPod, err := controller.GetRunningPodFromJob(&batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "job1", Namespace: "job1-namespace"}, Spec: batchv1.JobSpec{Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"a": "b"}}}})
				return targetPod.Name == "pod1" && err == nil
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
	genController := func() Controller {
		return Controller{
			Client:              newFakeClient(),
			ClientSet:           clientsetfake.NewSimpleClientset(),
			KubeanClusterSet:    clusterv1alpha1fake.NewSimpleClientset(),
			KubeanClusterOpsSet: clusteroperationv1alpha1fake.NewSimpleClientset(),
		}
	}
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "EntrypointSHRef not empty",
			args: func() bool {
				controller := genController()
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
				controller := genController()
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
		{
			name: "generate new EntrypointSHRef when entrypoint configmap exists",
			args: func() bool {
				controller := genController()
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
				newConfigMap := &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-entrypoint", clusterOps.Name),
						Namespace: util.GetCurrentNSOrDefault(),
					},
					Data: map[string]string{"entrypoint.sh": strings.TrimSpace("entrypoint_data")}, // |2+
				}
				if _, err := controller.ClientSet.CoreV1().ConfigMaps(newConfigMap.Namespace).Create(context.Background(), newConfigMap, metav1.CreateOptions{}); err != nil {
					t.Fatal(err)
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

func TestStart(t *testing.T) {
	controller := Controller{
		Client:              newFakeClient(),
		ClientSet:           clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:    clusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet: clusteroperationv1alpha1fake.NewSimpleClientset(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	controller.Start(ctx)
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
				_, _, err := controller.FetchJobConditionStatusAndCompletionTime(clusterOps)
				return err != nil
			},
			want: true,
		},
		{
			name: "job is not found",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
				clusterOps.Status.JobRef = &apis.JobRef{NameSpace: "kubean-system", Name: "job123"}
				status, _, err := controller.FetchJobConditionStatusAndCompletionTime(clusterOps)
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
				status, _, err := controller.FetchJobConditionStatusAndCompletionTime(clusterOps)
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
				status, _, err := controller.FetchJobConditionStatusAndCompletionTime(clusterOps)
				return err == nil && status == clusteroperationv1alpha1.FailedStatus
			},
			want: true,
		},
		{
			name: "job JobFailureTarget",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
				clusterOps.Status.JobRef = &apis.JobRef{NameSpace: "kubean-system", Name: "job-fail-target"}
				targetJob := &batchv1.Job{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "batch/v1",
						Kind:       "Job",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "job-fail-target",
					},
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type:   batchv1.JobFailureTarget,
								Status: corev1.ConditionTrue,
							},
						},
					},
				}
				controller.ClientSet.BatchV1().Jobs("kubean-system").Create(context.Background(), targetJob, metav1.CreateOptions{})
				status, _, err := controller.FetchJobConditionStatusAndCompletionTime(clusterOps)
				return err == nil && status == clusteroperationv1alpha1.FailedStatus
			},
			want: true,
		},
		{
			name: "job JobSuspended",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
				clusterOps.Status.JobRef = &apis.JobRef{NameSpace: "kubean-system", Name: "job-suspend"}
				targetJob := &batchv1.Job{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "batch/v1",
						Kind:       "Job",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "job-suspend",
					},
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type:   batchv1.JobSuspended,
								Status: corev1.ConditionTrue,
							},
						},
					},
				}
				controller.ClientSet.BatchV1().Jobs("kubean-system").Create(context.Background(), targetJob, metav1.CreateOptions{})
				status, _, err := controller.FetchJobConditionStatusAndCompletionTime(clusterOps)
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
				status, completionTime, err := controller.FetchJobConditionStatusAndCompletionTime(clusterOps)
				return err == nil && status == clusteroperationv1alpha1.RunningStatus && completionTime == nil
			},
			want: true,
		},
		{
			name: "get job but error",
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
				fetchTestingFake(controller.ClientSet.CoreV1()).PrependReactor("get", "jobs", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, fmt.Errorf("this is error")
				})
				_, _, err := controller.FetchJobConditionStatusAndCompletionTime(clusterOps)
				removeReactorFromTestingTake(controller.ClientSet.CoreV1(), "get", "jobs")
				return err != nil && err.Error() == "this is error"
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
				_, err := controller.CopyConfigMap(clusterOps, &apis.ConfigMapRef{}, "", "")
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
				result, _ := controller.CopyConfigMap(clusterOps, &apis.ConfigMapRef{NameSpace: "kubean-system", Name: "a-configmap"}, "b-configmap", "")
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
				_, err := controller.CopySecret(clusterOps, &apis.SecretRef{}, "", "")
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
				result, _ := controller.CopySecret(clusterOps, &apis.SecretRef{NameSpace: "kubean-system", Name: "a-secret"}, "b-secret", "")
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
			name: "hostsConf are not empty",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster1-ops",
						// Labels: map[string]string{constants.KubeanClusterLabelKey: "cluster1"},
					},
					Spec: clusteroperationv1alpha1.Spec{
						HostsConfRef: &apis.ConfigMapRef{NameSpace: "kubean-system", Name: "hosts-a-back-1"},
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
		{
			name: "varsConf are not empty",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   "cluster1-ops-2",
						Labels: map[string]string{constants.KubeanClusterLabelKey: "cluster1-1"},
					},
					Spec: clusteroperationv1alpha1.Spec{
						VarsConfRef: &apis.ConfigMapRef{
							NameSpace: "kubean-system", Name: "vars-a-back-1",
						},
					},
				}
				cluster := &clusterv1alpha1.Cluster{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Cluster",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster1-1",
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
		{
			name: "SSHAuth are not empty",
			args: func() bool {
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   "cluster1-ops-3",
						Labels: map[string]string{constants.KubeanClusterLabelKey: "cluster1-2"},
					},
					Spec: clusteroperationv1alpha1.Spec{
						HostsConfRef: &apis.ConfigMapRef{NameSpace: "kubean-system", Name: "hosts-a-back-1"},
						VarsConfRef: &apis.ConfigMapRef{
							NameSpace: "kubean-system", Name: "vars-a-back-1",
						},
					},
				}
				cluster := &clusterv1alpha1.Cluster{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Cluster",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster1-2",
					},
					Spec: clusterv1alpha1.Spec{
						HostsConfRef: &apis.ConfigMapRef{NameSpace: "kubean-system", Name: "hosts-a"},
						VarsConfRef:  &apis.ConfigMapRef{NameSpace: "kubean-system", Name: "vars-a"},
						SSHAuthRef:   &apis.SecretRef{NameSpace: "kubean-system", Name: "secret-a"},
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
				sshAuthSecret := &corev1.Secret{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Secret",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "kubean-system",
						Name:      "secret-a",
					},
				}
				controller.ClientSet.CoreV1().ConfigMaps("kubean-system").Create(context.Background(), hostsConfigMap, metav1.CreateOptions{})
				controller.ClientSet.CoreV1().ConfigMaps("kubean-system").Create(context.Background(), varsConfigMap, metav1.CreateOptions{})
				controller.ClientSet.CoreV1().Secrets("kubean-system").Create(context.Background(), sshAuthSecret, metav1.CreateOptions{})
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

func TestUpdateOperationOwnReferenceForCluster(t *testing.T) {
	genController := func() *Controller {
		return &Controller{
			Client:                newFakeClient(),
			ClientSet:             clientsetfake.NewSimpleClientset(),
			KubeanClusterSet:      clusterv1alpha1fake.NewSimpleClientset(),
			KubeanClusterOpsSet:   clusteroperationv1alpha1fake.NewSimpleClientset(),
			InfoManifestClientSet: manifestv1alpha1fake.NewSimpleClientset(),
		}
	}
	tests := []struct {
		name string
		arg  func() bool
		want bool
	}{
		{
			name: "the names are not same",
			arg: func() bool {
				controller := genController()
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
				clusterOps.Spec.Cluster = "cluster-origin"
				cluster := &clusterv1alpha1.Cluster{}
				cluster.Name = "cluster1"
				needRequeue, err := controller.UpdateOperationOwnReferenceForCluster(clusterOps, cluster)
				return err == nil && needRequeue == false
			},
			want: true,
		},
		{
			name: "the ownreference has been set",
			arg: func() bool {
				controller := genController()
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
				clusterOps.Spec.Cluster = "cluster-origin"

				cluster := &clusterv1alpha1.Cluster{}
				cluster.Name = "cluster-origin"
				cluster.UID = "cluster-uid-1"
				clusterOps.OwnerReferences = append(clusterOps.OwnerReferences, metav1.OwnerReference{UID: "cluster-uid-1"})
				needRequeue, err := controller.UpdateOperationOwnReferenceForCluster(clusterOps, cluster)
				return err == nil && needRequeue == false
			},
			want: true,
		},
		{
			name: "the ownreference has been set",
			arg: func() bool {
				controller := genController()
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
				clusterOps.Spec.Cluster = "cluster-origin"

				cluster := &clusterv1alpha1.Cluster{}
				cluster.Name = "cluster-origin"
				cluster.UID = "cluster-uid-1"
				clusterOps.OwnerReferences = append(clusterOps.OwnerReferences, metav1.OwnerReference{UID: "cluster-uid-1"})
				needRequeue, err := controller.UpdateOperationOwnReferenceForCluster(clusterOps, cluster)
				return err == nil && needRequeue == false
			},
			want: true,
		},
		{
			name: "the ownreference has been not set",
			arg: func() bool {
				controller := genController()
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{}
				clusterOps.Spec.Cluster = "cluster-origin"
				clusterOps.Name = "ops-1"
				controller.Client.Create(context.Background(), clusterOps)

				cluster := &clusterv1alpha1.Cluster{}
				cluster.Name = "cluster-origin"
				cluster.UID = "cluster-uid-1"
				clusterOps.OwnerReferences = nil
				needRequeue, err := controller.UpdateOperationOwnReferenceForCluster(clusterOps, cluster)
				clusterOpsResult := &clusteroperationv1alpha1.ClusterOperation{}
				controller.Client.Get(context.Background(), types.NamespacedName{Name: "ops-1"}, clusterOpsResult)
				return err == nil && needRequeue && len(clusterOpsResult.OwnerReferences) != 0 && clusterOpsResult.OwnerReferences[0].UID == cluster.UID
			},
			want: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.arg() != test.want {
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

func TestSetupWithManager(t *testing.T) {
	controller := Controller{
		Client:                newFakeClient(),
		ClientSet:             clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:      clusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet:   clusteroperationv1alpha1fake.NewSimpleClientset(),
		InfoManifestClientSet: manifestv1alpha1fake.NewSimpleClientset(),
	}
	if controller.SetupWithManager(MockManager{}) != nil {
		t.Fatal()
	}
}

type MockClusterForManager struct {
	_ string
}

func (MockClusterForManager) SetFields(interface{}) error { return nil }

func (MockClusterForManager) GetConfig() *rest.Config { return &rest.Config{} }

func (MockClusterForManager) GetScheme() *runtime.Scheme {
	sch := scheme.Scheme
	if err := manifestv1alpha1.AddToScheme(sch); err != nil {
		panic(err)
	}
	if err := localartifactsetv1alpha1.AddToScheme(sch); err != nil {
		panic(err)
	}
	return sch
}

func (MockClusterForManager) GetClient() client.Client { return nil }

func (MockClusterForManager) GetFieldIndexer() client.FieldIndexer { return nil }

func (MockClusterForManager) GetCache() cache.Cache { return nil }

func (MockClusterForManager) GetEventRecorderFor(name string) record.EventRecorder { return nil }

func (MockClusterForManager) GetRESTMapper() meta.RESTMapper { return nil }

func (MockClusterForManager) GetAPIReader() client.Reader { return nil }

func (MockClusterForManager) Start(ctx context.Context) error { return nil }

type MockManager struct {
	MockClusterForManager
}

func (MockManager) Add(manager.Runnable) error { return nil }

func (MockManager) Elected() <-chan struct{} { return nil }

func (MockManager) AddMetricsExtraHandler(path string, handler http.Handler) error { return nil }

func (MockManager) AddHealthzCheck(name string, check healthz.Checker) error { return nil }

func (MockManager) AddReadyzCheck(name string, check healthz.Checker) error { return nil }

func (MockManager) Start(ctx context.Context) error { return nil }

func (MockManager) GetWebhookServer() *webhook.Server { return nil }

func (MockManager) GetLogger() logr.Logger { return logr.Logger{} }

func (MockManager) GetControllerOptions() v1alpha1.ControllerConfigurationSpec {
	return v1alpha1.ControllerConfigurationSpec{}
}
