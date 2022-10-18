package cluster

import (
	"testing"
	"time"

	clusterv1alpha1 "kubean.io/api/apis/cluster/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
