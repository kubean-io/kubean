package clusterops

import (
	"context"
	"fmt"
	"time"

	"github.com/daocloud/kubean/pkg/apis"
	kubeanclusterv1alpha1 "github.com/daocloud/kubean/pkg/apis/kubeancluster/v1alpha1"
	kubeanclusteropsv1alpha1 "github.com/daocloud/kubean/pkg/apis/kubeanclusterops/v1alpha1"
	kubeanClusterClientSet "github.com/daocloud/kubean/pkg/generated/kubeancluster/clientset/versioned"
	kubeanClusterOpsClientSet "github.com/daocloud/kubean/pkg/generated/kubeanclusterops/clientset/versioned"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const RequeueAfter = time.Second * 1

type Controller struct {
	client.Client
	ClientSet           *kubernetes.Clientset
	KubeanClusterSet    *kubeanClusterClientSet.Clientset
	KubeanClusterOpsSet *kubeanClusterOpsClientSet.Clientset
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
		klog.Error(err)
		return controllerruntime.Result{RequeueAfter: RequeueAfter}, err
	}
	cluster, err := c.GetKuBeanCluster(clusterOps)
	if err != nil {
		klog.Error(err)
		return controllerruntime.Result{RequeueAfter: RequeueAfter}, err
	}
	needRequeue, err := c.BackUpDataRef(clusterOps, cluster)
	if err != nil {
		klog.Error(err)
		return controllerruntime.Result{RequeueAfter: RequeueAfter}, err
	}
	if needRequeue {
		// something(spec) updated ,so continue the next loop.
		return controllerruntime.Result{RequeueAfter: RequeueAfter}, nil
	}
	return controllerruntime.Result{Requeue: false}, nil
}

// GetKuBeanCluster fetch the cluster which clusterOps belongs to.
func (c *Controller) GetKuBeanCluster(clusterOps *kubeanclusteropsv1alpha1.KuBeanClusterOps) (*kubeanclusterv1alpha1.KuBeanCluster, error) {
	// cluster has many clusterOps.
	return c.KubeanClusterSet.KubeanclusterV1alpha1().KuBeanClusters().Get(context.Background(), clusterOps.Spec.KuBeanCluster, metav1.GetOptions{})
}

func (c *Controller) CopyConfigMap(oldConfigMapRef *apis.ConfigMapRef, newName string) (*corev1.ConfigMap, error) {
	// todo ownreferences
	oldConfigMap, err := c.ClientSet.CoreV1().ConfigMaps(oldConfigMapRef.NameSpace).Get(context.Background(), oldConfigMapRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	newConfigMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      newName,
			Namespace: oldConfigMapRef.NameSpace,
		},
		Data: oldConfigMap.Data,
	}
	newConfigMap, err = c.ClientSet.CoreV1().ConfigMaps(newConfigMap.Namespace).Create(context.Background(), newConfigMap, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return newConfigMap, nil
}

func (c *Controller) CopySecret(oldSecretRef *apis.SecretRef, newName string) (*corev1.Secret, error) {
	// todo ownreferences
	oldSecret, err := c.ClientSet.CoreV1().Secrets(oldSecretRef.NameSpace).Get(context.Background(), oldSecretRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	newSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      newName,
			Namespace: oldSecretRef.NameSpace,
		},
		Data: oldSecret.Data,
	}
	newSecret, err = c.ClientSet.CoreV1().Secrets(newSecret.Namespace).Create(context.Background(), newSecret, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return newSecret, nil
}

// BackUpDataRef perform the backup of configRef and secretRef and return (needRequeue,error).
func (c *Controller) BackUpDataRef(clusterOps *kubeanclusteropsv1alpha1.KuBeanClusterOps, cluster *kubeanclusterv1alpha1.KuBeanCluster) (bool, error) {
	timestamp := fmt.Sprintf("-%d", time.Now().UnixMilli())
	if cluster.Spec.HostsConfRef.IsEmpty() || cluster.Spec.VarsConfRef.IsEmpty() || cluster.Spec.SSHAuthRef.IsEmpty() {
		return false, fmt.Errorf("cluster %s DataRef has empty value", cluster.Name)
	}
	if clusterOps.Spec.HostsConfRef.IsEmpty() {
		newConfigMap, err := c.CopyConfigMap(cluster.Spec.HostsConfRef, cluster.Spec.HostsConfRef.Name+timestamp)
		if err != nil {
			return false, err
		}
		clusterOps.Spec.HostsConfRef = &apis.ConfigMapRef{
			NameSpace: newConfigMap.Namespace,
			Name:      newConfigMap.Name,
		}
		if err := c.Client.Update(context.Background(), clusterOps); err != nil {
			return false, err
		}
		return true, nil
	}
	if clusterOps.Spec.VarsConfRef.IsEmpty() {
		newConfigMap, err := c.CopyConfigMap(cluster.Spec.VarsConfRef, cluster.Spec.VarsConfRef.Name+timestamp)
		if err != nil {
			return false, err
		}
		clusterOps.Spec.VarsConfRef = &apis.ConfigMapRef{
			NameSpace: newConfigMap.Namespace,
			Name:      newConfigMap.Name,
		}
		if err := c.Client.Update(context.Background(), clusterOps); err != nil {
			return false, err
		}
		return true, nil
	}
	if clusterOps.Spec.SSHAuthRef.IsEmpty() {
		newSecret, err := c.CopySecret(cluster.Spec.SSHAuthRef, cluster.Spec.SSHAuthRef.Name+timestamp)
		if err != nil {
			return false, err
		}
		clusterOps.Spec.SSHAuthRef = &apis.SecretRef{
			NameSpace: newSecret.Namespace,
			Name:      newSecret.Name,
		}
		if err := c.Client.Update(context.Background(), clusterOps); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil // needRequeue,err
}

func (c *Controller) SetupWithManager(mgr controllerruntime.Manager) error {
	return utilerrors.NewAggregate([]error{
		controllerruntime.NewControllerManagedBy(mgr).For(&kubeanclusteropsv1alpha1.KuBeanClusterOps{}).Complete(c),
		mgr.Add(c),
	})
}
