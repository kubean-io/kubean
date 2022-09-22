package cluster

import (
	"context"
	"fmt"
	"sort"
	"time"

	kubeanclusterv1alpha1 "kubean.io/api/apis/kubeancluster/v1alpha1"
	kubeanClusterClientSet "kubean.io/api/generated/kubeancluster/clientset/versioned"
	kubeanClusterOpsClientSet "kubean.io/api/generated/kubeanclusterops/clientset/versioned"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	RequeueAfter = time.Second * 5
	OpsBackupNum = 5
)

type Controller struct {
	client.Client
	ClientSet           *kubernetes.Clientset
	KubeanClusterSet    *kubeanClusterClientSet.Clientset
	KubeanClusterOpsSet *kubeanClusterOpsClientSet.Clientset
}

func (c *Controller) Start(ctx context.Context) error {
	klog.Warningf("KuBeanCluster Controller Start")
	<-ctx.Done()
	return nil
}

func CompareClusterCondition(conditionA, conditionB kubeanclusterv1alpha1.ClusterCondition) bool {
	unixMilli := func(t *metav1.Time) int64 {
		if t == nil {
			return -1
		}
		return t.UnixMilli()
	}
	if conditionA.ClusterOps != conditionB.ClusterOps {
		return false
	}
	if conditionA.Status != conditionB.Status {
		return false
	}
	if unixMilli(conditionA.StartTime) != unixMilli(conditionB.StartTime) {
		return false
	}
	if unixMilli(conditionA.EndTime) != unixMilli(conditionB.EndTime) {
		return false
	}
	return true
}

func CompareClusterConditions(condAList, condBList []kubeanclusterv1alpha1.ClusterCondition) bool {
	if len(condAList) != len(condBList) {
		return false
	}
	for i := range condAList {
		if !CompareClusterCondition(condAList[i], condBList[i]) {
			return false
		}
	}
	return true
}

func (c *Controller) UpdateStatus(cluster *kubeanclusterv1alpha1.KuBeanCluster) error {
	listOpt := metav1.ListOptions{LabelSelector: fmt.Sprintf("clusterName=%s", cluster.Name)}
	clusterOpsList, err := c.KubeanClusterOpsSet.KubeanV1alpha1().KuBeanClusterOps().List(context.Background(), listOpt)
	if err != nil {
		return err
	}
	// clusterOps list sort by creation timestamp
	sort.Slice(clusterOpsList.Items, func(i, j int) bool {
		return clusterOpsList.Items[i].CreationTimestamp.After(clusterOpsList.Items[j].CreationTimestamp.Time)
	})
	newConditions := make([]kubeanclusterv1alpha1.ClusterCondition, 0)
	for _, item := range clusterOpsList.Items {
		newConditions = append(newConditions, kubeanclusterv1alpha1.ClusterCondition{
			ClusterOps: item.Name,
			Status:     kubeanclusterv1alpha1.ClusterConditionType(item.Status.Status),
			StartTime:  item.Status.StartTime,
			EndTime:    item.Status.EndTime,
		})
	}
	if !CompareClusterConditions(cluster.Status.Conditions, newConditions) {
		// need update for newCondition
		cluster.Status.Conditions = newConditions
		klog.Warningf("update cluster %s status.condition", cluster.Name)
		return c.Status().Update(context.Background(), cluster)
	}
	return nil
}

// CleanExcessClusterOps clean up excess KuBeanClusterOps.
func (c *Controller) CleanExcessClusterOps(cluster *kubeanclusterv1alpha1.KuBeanCluster) (bool, error) {
	listOpt := metav1.ListOptions{LabelSelector: fmt.Sprintf("clusterName=%s", cluster.Name)}
	clusterOpsList, err := c.KubeanClusterOpsSet.KubeanV1alpha1().KuBeanClusterOps().List(context.Background(), listOpt)
	if err != nil {
		return false, err
	}
	if len(clusterOpsList.Items) <= OpsBackupNum {
		return false, nil
	}

	// clusterOps list sort by creation timestamp
	sort.Slice(clusterOpsList.Items, func(i, j int) bool {
		return clusterOpsList.Items[i].CreationTimestamp.After(clusterOpsList.Items[j].CreationTimestamp.Time)
	})
	excessClusterOpsList := clusterOpsList.Items[OpsBackupNum:]
	for _, item := range excessClusterOpsList {
		klog.Warningf("Delete KuBeanClusterOps: name: %s, createTime: %s", item.Name, item.CreationTimestamp.String())
		c.KubeanClusterOpsSet.KubeanV1alpha1().KuBeanClusterOps().Delete(context.Background(), item.Name, metav1.DeleteOptions{})
	}
	return true, nil
}

func (c *Controller) Reconcile(ctx context.Context, req controllerruntime.Request) (controllerruntime.Result, error) {
	cluster := &kubeanclusterv1alpha1.KuBeanCluster{}
	if err := c.Client.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			return controllerruntime.Result{Requeue: false}, nil
		}
		klog.Error(err)
		return controllerruntime.Result{RequeueAfter: RequeueAfter}, err
	}

	needRequeue, err := c.CleanExcessClusterOps(cluster)
	if err != nil {
		klog.Error(err)
		return controllerruntime.Result{RequeueAfter: RequeueAfter}, err
	}
	if needRequeue {
		return controllerruntime.Result{RequeueAfter: RequeueAfter}, nil
	}

	if err := c.UpdateStatus(cluster); err != nil {
		klog.Error(err)
	}
	return controllerruntime.Result{RequeueAfter: RequeueAfter}, nil // loop
}

func (c *Controller) SetupWithManager(mgr controllerruntime.Manager) error {
	return utilerrors.NewAggregate([]error{
		controllerruntime.NewControllerManagedBy(mgr).For(&kubeanclusterv1alpha1.KuBeanCluster{}).Complete(c),
		mgr.Add(c),
	})
}
