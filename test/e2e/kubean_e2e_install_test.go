package e2e

import (
	"context"
	//"fmt"
	//"strings"

	tools "github.com/daocloud/kubean/test/tools"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = ginkgo.Describe("[install] Test Kubean Operator All Info", func() {
	//kubeconfig := tools.Path("demo_dev_config")
	kubeconfig := tools.Path("/tmp/kind_cluster.conf")
	NamespaceName := "kubean-system"
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	tools.CheckError(err)
	kubeClient, err := kubernetes.NewForConfig(config)
	tools.CheckError(err)

	// pod, err := kubeClient.CoreV1().Pods(NamespaceName).List(context.TODO(), metav1.ListOptions{labels.Selector == "kubean"})
	// fmt.Println(pod)
	ginkgo.Context("When fetching kubean deployment info", func() {
		deploymentList, err := kubeClient.AppsV1().Deployments(NamespaceName).List(context.TODO(), metav1.ListOptions{})
		tools.CheckError(err)
		for _, dm := range deploymentList.Items {
			ginkgo.It("Kubean deployment should be ready", func() {
				gomega.Expect(dm.Status.ReadyReplicas).To(gomega.Equal(dm.Status.AvailableReplicas))
			})
		}
	})

})
