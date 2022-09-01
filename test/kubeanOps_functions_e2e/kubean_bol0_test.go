package kubeanOps_functions_e2e

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/kubean-io/kubean/test/tools"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = ginkgo.Describe("kubean ops e2e test backofflimit=0", func() {
	config, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
	kubeanNamespace := "kubean-system"

	ginkgo.Context("when installation fail then retry, set backofflimit=0", func() {
		kubeanClusterOpsName := "backofflimit0-clusterops-test"
		clusterInstallYamlsPath := "backofflimit-clusterops"
		installYamlPath := fmt.Sprint(tools.GetKuBeanPath(), clusterInstallYamlsPath)
		cmd := exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", installYamlPath)
		out, _ := tools.DoCmd(*cmd)
		fmt.Println("backofflimit=0 kubeanclusterops: ", out.String())
		time.Sleep(100 * time.Second)
		kubeClient, err := kubernetes.NewForConfig(config)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")
		pods, _ := kubeClient.CoreV1().Pods(kubeanNamespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=kubean-%s-job", kubeanClusterOpsName),
		})
		fmt.Println(len(pods.Items))

		ginkgo.It("there is 1 kubeanclusterops related pod", func() {
			gomega.Expect(len(pods.Items)).Should(gomega.BeNumerically("==", 1))
		})
	})

})
