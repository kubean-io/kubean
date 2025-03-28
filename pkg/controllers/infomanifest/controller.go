// Copyright 2023 Authors of kubean-io
// SPDX-License-Identifier: Apache-2.0

package infomanifest

import (
	"context"
	"fmt"
	"reflect"
	"sync"
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

const LocalServiceConfigMap = "kubean-localservice"

var versionedManifest *VersionedManifest

type Controller struct {
	Client                    client.Client
	InfoManifestClientSet     manifestClientSet.Interface
	ClientSet                 kubernetes.Interface
	LocalArtifactSetClientSet localartifactsetClientSet.Interface
}

type VersionedManifest struct {
	mutex     sync.Mutex
	Manifests map[string][]*manifestv1alpha1.Manifest
}

func (c *Controller) Start(ctx context.Context) error {
	klog.Warningf("InfoManifest Controller Start")
	<-ctx.Done()
	return nil
}

func GetVersionedManifest() *VersionedManifest {
	if versionedManifest == nil {
		versionedManifest = &VersionedManifest{
			Manifests: make(map[string][]*manifestv1alpha1.Manifest, 0),
		}
	}
	return versionedManifest
}

func (m *VersionedManifest) Op(op string, m1, m2 *manifestv1alpha1.Manifest) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if op == "add" {
		m.add(m1)
	}
	if op == "update" {
		m.update(m1, m2)
	}
	if op == "delete" {
		m.delete(m1)
	}
}

func (m *VersionedManifest) add(manifest *manifestv1alpha1.Manifest) {
	sprayRelease, ok := manifest.Labels[constants.KeySprayRelease]
	if !ok {
		return
	}

	manifests, ok := m.Manifests[sprayRelease]
	if !ok {
		manifests = make([]*manifestv1alpha1.Manifest, 0)
	}
	m.Manifests[sprayRelease] = append(manifests, manifest)
}

func (m *VersionedManifest) update(manifest1, manifest2 *manifestv1alpha1.Manifest) {
	oldSprayRelease, ok1 := manifest1.Labels[constants.KeySprayRelease]
	newSprayRelease, ok2 := manifest2.Labels[constants.KeySprayRelease]
	if !ok1 && !ok2 {
		return
	} else if ok1 && !ok2 {
		m.delete(manifest1)
	} else if !ok1 && ok2 {
		m.add(manifest2)
	} else if oldSprayRelease != newSprayRelease || !reflect.DeepEqual(manifest1, manifest2) {
		m.delete(manifest1)
		m.add(manifest2)
	}
}

func (m *VersionedManifest) delete(manifest *manifestv1alpha1.Manifest) {
	sprayRelease, ok := manifest.Labels[constants.KeySprayRelease]
	if !ok {
		return
	}
	manifests, ok := m.Manifests[sprayRelease]
	if !ok {
		return
	}

	removeManifest := func(manifests []*manifestv1alpha1.Manifest, manifest *manifestv1alpha1.Manifest) []*manifestv1alpha1.Manifest {
		for i, m := range manifests {
			if m.Name == manifest.Name {
				manifests = append(manifests[:i], manifests[i+1:]...)
				break
			}
		}
		return manifests
	}

	m.Manifests[sprayRelease] = removeManifest(manifests, manifest)
	if len(m.Manifests[sprayRelease]) == 0 {
		delete(m.Manifests, sprayRelease)
	}
}

func (c *Controller) FetchLocalServiceCM(namespace string) (*corev1.ConfigMap, error) {
	localServiceCM, err := c.ClientSet.CoreV1().ConfigMaps(namespace).Get(context.Background(), LocalServiceConfigMap, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) && namespace != "default" {
		localServiceCM, err = c.ClientSet.CoreV1().ConfigMaps("default").Get(context.Background(), LocalServiceConfigMap, metav1.GetOptions{})
	}
	if err != nil {
		return nil, err
	}
	return localServiceCM, nil
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

// UpdateLocalService sync the content of local-service configmap into spec.
func (c *Controller) UpdateLocalService(manifests []manifestv1alpha1.Manifest) bool {
	if c.IsOnlineENV() {
		// if not airgap environment, do nothing and return
		return false
	}
	localServiceConfigMap, err := c.FetchLocalServiceCM(util.GetCurrentNSOrDefault())
	if err != nil {
		if !apierrors.IsNotFound(err) {
			klog.Warningf("ignoring error: %s", err.Error())
		}
		return false
	}
	localService, err := c.ParseConfigMapToLocalService(localServiceConfigMap)
	if err != nil {
		klog.Warningf("ignoring error: %s", err.Error())
		return false
	}
	for _, manifest := range manifests {
		if !reflect.DeepEqual(&manifest.Spec.LocalService, localService) {
			manifest.Spec.LocalService = *localService
			klog.Infof("Update local-service for %s", manifest.Name)
			_, err = c.InfoManifestClientSet.KubeanV1alpha1().Manifests().Update(context.Background(), &manifest, metav1.UpdateOptions{})
			if err != nil {
				klog.Warningf("ignoring error: %s", err.Error())
			}
			return true
		}
	}
	return false
}

// UpdateLocalAvailableImage update image infos into status.
func (c *Controller) UpdateLocalAvailableImage(manifests []manifestv1alpha1.Manifest) {
	imageRepo := util.FetchKubeanConfigProperty(c.ClientSet).SprayJobImageRegistry
	if imageRepo == "" {
		imageRepo = "ghcr.m.daocloud.io"
	}
	for _, manifest := range manifests {
		var newImageName string
		sprayRelease := manifest.Annotations[constants.KeySprayRelease]
		sprayCommit := manifest.Annotations[constants.KeySprayCommit]
		if sprayRelease != "" && sprayCommit != "" {
			newImageName = fmt.Sprintf("%s/kubean-io/spray-job:%s-%s", imageRepo, sprayRelease, sprayCommit)
		} else {
			newImageName = fmt.Sprintf("%s/kubean-io/spray-job:%s", imageRepo, manifest.Spec.KubeanVersion)
		}
		if manifest.Status.LocalAvailable.KubesprayImage != newImageName {
			manifest.Status.LocalAvailable.KubesprayImage = newImageName
			_, err := c.InfoManifestClientSet.KubeanV1alpha1().Manifests().UpdateStatus(context.Background(), &manifest, metav1.UpdateOptions{})
			if err != nil {
				klog.Warningf("ignoring error: %s", err.Error())
				return
			}
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
	manifestItems, err := c.InfoManifestClientSet.KubeanV1alpha1().Manifests().List(ctx, metav1.ListOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return controllerruntime.Result{}, nil
		}
		klog.Error(err)
		return controllerruntime.Result{RequeueAfter: Loop}, nil
	}

	if requeue := c.UpdateLocalService(manifestItems.Items); requeue {
		return controllerruntime.Result{RequeueAfter: Loop}, nil
	}
	c.UpdateLocalAvailableImage(manifestItems.Items)
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
