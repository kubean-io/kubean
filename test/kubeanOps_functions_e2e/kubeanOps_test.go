package kubeanOps_functions_e2e

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kubean-io/kubean/test/tools"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	kubeanClusterOpsClientSet "kubean.io/api/generated/kubeanclusterops/clientset/versioned"
)

var _ = ginkgo.Describe("kubean ops e2e test", func() {
	config, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")

	ginkgo.Context("when apply two jobs for one cluster hosts", func() {
		clusterInstallYamlsPath := "e2e-install-cluster"
		var _, currentFile, _, _ = runtime.Caller(0)
		var basepath = filepath.Dir(currentFile)
		opsFile := filepath.Join(basepath, "/e2e-install-cluster/kubeanClusterOps.yml")
		// step1 create the first ops
		clusterOpsName := "e2e-cluster1-ops-1st"
		tools.UpdateOpsYml(clusterOpsName, opsFile)
		cmd := exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", clusterInstallYamlsPath)
		ginkgo.GinkgoWriter.Printf("cmd: %s\n", cmd.String())
		var out, stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			ginkgo.GinkgoWriter.Printf("apply cmd error: %s\n", err.Error())
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), stderr.String())
		}
		clusterClientOpsSet, err := kubeanClusterOpsClientSet.NewForConfig(config)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")
		for {
			clusterOps, _ := clusterClientOpsSet.KubeanclusteropsV1alpha1().KuBeanClusterOps().Get(context.Background(), clusterOpsName, metav1.GetOptions{})
			status := string(clusterOps.Status.Status)
			ginkgo.GinkgoWriter.Printf("* wait for ops status:", status)
			if status == "Running" {
				ginkgo.It("the first ops should be running", func() {
					gomega.Expect(status).To(gomega.Equal("Running"))
				})
				break
			} else {
				time.Sleep(10 * time.Second)
			}
		}

		// step2 create the second ops
		clusterOpsNameSecond := "e2e-cluster1-ops-second"
		tools.UpdateOpsYml(clusterOpsNameSecond, opsFile)
		cmd = exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", clusterInstallYamlsPath)
		ginkgo.GinkgoWriter.Printf("cmd: %s\n", cmd.String())
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			ginkgo.GinkgoWriter.Printf("apply cmd error: %s\n", err.Error())
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), stderr.String())
		}
		for {
			clusterOpsSecond, _ := clusterClientOpsSet.KubeanclusteropsV1alpha1().KuBeanClusterOps().Get(context.Background(), clusterOpsNameSecond, metav1.GetOptions{})
			statusSecond := string(clusterOpsSecond.Status.Status)
			if statusSecond == "" {
				time.Sleep(10 * time.Second)
			} else {
				ginkgo.It("the first ops should be running", func() {
					gomega.Expect(statusSecond).To(gomega.Equal("Blocked"))
				})
				break

			}
		}
	})
})
