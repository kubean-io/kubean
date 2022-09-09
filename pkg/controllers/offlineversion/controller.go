package offlineversion

import (
	"context"
	"time"

	kubeanclusterconfigv1alpha1 "kubean.io/api/apis/kubeanclusterconfig/v1alpha1"
	kubeanofflineversionv1alpha1 "kubean.io/api/apis/kubeanofflineversion/v1alpha1"
	"kubean.io/api/constants"
	kubeanclusterconfigClientSet "kubean.io/api/generated/kubeanclusterconfig/clientset/versioned"
	kubeanofflineversionClientSet "kubean.io/api/generated/kubeanofflineversion/clientset/versioned"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const Loop = time.Second * 20

type Controller struct {
	client.Client
	ClientSet               kubernetes.Interface
	ClusterConfigClientSet  kubeanclusterconfigClientSet.Interface
	OfflineversionClientSet kubeanofflineversionClientSet.Interface
}

func (c *Controller) Start(ctx context.Context) error {
	klog.Warningf("KubeanOfflineVersion Controller Start")
	<-ctx.Done()
	return nil
}

func (c *Controller) FetchGlobalKubeanClusterConfig() (*kubeanclusterconfigv1alpha1.KubeanClusterConfig, error) {
	componentsVersion, err := c.ClusterConfigClientSet.KubeanV1alpha1().KubeanClusterConfigs().Get(context.Background(), constants.ClusterConfigGlobal, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return componentsVersion, nil
}

func (c *Controller) MergeOfflineVersionStatus(offlineVersion *kubeanofflineversionv1alpha1.KuBeanOfflineVersion, clusterConfig *kubeanclusterconfigv1alpha1.KubeanClusterConfig) (bool, *kubeanclusterconfigv1alpha1.KubeanClusterConfig) {
	updated := false
	for _, dockerInfo := range offlineVersion.Spec.Docker {
		if clusterConfig.Status.AirGapStatus.MergeDockerInfo(dockerInfo.OS, dockerInfo.VersionRange) {
			updated = true
		}
	}
	for _, softItem := range offlineVersion.Spec.Items {
		if clusterConfig.Status.AirGapStatus.MergeSoftwareInfo(softItem.Name, softItem.VersionRange) {
			updated = true
		}
	}
	return updated, clusterConfig
}

func (c *Controller) Reconcile(ctx context.Context, req controllerruntime.Request) (controllerruntime.Result, error) {
	offlineVersion := &kubeanofflineversionv1alpha1.KuBeanOfflineVersion{}
	if err := c.Client.Get(context.Background(), req.NamespacedName, offlineVersion); err != nil {
		if apierrors.IsNotFound(err) {
			return controllerruntime.Result{Requeue: false}, nil
		}
		klog.Error(err)
		return controllerruntime.Result{RequeueAfter: Loop}, nil
	}
	globalClusterConfig, err := c.FetchGlobalKubeanClusterConfig()
	if err != nil {
		klog.Errorf("Fetch %s , ignoring %s", constants.ClusterConfigGlobal, err)
		return controllerruntime.Result{RequeueAfter: Loop}, nil
	}
	if needUpdate, newGlobalClusterConfig := c.MergeOfflineVersionStatus(offlineVersion, globalClusterConfig); needUpdate {
		klog.Info("Update componentsVersion")
		if _, err := c.ClusterConfigClientSet.KubeanV1alpha1().KubeanClusterConfigs().UpdateStatus(context.Background(), newGlobalClusterConfig, metav1.UpdateOptions{}); err != nil {
			klog.Error(err)
		}
	}
	return controllerruntime.Result{RequeueAfter: Loop}, nil // endless loop
}

func (c *Controller) SetupWithManager(mgr controllerruntime.Manager) error {
	return utilerrors.NewAggregate([]error{
		controllerruntime.NewControllerManagedBy(mgr).For(&kubeanofflineversionv1alpha1.KuBeanOfflineVersion{}).Complete(c),
		mgr.Add(c),
	})
}
