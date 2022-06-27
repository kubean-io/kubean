package e2e

import (
	"context"

	"github.com/daocloud/kubean/test/tools"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = ginkgo.Describe("e2e test: kubean operation", func() {
	config, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
	kubeClient, err := kubernetes.NewForConfig(config)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")

	ginkgo.Context("When fetching kubean deployment info", func() {
		kubeanNamespace := "kubean-system"
		deploymentList, err := kubeClient.AppsV1().Deployments(kubeanNamespace).List(context.TODO(), metav1.ListOptions{})
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed kubean deployment info")
		for _, dm := range deploymentList.Items {
			ginkgo.It("Kubean deployment should be ready", func() {
				gomega.Expect(dm.Status.ReadyReplicas).To(gomega.Equal(dm.Status.AvailableReplicas))
			})
		}
	})

})
