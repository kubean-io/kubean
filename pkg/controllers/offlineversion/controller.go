package offlineversion

import (
	"context"
	"time"

	kubeancomponentsversionv1alpha1 "kubean.io/api/apis/kubeancomponentsversion/v1alpha1"
	kubeanofflineversionv1alpha1 "kubean.io/api/apis/kubeanofflineversion/v1alpha1"
	"kubean.io/api/constants"
	kubeancomponentsversionClientSet "kubean.io/api/generated/kubeancomponentsversion/clientset/versioned"
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
	ClientSet                  kubernetes.Interface
	ComponentsversionClientSet kubeancomponentsversionClientSet.Interface
	OfflineversionClientSet    kubeanofflineversionClientSet.Interface
}

func (c *Controller) Start(ctx context.Context) error {
	klog.Warningf("KubeanComponentsVersion Controller Start")
	<-ctx.Done()
	return nil
}

func (c *Controller) FetchGlobalKuBeanComponentsVersion() (*kubeancomponentsversionv1alpha1.KuBeanComponentsVersion, error) {
	componentsVersion, err := c.ComponentsversionClientSet.KubeanV1alpha1().KuBeanComponentsVersions().Get(context.Background(), constants.ComponentsversionGlobalName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return componentsVersion, nil
}

func (c *Controller) MergeOfflineVersionStatus(offlineVersion *kubeanofflineversionv1alpha1.KuBeanOfflineVersion, componentsVersion *kubeancomponentsversionv1alpha1.KuBeanComponentsVersion) (bool, *kubeancomponentsversionv1alpha1.KuBeanComponentsVersion) {
	updated := false
	for _, dockerInfo := range offlineVersion.Spec.Docker {
		if componentsVersion.Status.Offline.MergeDockerInfo(dockerInfo.OS, dockerInfo.VersionRange) {
			updated = true
		}
	}
	for _, softItem := range offlineVersion.Spec.Items {
		if componentsVersion.Status.Offline.MergeSoftwareInfo(softItem.Name, softItem.VersionRange) {
			updated = true
		}
	}
	return updated, componentsVersion
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
	componentsVersion, err := c.FetchGlobalKuBeanComponentsVersion()
	if err != nil {
		klog.Errorf("Fetch %s , ignoring %s", constants.ComponentsversionGlobalName, err)
		return controllerruntime.Result{RequeueAfter: Loop}, nil
	}
	if needUpdate, newComponentsVersion := c.MergeOfflineVersionStatus(offlineVersion, componentsVersion); needUpdate {
		klog.Info("Update componentsVersion")
		if _, err := c.ComponentsversionClientSet.KubeanV1alpha1().KuBeanComponentsVersions().UpdateStatus(context.Background(), newComponentsVersion, metav1.UpdateOptions{}); err != nil {
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
