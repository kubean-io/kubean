package e2e

import (
	"context"
	"fmt"
	//"strings"

	tools "github.com/daocloud/kubean/test/tools"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = ginkgo.Describe("[create] Test K8 cluster All Info", func() {
	// 此处的config来自于从k8 cluster master node上scp取到
	//kubeconfig := tools.Path("demo_dev_config")
	kubeconfig := tools.Path("/tmp/kind_cluster.conf")
	NamespaceName := "kube-system"
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	tools.CheckError(err)
	kubeClient, err := kubernetes.NewForConfig(config)
	tools.CheckError(err)

	ginkgo.Context("When fetching kube-system pods status", func() {
		podList, err := kubeClient.CoreV1().Pods(NamespaceName).List(context.TODO(), metav1.ListOptions{})
		tools.CheckError(err)
		ginkgo.It("every k8 pod should be in running status", func() {
			for _, pod := range podList.Items {
				fmt.Println(pod.Name, string(pod.Status.Phase))
				gomega.Expect(string(pod.Status.Phase)).To(gomega.Equal("Running"))
			}
		})
	})

	ginkgo.Context("When fetching kube-system nodes status", func() {
		nodeList, err := kubeClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		tools.CheckError(err)
		ginkgo.It("every node should be in Ready status", func() {
			for _, node := range nodeList.Items {
				fmt.Println(node.Name, node.Status.Conditions[len(node.Status.Conditions)-1].Type)
				gomega.Expect(string(node.Status.Conditions[len(node.Status.Conditions)-1].Type)).To(gomega.Equal("Ready"))
			}
		})
	})

})
