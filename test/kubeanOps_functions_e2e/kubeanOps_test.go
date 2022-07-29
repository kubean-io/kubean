package kubeanOps_functions_e2e

import (
	"bytes"
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
	"k8s.io/client-go/tools/clientcmd"
	kubeanClusterOpsClientSet "kubean.io/api/generated/kubeanclusterops/clientset/versioned"
)

var _ = ginkgo.Describe("kubean ops e2e test", func() {
	config, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
	var _, currentFile, _, _ = runtime.Caller(0)
	var basepath = filepath.Dir(currentFile)
	clusterInstallYamlsPath := "e2e-install-cluster"
	opsFile := filepath.Join(basepath, "/e2e-install-cluster/kubeanClusterOps.yml")
	clusterClientOpsSet, err := kubeanClusterOpsClientSet.NewForConfig(config)
	var out, stderr bytes.Buffer

	ginkgo.Context("when apply two jobs for one cluster hosts", func() {
		// step1 create the first ops
		clusterOpsName := "e2e-cluster1-ops-1st"
		tools.UpdateOpsYml(clusterOpsName, opsFile)
		cmd := exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", clusterInstallYamlsPath)
		ginkgo.GinkgoWriter.Printf("cmd: %s\n", cmd.String())
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			ginkgo.GinkgoWriter.Printf("apply cmd error: %s\n", err.Error())
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), stderr.String())
		}
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")
		for {
			clusterOps, _ := clusterClientOpsSet.KubeanclusteropsV1alpha1().KuBeanClusterOps().Get(context.Background(), clusterOpsName, metav1.GetOptions{})
			status := string(clusterOps.Status.Status)
			ginkgo.GinkgoWriter.Printf("* wait for ops status: %s\n", status)
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
				ginkgo.It("the second ops should be blocked", func() {
					gomega.Expect(statusSecond).To(gomega.Equal("Blocked"))
				})
				break
			}
		}
		//delete all ops to prepare for the next testcase
		clusterOpsList, _ := clusterClientOpsSet.KubeanclusteropsV1alpha1().KuBeanClusterOps().List(context.Background(), metav1.ListOptions{})
		for _, ops := range clusterOpsList.Items {
			fmt.Println("delete cluster1 clusterOps: ", ops.Name)
			clusterClientOpsSet.KubeanclusteropsV1alpha1().KuBeanClusterOps().Delete(context.Background(), ops.Name, metav1.DeleteOptions{})
		}

	})

	ginkgo.Context("ClusterOps only save 5 copies for one cluster hosts, the oldest Ops will be removed", func() {
		// step1 create 5 ops for clusterName: cluster1
		for num := 1; num <= 5; num++ {
			clusterOpsName := fmt.Sprintf("e2e-cluster1-ops-copies%d", num)
			tools.UpdateOpsYml(clusterOpsName, opsFile)
			cmd := exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", clusterInstallYamlsPath)
			ginkgo.GinkgoWriter.Printf("cmd: %s\n", cmd.String())
			cmd.Stdout = &out
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), stderr.String())
			}
			time.Sleep(10 * time.Second)
		}
		// step2 check cluster1 should exists 5 ops
		clusterOpsList, _ := clusterClientOpsSet.KubeanclusteropsV1alpha1().KuBeanClusterOps().List(context.Background(), metav1.ListOptions{})
		fmt.Println("step2 - clusterOps count in cluster1: ", len(clusterOpsList.Items))
		ginkgo.It("the cluster1 should exists 5 ops", func() {
			gomega.Expect(len(clusterOpsList.Items)).Should(gomega.BeNumerically("==", 5))
		})
		// step3 create one more ops
		time.Sleep(2 * time.Second)
		clusterOpsName1 := "e2e-cluster1-ops-copies6"
		tools.UpdateOpsYml(clusterOpsName1, opsFile)
		cmd := exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", clusterInstallYamlsPath)
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), stderr.String())
		}
		// step4 check the oldest ops is removed
		for {
			_, err := clusterClientOpsSet.KubeanclusteropsV1alpha1().KuBeanClusterOps().Get(context.Background(), "e2e-cluster1-ops-copies1", metav1.GetOptions{})
			if err == nil {
				time.Sleep(5 * time.Second)
			} else {
				fmt.Println("cluster1 the oldest clusterOps: ", err)
				gomega.Expect(err).ShouldNot(gomega.BeNil())
				break
			}
		}
	})

})
