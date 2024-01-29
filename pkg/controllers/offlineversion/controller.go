// Copyright 2023 Authors of kubean-io
// SPDX-License-Identifier: Apache-2.0

package offlineversion

import (
	"context"
	"fmt"
	"time"

	localartifactsetv1alpha1 "github.com/kubean-io/kubean-api/apis/localartifactset/v1alpha1"
	manifestv1alpha1 "github.com/kubean-io/kubean-api/apis/manifest/v1alpha1"
	"github.com/kubean-io/kubean-api/constants"
	localartifactsetClientSet "github.com/kubean-io/kubean-api/generated/localartifactset/clientset/versioned"
	manifestClientSet "github.com/kubean-io/kubean-api/generated/manifest/clientset/versioned"
	"github.com/kubean-io/kubean/pkg/controllers/infomanifest"

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

// MergeManifestsStatus merge the status of manifests which has same sprayRelease label of the localartifactset.
func (c *Controller) MergeManifestsStatus(localartifactset *localartifactsetv1alpha1.LocalArtifactSet, manifests []*manifestv1alpha1.Manifest) ([]*manifestv1alpha1.Manifest, error) {
	for _, manifest := range manifests {
		updated := false
		for _, dockerInfo := range localartifactset.Spec.Docker {
			if manifest.Status.LocalAvailable.MergeDockerInfo(dockerInfo.OS, dockerInfo.VersionRange) {
				updated = true
			}
		}
		for _, softItem := range localartifactset.Spec.Items {
			if manifest.Status.LocalAvailable.MergeSoftwareInfo(softItem.Name, softItem.VersionRange) {
				updated = true
			}
		}
		if !updated {
			continue
		}
		klog.Infof("Update manifest status for %s since %s", manifest.Name, localartifactset.Name)
		if _, err := c.InfoManifestClientSet.KubeanV1alpha1().Manifests().UpdateStatus(context.Background(), manifest, metav1.UpdateOptions{}); err != nil {
			return nil, fmt.Errorf("failed to merge status for manifest %s, %v", manifest.Name, err)
		}
	}
	return manifests, nil
}

func (c *Controller) Reconcile(ctx context.Context, req controllerruntime.Request) (controllerruntime.Result, error) {
	localartifactset := &localartifactsetv1alpha1.LocalArtifactSet{}
	if err := c.Client.Get(context.Background(), req.NamespacedName, localartifactset); err != nil {
		if apierrors.IsNotFound(err) {
			return controllerruntime.Result{}, nil
		}
		klog.Error(err)
		return controllerruntime.Result{RequeueAfter: Loop}, nil
	}

	sprayRelease, ok := localartifactset.ObjectMeta.Labels[constants.KeySprayRelease]
	if !ok {
		klog.Warningf("No label %s found in %s", constants.KeySprayRelease, localartifactset.Name)

		// Migrate old versions of LocalArtifactSet without a release version label.
		sprayRelease = "master"
		if localartifactset.ObjectMeta.Labels != nil {
			localartifactset.ObjectMeta.Labels[constants.KeySprayRelease] = sprayRelease
		} else {
			localartifactset.ObjectMeta.Labels = map[string]string{constants.KeySprayRelease: sprayRelease}
		}
		_, err := c.LocalArtifactSetClientSet.KubeanV1alpha1().LocalArtifactSets().Update(context.Background(), localartifactset, metav1.UpdateOptions{})
		if err != nil {
			klog.Error(err)
			return controllerruntime.Result{RequeueAfter: Loop}, nil
		}
		return controllerruntime.Result{}, nil
	}

	manifests, ok := infomanifest.GetVersionedManifest().Manifests[sprayRelease]
	if !ok {
		return controllerruntime.Result{RequeueAfter: Loop}, nil
	}
	if _, err := c.MergeManifestsStatus(localartifactset, manifests); err != nil {
		klog.Error(err)
		return controllerruntime.Result{RequeueAfter: Loop}, nil
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
