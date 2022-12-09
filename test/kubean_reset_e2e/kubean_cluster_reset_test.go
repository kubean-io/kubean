package kubean_reset_e2e

import (
	"bytes"
	"context"
	"fmt"
	"k8s.io/klog/v2"
	"os/exec"
	"time"

	"github.com/kubean-io/kubean/test/tools"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = ginkgo.Describe("e2e test cluster reset operation", func() {

	ginkgo.Context("Reset a cluster", func() {
		testClusterName := tools.TestClusterName
		clusterResetYamlsPath := "e2e-reset-cluster"
		kubeanClusterOpsName := "e2e-cluster1-reset"
		var masterSSH = fmt.Sprintf("root@%s", tools.Vmipaddr)

		ginkgo.It("Kubean cluster podStatus should be Succeeded", func() {
			kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
			kindClient, err := kubernetes.NewForConfig(kindConfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")

			// Start reset cluster job
			resetYamlPath := fmt.Sprint(tools.GetKuBeanPath(), clusterResetYamlsPath)
			cmd := exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", resetYamlPath)
			ginkgo.GinkgoWriter.Printf("cmd: %s\n", cmd.String())
			out, _ := tools.DoCmd(*cmd)
			klog.Info("reset cluster result:", out.String())
			time.Sleep(10 * time.Second)

			// Fetch the job-related pod
			pods := &corev1.PodList{}
			klog.Info("Wait job related pod to be created")
			labelStr := fmt.Sprintf("job-name=kubean-%s-job", kubeanClusterOpsName)
			klog.Info("label is: ", labelStr)
			gomega.Eventually(func() bool {
				pods, _ = kindClient.CoreV1().Pods(tools.KubeanNamespace).List(context.Background(), metav1.ListOptions{
					LabelSelector: labelStr,
				})
				if len(pods.Items) > 0 {
					return true
				}
				return false
			}, 60*time.Second, 5*time.Second).Should(gomega.BeTrue())
			jobPodName := pods.Items[0].Name

			// Wait job-related pod to be succeeded
			tools.WaitKubeanJobPodToSuccess(kindClient, tools.KubeanNamespace, jobPodName, tools.PodStatusSucceeded)

		})

		// after reest login node， check node functions
		ginkgo.It("Check node is retested cleanly", func() {
			masterSSH := fmt.Sprintf("root@%s", tools.Vmipaddr)
			masterCmd := exec.Command("sshpass", "-p", "root", "ssh", "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", masterSSH, "command", "-v", "kubectl")
			outPut, _ := tools.DoErrCmd(*masterCmd)
			klog.Info("5.1 kubectl check: execute kubectl, output should contain command not found------")
			gomega.Expect(outPut.String()).Should(gomega.BeEmpty())

			masterCmd = exec.Command("sshpass", "-p", "root", "ssh", "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", masterSSH, "systemctl", "status", "containerd.service")
			_, err1 := tools.DoErrCmd(*masterCmd)
			fmt.Println(err1.String())
			klog.Info("5.2 CRI check: execute systemctl status containerd.service")
			gomega.Expect(err1).ShouldNot(gomega.BeNil())

			newMasterCmd := tools.RemoteSSHCmdArray([]string{masterSSH, "ls", "-al", "/opt"})
			out2, _ := tools.NewDoCmd("sshpass", newMasterCmd...)
			klog.Info("5.3 CNI check1: execute ls -al /opt, the output should not contain cni")
			gomega.Expect(out2.String()).ShouldNot(gomega.ContainSubstring("cni"))

			newMasterCmd = tools.RemoteSSHCmdArray([]string{masterSSH, "ls", "-al", "/etc"})
			out3, _ := tools.NewDoCmd("sshpass", newMasterCmd...)
			klog.Info("5.4 CNI check2: execute ls -al /etc,the output should not contain cni")
			gomega.Expect(out3.String()).ShouldNot(gomega.ContainSubstring("cni"))

			newMasterCmd = tools.RemoteSSHCmdArray([]string{masterSSH, "ls", "-al", "/root"})
			out4, _ := tools.NewDoCmd("sshpass", newMasterCmd...)
			klog.Info("5.6 k8s config file check: execute ls -al /root, the output should not contain .kube")
			gomega.Expect(out4.String()).ShouldNot(gomega.ContainSubstring(".kube"))

			newMasterCmd = tools.RemoteSSHCmdArray([]string{masterSSH, "ls", "-al", "/usr/local/bin"})
			out5, _ := tools.NewDoCmd("sshpass", newMasterCmd...)
			klog.Info("5.7 kubelet check: execute ls -al /usr/local/bin, the output should not contain kubelet")
			gomega.Expect(out5.String()).ShouldNot(gomega.ContainSubstring("kubelet"))
		})

		// Create cluster after reset
		//issue link: https://github.com/kubean-io/kubean/issues/295
		ginkgo.It("Create cluster after reset with Docker CRI", func() {
			kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
			kindClient, err := kubernetes.NewForConfig(kindConfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")

			clusterInstallYamlsPath := "e2e-install-cluster-docker"
			kubeanClusterOpsName := tools.ClusterOperationName
			localKubeConfigPath := "cluster1-config-in-docker"

			// modify hostname before reInstall
			cmd := tools.RemoteSSHCmdArray([]string{masterSSH, "hostnamectl", "set-hostname", "hello-kubean"})
			_, _ = tools.NewDoCmd("sshpass", cmd...)

			//Create yaml for kuBean CR and related configuration
			installYamlPath := fmt.Sprint(tools.GetKuBeanPath(), clusterInstallYamlsPath)
			cmd1 := exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", installYamlPath)
			ginkgo.GinkgoWriter.Printf("cmd: %s\n", cmd1.String())
			var out, stderr bytes.Buffer
			cmd1.Stdout = &out
			cmd1.Stderr = &stderr
			if err := cmd1.Run(); err != nil {
				ginkgo.GinkgoWriter.Printf("apply cmd error: %s\n", err.Error())
				gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), stderr.String())
			}

			// Fetch the job-related pod
			time.Sleep(10 * time.Second)
			pods := &corev1.PodList{}
			klog.Info("Wait job related pod to be created")
			labelStr := fmt.Sprintf("job-name=kubean-%s-job", kubeanClusterOpsName)
			klog.Info("label is: ", labelStr)
			gomega.Eventually(func() bool {
				pods, _ = kindClient.CoreV1().Pods(tools.KubeanNamespace).List(context.Background(), metav1.ListOptions{
					LabelSelector: labelStr,
				})
				if len(pods.Items) > 0 {
					return true
				}
				return false
			}, 60*time.Second, 5*time.Second).Should(gomega.BeTrue())
			jobPodName := pods.Items[0].Name
			// Wait for job-related pod status to be succeeded
			tools.WaitKubeanJobPodToSuccess(kindClient, tools.KubeanNamespace, jobPodName, tools.PodStatusSucceeded)
			// Save testCluster kubeConfig to local path
			tools.SaveKubeConf(kindConfig, testClusterName, localKubeConfigPath)
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")
			// Wait all pods in kube-syste to be Running
			tools.WaitPodSInKubeSystemBeRunning(cluster1Client, 1800)

			// check hostname after deploy: hostname should be hello-kubean
			cmd = tools.RemoteSSHCmdArray([]string{masterSSH, "hostname"})
			out, _ = tools.NewDoCmd("sshpass", cmd...)
			fmt.Println("Fetched node hostname is: ", out.String())
			gomega.Expect(out.String()).Should(gomega.ContainSubstring("hello-kubean"))
		})

		//issue link: https://github.com/kubean-io/kubean/issues/295
		ginkgo.It("Docker: when check docker functions", func() {
			masterCmd := tools.RemoteSSHCmdArray([]string{masterSSH, "docker", "info"})
			out, _ := tools.NewDoCmd("sshpass", masterCmd...)
			klog.Info("docker info to check if server running: ")
			gomega.Expect(out.String()).Should(gomega.ContainSubstring("/var/lib/docker"))
			gomega.Expect(out.String()).Should(gomega.ContainSubstring("Cgroup Driver: systemd"))

			masterCmd = tools.RemoteSSHCmdArray([]string{masterSSH, "systemctl", "status", "docker"})
			out1, _ := tools.NewDoCmd("sshpass", masterCmd...)
			klog.Info("systemctl status containerd to check if docker running: ")
			gomega.Expect(out1.String()).Should(gomega.ContainSubstring("Active: active (running)"))
		})
	})
})
