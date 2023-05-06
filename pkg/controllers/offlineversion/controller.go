package offlineversion

import (
	"context"
	"time"

	localartifactsetv1alpha1 "github.com/kubean-io/kubean-api/apis/localartifactset/v1alpha1"
	manifestv1alpha1 "github.com/kubean-io/kubean-api/apis/manifest/v1alpha1"
	"github.com/kubean-io/kubean-api/constants"
	localartifactsetClientSet "github.com/kubean-io/kubean-api/generated/localartifactset/clientset/versioned"
	manifestClientSet "github.com/kubean-io/kubean-api/generated/manifest/clientset/versioned"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	klog "k8s.io/klog/v2"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const Loop = time.Second * 15

type Controller struct {
	Client                    client.Client
	ClientSet                 kubernetes.Interface
	InfoManifestClientSet     manifestClientSet.Interface
	LocalArtifactSetClientSet localartifactsetClientSet.Interface
}

func (c *Controller) Start(ctx context.Context) error {
	klog.Warningf("KubeanOfflineVersion Controller Start")
	<-ctx.Done()
	return nil
}

func (c *Controller) FetchGlobalKubeanClusterConfig() (*manifestv1alpha1.Manifest, error) {
	infoManifest, err := c.InfoManifestClientSet.KubeanV1alpha1().Manifests().Get(context.Background(), constants.InfoManifestGlobal, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return infoManifest, nil
}

func (c *Controller) MergeOfflineVersionStatus(offlineVersion *localartifactsetv1alpha1.LocalArtifactSet, clusterConfig *manifestv1alpha1.Manifest) (bool, *manifestv1alpha1.Manifest) {
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
	offlineVersion := &localartifactsetv1alpha1.LocalArtifactSet{}
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
		if _, err := c.InfoManifestClientSet.KubeanV1alpha1().Manifests().UpdateStatus(context.Background(), newGlobalInfoManifest, metav1.UpdateOptions{}); err != nil {
			klog.Error(err)
		}
	}
	return controllerruntime.Result{RequeueAfter: Loop}, nil // endless loop
}

func (c *Controller) SetupWithManager(mgr controllerruntime.Manager) error {
	return utilerrors.NewAggregate([]error{
		controllerruntime.NewControllerManagedBy(mgr).
			For(&localartifactsetv1alpha1.LocalArtifactSet{}).
			Complete(c),
		mgr.Add(c),
	})
}
