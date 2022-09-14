package infomanifest

import (
	"context"
	"time"

	kubeaninfomanifestv1alpha1 "kubean.io/api/apis/kubeaninfomanifest/v1alpha1"
	"kubean.io/api/constants"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const Loop = time.Second * 20

type Controller struct {
	client.Client
	ClientSet kubernetes.Interface
}

func (c *Controller) Start(ctx context.Context) error {
	klog.Warningf("KubeanClusterConfig Controller Start")
	<-ctx.Done()
	return nil
}

func (c *Controller) Reconcile(ctx context.Context, req controllerruntime.Request) (controllerruntime.Result, error) {
	infoManifest := &kubeaninfomanifestv1alpha1.KubeanInfoManifest{}
	if err := c.Client.Get(context.Background(), req.NamespacedName, infoManifest); err != nil {
		if apierrors.IsNotFound(err) {
			return controllerruntime.Result{Requeue: false}, nil
		}
		klog.Error(err)
		return controllerruntime.Result{RequeueAfter: Loop}, nil
	}
	if infoManifest.Name != constants.InfoManifestGlobal {
		klog.Errorf("KubeanClusterConfig % is not global", infoManifest.Name)
		return controllerruntime.Result{Requeue: false}, nil
	}
	// todo process globalClusterConfig
	return controllerruntime.Result{Requeue: false}, nil
}

func (c *Controller) SetupWithManager(mgr controllerruntime.Manager) error {
	return utilerrors.NewAggregate([]error{
		controllerruntime.NewControllerManagedBy(mgr).For(&kubeaninfomanifestv1alpha1.KubeanInfoManifest{}).Complete(c),
		mgr.Add(c),
	})
}
