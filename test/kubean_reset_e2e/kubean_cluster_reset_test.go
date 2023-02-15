package kubean_reset_e2e

import (
	"bytes"
	"fmt"
	"github.com/kubean-io/kubean/test/tools"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os/exec"
)

var _ = ginkgo.Describe("e2e test cluster reset operation", func() {

	ginkgo.Context("Reset a cluster", func() {
		testClusterName := tools.TestClusterName
		clusterResetYamlsPath := "e2e-reset-cluster"
		kubeanClusterOpsName := "e2e-cluster1-reset"
		var masterSSH = fmt.Sprintf("root@%s", tools.Vmipaddr)
		var password = tools.VmPassword
		ginkgo.It("Kubean cluster podStatus should be Succeeded", func() {
			kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")

			// Create cluster by apply yaml
			tools.OperateClusterByYaml(clusterResetYamlsPath, kubeanClusterOpsName, kindConfig)
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

			newMasterCmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "ls", "-al", "/opt"})
			out2, _ := tools.NewDoCmd("sshpass", newMasterCmd...)
			klog.Info("5.3 CNI check1: execute ls -al /opt, the output should not contain cni")
			gomega.Expect(out2.String()).ShouldNot(gomega.ContainSubstring("cni"))

			newMasterCmd = tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "ls", "-al", "/etc"})
			out3, _ := tools.NewDoCmd("sshpass", newMasterCmd...)
			klog.Info("5.4 CNI check2: execute ls -al /etc,the output should not contain cni")
			gomega.Expect(out3.String()).ShouldNot(gomega.ContainSubstring("cni"))

			newMasterCmd = tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "ls", "-al", "/root"})
			out4, _ := tools.NewDoCmd("sshpass", newMasterCmd...)
			klog.Info("5.6 k8s config file check: execute ls -al /root, the output should not contain .kube")
			gomega.Expect(out4.String()).ShouldNot(gomega.ContainSubstring(".kube"))

			newMasterCmd = tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "ls", "-al", "/usr/local/bin"})
			out5, _ := tools.NewDoCmd("sshpass", newMasterCmd...)
			klog.Info("5.7 kubelet check: execute ls -al /usr/local/bin, the output should not contain kubelet")
			gomega.Expect(out5.String()).ShouldNot(gomega.ContainSubstring("kubelet"))
		})

		// Create cluster after reset
		ginkgo.It("Create cluster after reset with Docker CRI", func() {
			clusterInstallYamlsPath := "e2e-install-cluster-docker"
			kubeanClusterOpsName := tools.ClusterOperationName
			localKubeConfigPath := "cluster1-config-in-docker"
			kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")

			// modify hostname before reInstall
			cmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "hostnamectl", "set-hostname", "hello-kubean"})
			_, _ = tools.NewDoCmd("sshpass", cmd...)
			// check hostname after deploy: hostname should be hello-kubean
			var out bytes.Buffer
			cmd = tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "hostname"})
			out, _ = tools.NewDoCmd("sshpass", cmd...)
			fmt.Println("Fetched node hostname is: ", out.String())
			gomega.Expect(out.String()).Should(gomega.ContainSubstring("hello-kubean"))
			tools.OperateClusterByYaml(clusterInstallYamlsPath, kubeanClusterOpsName, kindConfig)
			tools.SaveKubeConf(kindConfig, testClusterName, localKubeConfigPath)
			cluster1Client := tools.GenerateClusterClient(localKubeConfigPath)
			tools.WaitPodSInKubeSystemBeRunning(cluster1Client, 3600)

			// check hostname after deploy: hostname should be hello-kubean
			cmd = tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "hostname"})
			out, _ = tools.NewDoCmd("sshpass", cmd...)
			fmt.Println("Fetched node hostname is: ", out.String())
			gomega.Expect(out.String()).Should(gomega.ContainSubstring("hello-kubean"))
		})

		ginkgo.It("[bug]Docker: when check docker functions", func() {
			masterCmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "docker", "info"})
			out, _ := tools.NewDoCmd("sshpass", masterCmd...)
			klog.Info("docker info to check if server running: ")
			gomega.Expect(out.String()).Should(gomega.ContainSubstring("/var/lib/docker"))
			gomega.Expect(out.String()).Should(gomega.ContainSubstring("Cgroup Driver: systemd"))

			masterCmd = tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "systemctl", "status", "docker"})
			out1, _ := tools.NewDoCmd("sshpass", masterCmd...)
			klog.Info("systemctl status containerd to check if docker running: ")
			gomega.Expect(out1.String()).Should(gomega.ContainSubstring("Active: active (running)"))
		})
	})
})
