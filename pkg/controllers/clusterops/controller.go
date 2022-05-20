package clusterops

import (
	"context"

	kubeanclusteropsv1alpha1 "github.com/daocloud/kubean/pkg/apis/kubeanclusterops/v1alpha1"

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
	klog.Warningf("KuBeanClusterOps Controller Start")
	<-ctx.Done()
	return nil
}

func (c *Controller) Reconcile(ctx context.Context, req controllerruntime.Request) (controllerruntime.Result, error) {
	clusterOps := &kubeanclusteropsv1alpha1.KuBeanClusterOps{}
	if err := c.Client.Get(ctx, req.NamespacedName, clusterOps); err != nil {
		if apierrors.IsNotFound(err) {
			return controllerruntime.Result{Requeue: false}, nil
		}
		return controllerruntime.Result{Requeue: true}, err
	}
	// todo
	return controllerruntime.Result{Requeue: false}, nil
}

func (c *Controller) SetupWithManager(mgr controllerruntime.Manager) error {
	return utilerrors.NewAggregate([]error{
		controllerruntime.NewControllerManagedBy(mgr).For(&kubeanclusteropsv1alpha1.KuBeanClusterOps{}).Complete(c),
		mgr.Add(c),
	})
}
