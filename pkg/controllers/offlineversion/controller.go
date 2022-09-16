package offlineversion

import (
	"context"
	"time"

	kubeaninfomanifestv1alpha1 "kubean.io/api/apis/kubeaninfomanifest/v1alpha1"
	kubeanofflineversionv1alpha1 "kubean.io/api/apis/kubeanofflineversion/v1alpha1"
	"kubean.io/api/constants"
	kubeaninfomanifestClientSet "kubean.io/api/generated/kubeaninfomanifest/clientset/versioned"
	kubeanofflineversionClientSet "kubean.io/api/generated/kubeanofflineversion/clientset/versioned"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const Loop = time.Second * 15

type Controller struct {
	client.Client
	ClientSet               kubernetes.Interface
	InfoManifestClientSet   kubeaninfomanifestClientSet.Interface
	OfflineversionClientSet kubeanofflineversionClientSet.Interface
}

func (c *Controller) Start(ctx context.Context) error {
	klog.Warningf("KubeanOfflineVersion Controller Start")
	<-ctx.Done()
	return nil
}

func (c *Controller) FetchGlobalKubeanClusterConfig() (*kubeaninfomanifestv1alpha1.KubeanInfoManifest, error) {
	infoManifest, err := c.InfoManifestClientSet.KubeanV1alpha1().KubeanInfoManifests().Get(context.Background(), constants.InfoManifestGlobal, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return infoManifest, nil
}

func (c *Controller) MergeOfflineVersionStatus(offlineVersion *kubeanofflineversionv1alpha1.KuBeanOfflineVersion, clusterConfig *kubeaninfomanifestv1alpha1.KubeanInfoManifest) (bool, *kubeaninfomanifestv1alpha1.KubeanInfoManifest) {
	updated := false
	for _, dockerInfo := range offlineVersion.Spec.Docker {
		if clusterConfig.Status.LocalAvailable.MergeDockerInfo(dockerInfo.OS, dockerInfo.VersionRange) {
			updated = true
		}
	}
	for _, softItem := range offlineVersion.Spec.Items {
		if clusterConfig.Status.LocalAvailable.MergeSoftwareInfo(softItem.Name, softItem.VersionRange) {
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
	globalInfoManifest, err := c.FetchGlobalKubeanClusterConfig()
	if err != nil {
		klog.Errorf("Fetch %s , ignoring %s", constants.InfoManifestGlobal, err)
		return controllerruntime.Result{RequeueAfter: Loop}, nil
	}
	if needUpdate, newGlobalInfoManifest := c.MergeOfflineVersionStatus(offlineVersion, globalInfoManifest); needUpdate {
		klog.Info("Update componentsVersion")
		if _, err := c.InfoManifestClientSet.KubeanV1alpha1().KubeanInfoManifests().UpdateStatus(context.Background(), newGlobalInfoManifest, metav1.UpdateOptions{}); err != nil {
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
