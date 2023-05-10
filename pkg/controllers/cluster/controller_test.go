package cluster

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/kubean-io/kubean-api/apis"
	clusterv1alpha1 "github.com/kubean-io/kubean-api/apis/cluster/v1alpha1"
	clusteroperationv1alpha1 "github.com/kubean-io/kubean-api/apis/clusteroperation/v1alpha1"
	"github.com/kubean-io/kubean-api/constants"
	clusterv1alpha1fake "github.com/kubean-io/kubean-api/generated/cluster/clientset/versioned/fake"
	clusteroperationv1alpha1fake "github.com/kubean-io/kubean-api/generated/clusteroperation/clientset/versioned/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCompareClusterCondition(t *testing.T) {
	tests := []struct {
		name string
		args func(condA, conB clusterv1alpha1.ClusterCondition) bool
		want bool
	}{
		{
			name: "same",
			args: func(condA, condB clusterv1alpha1.ClusterCondition) bool {
				return CompareClusterCondition(condA, condB)
			},
			want: true,
		},
		{
			name: "same again",
			args: func(condA, condB clusterv1alpha1.ClusterCondition) bool {
				condA.Status = "123"
				condB.Status = "123"
				return CompareClusterCondition(condA, condB)
			},
			want: true,
		},
		{
			name: "clusterOps",
			args: func(condA, condB clusterv1alpha1.ClusterCondition) bool {
				condA.ClusterOps = "12"
				return CompareClusterCondition(condA, condB)
			},
			want: false,
		},
		{
			name: "status",
			args: func(condA, condB clusterv1alpha1.ClusterCondition) bool {
				condA.Status = "121212"
				return CompareClusterCondition(condA, condB)
			},
			want: false,
		},
		{
			name: "startTime",
			args: func(condA, condB clusterv1alpha1.ClusterCondition) bool {
				condA.StartTime = &metav1.Time{Time: time.Now()}
				return CompareClusterCondition(condA, condB)
			},
			want: false,
		},
		{
			name: "endTime",
			args: func(condA, condB clusterv1alpha1.ClusterCondition) bool {
				condA.EndTime = &metav1.Time{Time: time.Now()}
				return CompareClusterCondition(condA, condB)
			},
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.args(clusterv1alpha1.ClusterCondition{}, clusterv1alpha1.ClusterCondition{}) != test.want {
				t.Fatal()
			}
		})
	}
}

func TestCompareClusterConditions(t *testing.T) {
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "zero length",
			args: func() bool {
				return CompareClusterConditions(nil, nil)
			},
			want: true,
		},
		{
			name: "different length",
			args: func() bool {
				return CompareClusterConditions(make([]clusterv1alpha1.ClusterCondition, 1), nil)
			},
			want: false,
		},
		{
			name: "one length",
			args: func() bool {
				return CompareClusterConditions(make([]clusterv1alpha1.ClusterCondition, 1), make([]clusterv1alpha1.ClusterCondition, 1))
			},
			want: true,
		},
		{
			name: "one length with different data",
			args: func() bool {
				condA := make([]clusterv1alpha1.ClusterCondition, 1)
				condB := make([]clusterv1alpha1.ClusterCondition, 1)
				condA[0].ClusterOps = "11"
				condB[0].ClusterOps = "22"
				return CompareClusterConditions(condA, condB)
			},
			want: false,
		},
	}

	for _, test := range tests {
		if test.args() != test.want {
			t.Fatal()
		}
	}
}

func TestSortClusterOperationsByCreation(t *testing.T) {
	controller := &Controller{
		Client:              newFakeClient(),
		ClientSet:           clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:    clusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet: clusteroperationv1alpha1fake.NewSimpleClientset(),
	}
	tests := []struct {
		name string
		args []clusteroperationv1alpha1.ClusterOperation
		want []clusteroperationv1alpha1.ClusterOperation
	}{
		{
			name: "empty slice",
			args: nil,
			want: nil,
		},
		{
			name: "unsorted slice",
			args: []clusteroperationv1alpha1.ClusterOperation{
				{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.Unix(2, 0)}},
				{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.Unix(1, 0)}},
				{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.Unix(3, 0)}},
			},
			want: []clusteroperationv1alpha1.ClusterOperation{
				{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.Unix(3, 0)}},
				{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.Unix(2, 0)}},
				{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.Unix(1, 0)}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			controller.SortClusterOperationsByCreation(test.args)
			if !reflect.DeepEqual(test.args, test.want) {
				t.Fatal()
			}
		})
	}
}

func Test_CleanExcessClusterOps(t *testing.T) {
	controller := &Controller{
		Client:              newFakeClient(),
		ClientSet:           clientsetfake.NewSimpleClientset(),
		KubeanClusterSet:    clusterv1alpha1fake.NewSimpleClientset(),
		KubeanClusterOpsSet: clusteroperationv1alpha1fake.NewSimpleClientset(),
	}
	exampleCluster := &clusterv1alpha1.Cluster{
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
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "get nothing",
			args: func() bool {
				result, _ := controller.CleanExcessClusterOps(exampleCluster)
				return result
			},
			want: false,
		},
		{
			name: "OpsBackupNum clusterOperations",
			args: func() bool {
				for i := 0; i < OpsBackupNum; i++ {
					clusterOperation := &clusteroperationv1alpha1.ClusterOperation{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterOperation",
							APIVersion: "kubean.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:              "my_kubean_ops_cluster_1_" + fmt.Sprint(i),
							Labels:            map[string]string{constants.KubeanClusterLabelKey: "cluster1"},
							CreationTimestamp: metav1.Unix(int64(i), 0),
						},
					}
					controller.KubeanClusterOpsSet.KubeanV1alpha1().ClusterOperations().Create(context.Background(), clusterOperation, metav1.CreateOptions{})
				}
				result, _ := controller.CleanExcessClusterOps(exampleCluster)
				return result
			},
			want: false,
		},
		{
			name: "clean clusterOperations",
			args: func() bool {
				for i := 0; i < OpsBackupNum*2; i++ {
					clusterOperation := &clusteroperationv1alpha1.ClusterOperation{
						TypeMeta: metav1.TypeMeta{
							Kind:       "ClusterOperation",
							APIVersion: "kubean.io/v1alpha1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:              "my_kubean_ops_cluster_2_" + fmt.Sprint(i),
							Labels:            map[string]string{constants.KubeanClusterLabelKey: "cluster1"},
							CreationTimestamp: metav1.Unix(int64(i), 0),
						},
					}
					controller.KubeanClusterOpsSet.KubeanV1alpha1().ClusterOperations().Create(context.Background(), clusterOperation, metav1.CreateOptions{})
				}
				result, _ := controller.CleanExcessClusterOps(exampleCluster)
				return result
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

func Test_UpdateStatus(t *testing.T) {
	controller := &Controller{
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
			name: "get nothing",
			args: func() bool {
				exampleCluster := &clusterv1alpha1.Cluster{
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
				return controller.UpdateStatus(exampleCluster) == nil
			},
			want: true,
		},
		{
			name: "get some ops",
			args: func() bool {
				exampleCluster := &clusterv1alpha1.Cluster{
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
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   "cluster1-ops",
						Labels: map[string]string{constants.KubeanClusterLabelKey: exampleCluster.Name},
					},
				}
				controller.Create(context.Background(), clusterOps)
				controller.Create(context.Background(), exampleCluster)
				controller.KubeanClusterSet.KubeanV1alpha1().Clusters().Create(context.Background(), exampleCluster, metav1.CreateOptions{})
				controller.KubeanClusterOpsSet.KubeanV1alpha1().ClusterOperations().Create(context.Background(), clusterOps, metav1.CreateOptions{})
				return controller.UpdateStatus(exampleCluster) == nil
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
