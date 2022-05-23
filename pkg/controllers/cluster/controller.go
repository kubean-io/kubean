package cluster

import (
	"context"
	"time"

	kubeanclusterv1alpha1 "github.com/daocloud/kubean/pkg/apis/kubeancluster/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Controller struct {
	client.Client
}

func (c *Controller) Start(ctx context.Context) error {
	klog.Warningf("KuBeanCluster Controller Start")
	<-ctx.Done()
	return nil
}

func (c *Controller) Reconcile(ctx context.Context, req controllerruntime.Request) (controllerruntime.Result, error) {
	cluster := &kubeanclusterv1alpha1.KuBeanCluster{}
	if err := c.Client.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			return controllerruntime.Result{Requeue: false}, nil
		}
		klog.Error(err)
		return controllerruntime.Result{RequeueAfter: time.Second}, err
	}
	// todo
	return controllerruntime.Result{Requeue: false}, nil
}

func (c *Controller) SetupWithManager(mgr controllerruntime.Manager) error {
	return utilerrors.NewAggregate([]error{
		controllerruntime.NewControllerManagedBy(mgr).For(&kubeanclusterv1alpha1.KuBeanCluster{}).Complete(c),
		mgr.Add(c),
	})
}
