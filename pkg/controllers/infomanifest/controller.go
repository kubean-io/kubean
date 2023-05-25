package infomanifest

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/kubean-io/kubean/pkg/util"

	manifestv1alpha1 "github.com/kubean-io/kubean-api/apis/manifest/v1alpha1"
	"github.com/kubean-io/kubean-api/constants"
	localartifactsetClientSet "github.com/kubean-io/kubean-api/generated/localartifactset/clientset/versioned"
	manifestClientSet "github.com/kubean-io/kubean-api/generated/manifest/clientset/versioned"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	klog "k8s.io/klog/v2"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const Loop = time.Second * 30

const OriginLabel = "origin"

const LocalServiceConfigMap = "kubean-localservice"

type Controller struct {
	Client                    client.Client
	InfoManifestClientSet     manifestClientSet.Interface
	ClientSet                 kubernetes.Interface
	LocalArtifactSetClientSet localartifactsetClientSet.Interface
}

func (c *Controller) Start(ctx context.Context) error {
	klog.Warningf("InfoManifest Controller Start")
	<-ctx.Done()
	return nil
}

// FetchLatestInfoManifest , get infomanifest exclude the global-infomanifest.
func (c *Controller) FetchLatestInfoManifest() (*manifestv1alpha1.Manifest, error) {
	result, err := c.InfoManifestClientSet.KubeanV1alpha1().Manifests().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	items := make([]*manifestv1alpha1.Manifest, 0)
	for i := range result.Items {
		item := result.Items[i]
		if item.Name == constants.InfoManifestGlobal {
			continue
		}
		items = append(items, &item)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("not found the latest InfoManifest")
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreationTimestamp.After(items[j].CreationTimestamp.Time)
	})
	return items[0], nil
}

func NewGlobalInfoManifest(latestInfoManifest *manifestv1alpha1.Manifest) *manifestv1alpha1.Manifest {
	return &manifestv1alpha1.Manifest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Manifest",
			APIVersion: "kubean.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   constants.InfoManifestGlobal,
			Labels: map[string]string{OriginLabel: latestInfoManifest.Name},
		},
		Spec: latestInfoManifest.Spec,
	}
}

func (c *Controller) FetchGlobalInfoManifest() (*manifestv1alpha1.Manifest, error) {
	global, err := c.InfoManifestClientSet.KubeanV1alpha1().Manifests().Get(context.Background(), constants.InfoManifestGlobal, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return global, nil
}

func (c *Controller) EnsureGlobalInfoManifestBeingLatest(latestInfoManifest *manifestv1alpha1.Manifest) (*manifestv1alpha1.Manifest, error) {
	currentGlobalInfoManifest, err := c.InfoManifestClientSet.KubeanV1alpha1().Manifests().Get(context.Background(), constants.InfoManifestGlobal, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		// create global-infomanifest
		global, err := c.InfoManifestClientSet.KubeanV1alpha1().Manifests().Create(context.Background(), NewGlobalInfoManifest(latestInfoManifest), metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
		return global, nil
	}
	if err != nil {
		// other error
		return nil, err
	}
	if currentGlobalInfoManifest.Labels == nil || len(currentGlobalInfoManifest.Labels) == 0 || currentGlobalInfoManifest.Labels[OriginLabel] != latestInfoManifest.Name {
		currentGlobalInfoManifest.Labels = map[string]string{OriginLabel: latestInfoManifest.Name}
		currentGlobalInfoManifest.Spec = latestInfoManifest.Spec
		global, err := c.InfoManifestClientSet.KubeanV1alpha1().Manifests().Update(context.Background(), currentGlobalInfoManifest, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}
		return global, nil
	}
	return currentGlobalInfoManifest, nil
}

func (c *Controller) FetchLocalServiceCM(namespace string) (*corev1.ConfigMap, error) {
	if localServiceCM, _ := c.ClientSet.CoreV1().ConfigMaps(namespace).Get(context.Background(), LocalServiceConfigMap, metav1.GetOptions{}); localServiceCM != nil {
		return localServiceCM, nil
	}
	if namespace != "default" {
		namespace = "default"
		if localServiceCM, _ := c.ClientSet.CoreV1().ConfigMaps(namespace).Get(context.Background(), LocalServiceConfigMap, metav1.GetOptions{}); localServiceCM != nil {
			return localServiceCM, nil
		}
	}
	return nil, fmt.Errorf("not found kubean localService ConfigMap")
}

func (c *Controller) ParseConfigMapToLocalService(localServiceConfigMap *corev1.ConfigMap) (*manifestv1alpha1.LocalService, error) {
	localService := &manifestv1alpha1.LocalService{}
	if len(localServiceConfigMap.Data) == 0 {
		return localService, fmt.Errorf("kubean localService ConfigMap not found data")
	}
	if len(localServiceConfigMap.Data["localService"]) == 0 {
		return localService, fmt.Errorf("kubean localService ConfigMap not found key localService")
	}
	err := yaml.Unmarshal([]byte(localServiceConfigMap.Data["localService"]), localService)
	if err != nil {
		return localService, fmt.Errorf("unable to parse kubean localService ConfigMap data to LocalService ,%s", err.Error())
	}
	return localService, nil
}

func (c *Controller) UpdateGlobalLocalService() {
	if c.IsOnlineENV() {
		// if not airgap env , then do nothing and return
		return
	}
	localServiceConfigMap, err := c.FetchLocalServiceCM(util.GetCurrentNSOrDefault())
	if err != nil {
		klog.Warningf("ignoring %s", err.Error())
		return
	}
	localService, err := c.ParseConfigMapToLocalService(localServiceConfigMap)
	if err != nil {
		klog.Warningf("ignoring %s", err.Error())
		return
	}
	global, err := c.FetchGlobalInfoManifest()
	if err != nil {
		klog.Warningf("ignoring %s", err.Error())
		return
	}
	if !reflect.DeepEqual(&global.Spec.LocalService, localService) {
		global.Spec.LocalService = *localService
		klog.Warningf("update the global InfoManifest LocalService")
		_, err = c.InfoManifestClientSet.KubeanV1alpha1().Manifests().Update(context.Background(), global, metav1.UpdateOptions{})
		if err != nil {
			klog.Warningf("ignoring %s", err.Error())
		}
	}
}

// UpdateLocalAvailableImage update image infos to global-infomanifest-cr.
func (c *Controller) UpdateLocalAvailableImage() {
	global, err := c.FetchGlobalInfoManifest()
	if err != nil {
		klog.Warningf("ignoring %s", err.Error())
		return
	}
	imageRepo := "ghcr.m.daocloud.io"
	if len(global.Spec.LocalService.GetGHCRImageRepo()) != 0 {
		imageRepo = global.Spec.LocalService.GetGHCRImageRepo() // ghcr.io
	}
	newImageName := fmt.Sprintf("%s/kubean-io/spray-job:%s", imageRepo, global.Spec.KubeanVersion)
	if global.Status.LocalAvailable.KubesprayImage != newImageName {
		global.Status.LocalAvailable.KubesprayImage = newImageName
		_, err := c.InfoManifestClientSet.KubeanV1alpha1().Manifests().UpdateStatus(context.Background(), global, metav1.UpdateOptions{})
		if err != nil {
			klog.Warningf("ignoring %s", err.Error())
			return
		}
	}
}

// IsOnlineENV indicates what the running env is onLine or air-gap.
func (c *Controller) IsOnlineENV() bool {
	result, err := c.LocalArtifactSetClientSet.KubeanV1alpha1().LocalArtifactSets().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("%s ", err.Error())
		return true
	}
	if len(result.Items) == 0 {
		return true
	}
	return false
}

func (c *Controller) Reconcile(ctx context.Context, req controllerruntime.Request) (controllerruntime.Result, error) {
	if req.Name == constants.InfoManifestGlobal {
		return controllerruntime.Result{Requeue: false}, nil
	}
	klog.Infof("InfoManifest Controller receive event %s", req.Name)
	latestInfoManifest, err := c.FetchLatestInfoManifest()
	if err != nil {
		klog.Warningf("%s ", err.Error())
		return controllerruntime.Result{RequeueAfter: Loop}, nil
	}
	_, err = c.EnsureGlobalInfoManifestBeingLatest(latestInfoManifest)
	if err != nil {
		klog.Warningf("%s ", err.Error())
		return controllerruntime.Result{RequeueAfter: Loop}, nil
	}
	c.UpdateGlobalLocalService()
	c.UpdateLocalAvailableImage()
	return controllerruntime.Result{RequeueAfter: Loop}, nil
}

func (c *Controller) SetupWithManager(mgr controllerruntime.Manager) error {
	return utilerrors.NewAggregate([]error{
		controllerruntime.
			NewControllerManagedBy(mgr).
			For(&manifestv1alpha1.Manifest{}).
			Complete(c),
		mgr.Add(c),
	})
}
