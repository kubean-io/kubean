package kubeanOps_functions_e2e

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kubean-io/kubean/test/tools"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = ginkgo.Describe("kubean ops e2e test backofflimit=1", func() {
	config, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
	var _, currentFile, _, _ = runtime.Caller(0)
	var basepath = filepath.Dir(currentFile)
	kubeanNamespace := "kubean-system"

	ginkgo.Context("when installation fail then retry, set backofflimit=1", func() {
		opsFile := filepath.Join(basepath, "/backofflimit-clusterops/kubeanClusterOps.yml")
		kubeanClusterOpsNewName := "backofflimit1-clusterops-test"
		tools.UpdateOpsYml(kubeanClusterOpsNewName, opsFile)
		tools.UpdateBackoffLimit(1, opsFile)

		clusterInstallYamlsPath := "backofflimit-clusterops"
		installYamlPath := fmt.Sprint(tools.GetKuBeanPath(), clusterInstallYamlsPath)
		cmd := exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", installYamlPath)
		out, _ := tools.DoCmd(*cmd)
		fmt.Println("backofflimit=1 kubeanclusterOps: ", out.String())
		time.Sleep(150 * time.Second)
		kubeClient, err := kubernetes.NewForConfig(config)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")
		pods, _ := kubeClient.CoreV1().Pods(kubeanNamespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=kubean-%s-job", kubeanClusterOpsNewName),
		})
		fmt.Println(len(pods.Items))

		ginkgo.It("there is 2 kubeanclusterops related pod", func() {
			gomega.Expect(len(pods.Items)).Should(gomega.BeNumerically("==", 2))
		})
	})

})
