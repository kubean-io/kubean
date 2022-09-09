package clusterconfig

import (
	"context"
	"time"

	kubeanclusterconfigv1alpha1 "kubean.io/api/apis/kubeanclusterconfig/v1alpha1"
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
	globalClusterConfig := &kubeanclusterconfigv1alpha1.KubeanClusterConfig{}
	if err := c.Client.Get(context.Background(), req.NamespacedName, globalClusterConfig); err != nil {
		if apierrors.IsNotFound(err) {
			return controllerruntime.Result{Requeue: false}, nil
		}
		klog.Error(err)
		return controllerruntime.Result{RequeueAfter: Loop}, nil
	}
	if globalClusterConfig.Name != constants.ClusterConfigGlobal {
		klog.Errorf("KubeanClusterConfig % is not global", globalClusterConfig.Name)
		return controllerruntime.Result{Requeue: false}, nil
	}
	// todo process globalClusterConfig
	return controllerruntime.Result{Requeue: false}, nil
}

func (c *Controller) SetupWithManager(mgr controllerruntime.Manager) error {
	return utilerrors.NewAggregate([]error{
		controllerruntime.NewControllerManagedBy(mgr).For(&kubeanclusterconfigv1alpha1.KubeanClusterConfig{}).Complete(c),
		mgr.Add(c),
	})
}
