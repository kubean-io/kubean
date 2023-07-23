// Copyright 2023 Authors of kubean-io
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	clusterv1alpha1 "github.com/kubean-io/kubean-api/apis/cluster/v1alpha1"
	clusteroperationv1alpha1 "github.com/kubean-io/kubean-api/apis/clusteroperation/v1alpha1"
	clusterClientSet "github.com/kubean-io/kubean-api/generated/cluster/clientset/versioned"
	clusterOperationClientSet "github.com/kubean-io/kubean-api/generated/clusteroperation/clientset/versioned"
	"github.com/kubean-io/kubean/pkg/util"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	klog "k8s.io/klog/v2"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	RequeueAfter                         = time.Second * 15
	KubeanConfigMapName                  = "kubean-config"
	DefaultClusterOperationsBackEndLimit = 30
	MaxClusterOperationsBackEndLimit     = 200
)

type Controller struct {
	Client              client.Client
	ClientSet           kubernetes.Interface
	KubeanClusterSet    clusterClientSet.Interface
	KubeanClusterOpsSet clusterOperationClientSet.Interface
}

func (c *Controller) Start(ctx context.Context) error {
	klog.Warningf("Cluster Controller Start")
	<-ctx.Done()
	return nil
}

func CompareClusterCondition(conditionA, conditionB clusterv1alpha1.ClusterCondition) bool {
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

func CompareClusterConditions(condAList, condBList []clusterv1alpha1.ClusterCondition) bool {
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

func (c *Controller) UpdateStatus(cluster *clusterv1alpha1.Cluster) error {
	listOpt := metav1.ListOptions{LabelSelector: fmt.Sprintf("clusterName=%s", cluster.Name)}
	clusterOpsList, err := c.KubeanClusterOpsSet.KubeanV1alpha1().ClusterOperations().List(context.Background(), listOpt)
	if err != nil {
		return err
	}
	// clusterOps list sort by creation timestamp
	c.SortClusterOperationsByCreation(clusterOpsList.Items)
	newConditions := make([]clusterv1alpha1.ClusterCondition, 0)
	for _, item := range clusterOpsList.Items {
		newConditions = append(newConditions, clusterv1alpha1.ClusterCondition{
			ClusterOps: item.Name,
			Status:     clusterv1alpha1.ClusterConditionType(item.Status.Status),
			StartTime:  item.Status.StartTime,
			EndTime:    item.Status.EndTime,
		})
	}
	if !CompareClusterConditions(cluster.Status.Conditions, newConditions) {
		// need update for newCondition
		cluster.Status.Conditions = newConditions
		klog.Warningf("update cluster %s status.condition", cluster.Name)
		return c.Client.Status().Update(context.Background(), cluster)
	}
	return nil
}

// SortClusterOperationsByCreation operations from large to small by creation timestamp.
func (c *Controller) SortClusterOperationsByCreation(operations []clusteroperationv1alpha1.ClusterOperation) {
	sort.Slice(operations, func(i, j int) bool {
		return operations[i].CreationTimestamp.After(operations[j].CreationTimestamp.Time)
	})
}

// CleanExcessClusterOps clean up excess ClusterOperation.
func (c *Controller) CleanExcessClusterOps(cluster *clusterv1alpha1.Cluster, OpsBackupNum int) (bool, error) {
	listOpt := metav1.ListOptions{LabelSelector: fmt.Sprintf("clusterName=%s", cluster.Name)}
	clusterOpsList, err := c.KubeanClusterOpsSet.KubeanV1alpha1().ClusterOperations().List(context.Background(), listOpt)
	if err != nil {
		return false, err
	}
	if len(clusterOpsList.Items) <= OpsBackupNum {
		return false, nil
	}

	c.SortClusterOperationsByCreation(clusterOpsList.Items)

	excessClusterOpsList := clusterOpsList.Items[OpsBackupNum:]
	for _, item := range excessClusterOpsList {
		if item.Status.Status == clusteroperationv1alpha1.RunningStatus { // keep running job
			continue
		}
		klog.Warningf("Delete ClusterOperation: name: %s, createTime: %s, status: %s", item.Name, item.CreationTimestamp.String(), item.Status.Status)
		c.KubeanClusterOpsSet.KubeanV1alpha1().ClusterOperations().Delete(context.Background(), item.Name, metav1.DeleteOptions{})
	}
	return true, nil
}

type ConfigProperty struct {
	ClusterOperationsBackEndLimit string `json:"CLUSTER_OPERATIONS_BACKEND_LIMIT"`
}

func (config *ConfigProperty) GetClusterOperationsBackEndLimit() int {

	value, _ := strconv.Atoi(config.ClusterOperationsBackEndLimit)
	if value <= 0 {
		klog.Warningf("GetClusterOperationsBackEndLimit and use default value %d", DefaultClusterOperationsBackEndLimit)
		return DefaultClusterOperationsBackEndLimit
	}
	if value >= MaxClusterOperationsBackEndLimit {
		klog.Warningf("GetClusterOperationsBackEndLimit and use max value %d", MaxClusterOperationsBackEndLimit)
		return MaxClusterOperationsBackEndLimit
	}
	return value
}

func (c *Controller) FetchKubeanConfigProperty() *ConfigProperty {
	configData, err := c.ClientSet.CoreV1().ConfigMaps(util.GetCurrentNSOrDefault()).Get(context.Background(), KubeanConfigMapName, metav1.GetOptions{})
	if err != nil {
		return &ConfigProperty{}
	}
	jsonData, err := json.Marshal(configData.Data)
	if err != nil {
		return &ConfigProperty{}
	}
	result := &ConfigProperty{}
	err = json.Unmarshal(jsonData, result)
	if err != nil {
		return &ConfigProperty{}
	}
	return result
}

func (c *Controller) Reconcile(ctx context.Context, req controllerruntime.Request) (controllerruntime.Result, error) {
	cluster := &clusterv1alpha1.Cluster{}
	if err := c.Client.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			return controllerruntime.Result{}, nil
		}
		klog.ErrorS(err, "failed to get cluster", "cluster", req.String())
		return controllerruntime.Result{RequeueAfter: RequeueAfter}, nil
	}
	OpsBackupNum := c.FetchKubeanConfigProperty().GetClusterOperationsBackEndLimit()
	needRequeue, err := c.CleanExcessClusterOps(cluster, OpsBackupNum)
	if err != nil {
		klog.ErrorS(err, "failed to clean excess cluster ops", "cluster", cluster.Name)
		return controllerruntime.Result{RequeueAfter: RequeueAfter}, nil
	}
	if needRequeue {
		return controllerruntime.Result{RequeueAfter: RequeueAfter}, nil
	}

	if err := c.UpdateStatus(cluster); err != nil {
		klog.ErrorS(err, "failed to update cluster status", "cluster", cluster.Name)
		return controllerruntime.Result{RequeueAfter: RequeueAfter}, nil
	}

	return controllerruntime.Result{RequeueAfter: RequeueAfter}, nil // loop
}

func (c *Controller) SetupWithManager(mgr controllerruntime.Manager) error {
	return utilerrors.NewAggregate([]error{
		controllerruntime.NewControllerManagedBy(mgr).For(&clusterv1alpha1.Cluster{}).Complete(c),
		mgr.Add(c),
	})
}
