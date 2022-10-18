package kubean_reset_e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/kubean-io/kubean/test/tools"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	kubeanClusterClientSet "kubean.io/api/generated/cluster/clientset/versioned"
)

var preCmdArray = []string{"-p", "root", "ssh", "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no"}

var _ = ginkgo.Describe("e2e test cluster reset operation", func() {

	config, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
	kubeClient, err := kubernetes.NewForConfig(config)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")
	var masterSSH = fmt.Sprintf("root@%s", tools.Vmipaddr)

	defer ginkgo.GinkgoRecover()

	// do cluster reset
	ginkgo.Context("when reset a cluster", func() {
		clusterInstallYamlsPath := "e2e-reset-cluster"
		kubeanNamespace := "kubean-system"
		kubeanClusterOpsName := "e2e-cluster1-reset"

		// Create yaml for kuBean CR and related configuration
		installYamlPath := fmt.Sprint(tools.GetKuBeanPath(), clusterInstallYamlsPath)
		cmd := exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", installYamlPath)
		ginkgo.GinkgoWriter.Printf("cmd: %s\n", cmd.String())
		var out, stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			ginkgo.GinkgoWriter.Printf("apply cmd error: %s\n", err.Error())
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), stderr.String())
		}

		// Check if reset job and related pods have been created
		config, err = clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
		kubeClient, err = kubernetes.NewForConfig(config)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")
		time.Sleep(30 * time.Second)
		pods, _ := kubeClient.CoreV1().Pods(kubeanNamespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=kubean-%s-job", kubeanClusterOpsName),
		})
		gomega.Expect(len(pods.Items)).NotTo(gomega.Equal(0))
		jobPodName := pods.Items[0].Name

		// Wait for reset job-related pod status to be succeeded
		for {
			pod, err := kubeClient.CoreV1().Pods(kubeanNamespace).Get(context.Background(), jobPodName, metav1.GetOptions{})
			ginkgo.GinkgoWriter.Printf("* wait for reset job related pod[%s] status: %s\n", pod.Name, pod.Status.Phase)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed get job related pod")
			podStatus := string(pod.Status.Phase)
			if podStatus == "Succeeded" || podStatus == "Failed" {
				ginkgo.It("cluster podStatus should be Succeeded", func() {
					gomega.Expect(podStatus).To(gomega.Equal("Succeeded"))
				})
				break
			}
			time.Sleep(1 * time.Minute)
		}

		// after reest login node， check node functions
		ginkgo.Context("Containerd: login node, check node reset:", func() {
			masterSSH := fmt.Sprintf("root@%s", tools.Vmipaddr)
			masterCmd := exec.Command("sshpass", "-p", "root", "ssh", "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", masterSSH, "kubectl")
			_, err := tools.DoErrCmd(*masterCmd)
			ginkgo.It("5.1 kubectl check: execute kubectl, output should contain command not found", func() {
				gomega.Expect(err.String()).Should(gomega.ContainSubstring("command not found"))
			})

			masterCmd = exec.Command("sshpass", "-p", "root", "ssh", "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", masterSSH, "systemctl", "status", "containerd.service")
			_, err1 := tools.DoErrCmd(*masterCmd)
			fmt.Println(err1.String())
			ginkgo.It("5.2 CRI check: execute systemctl status containerd.service", func() {
				gomega.Expect(err1).ShouldNot(gomega.BeNil())
			})

			newMasterCmd := tools.RemoteSSHCmdArray([]string{masterSSH, "ls", "-al", "/opt"})
			out2, _ := tools.NewDoCmd("sshpass", newMasterCmd...)
			ginkgo.It("5.3 CNI check1: execute ls -al /opt, the output should not contain cni", func() {
				gomega.Expect(out2.String()).ShouldNot(gomega.ContainSubstring("cni"))
			})

			newMasterCmd = tools.RemoteSSHCmdArray([]string{masterSSH, "ls", "-al", "/etc"})
			out3, _ := tools.NewDoCmd("sshpass", newMasterCmd...)
			ginkgo.It("5.4 CNI check2: execute ls -al /etc,the output should not contain cni", func() {
				gomega.Expect(out3.String()).ShouldNot(gomega.ContainSubstring("cni"))
			})

			newMasterCmd = tools.RemoteSSHCmdArray([]string{masterSSH, "ls", "-al", "/root"})
			out4, _ := tools.NewDoCmd("sshpass", newMasterCmd...)
			ginkgo.It("5.6 k8s config file check: execute ls -al /root, the output should not contain .kube", func() {
				gomega.Expect(out4.String()).ShouldNot(gomega.ContainSubstring(".kube"))
			})

			newMasterCmd = tools.RemoteSSHCmdArray([]string{masterSSH, "ls", "-al", "/usr/local/bin"})
			out5, _ := tools.NewDoCmd("sshpass", newMasterCmd...)
			ginkgo.It("5.7 kubelet check: execute ls -al /usr/local/bin, the output should not contain kubelet", func() {
				gomega.Expect(out5.String()).ShouldNot(gomega.ContainSubstring("kubelet"))
			})
		})
	})

	// do cluster installation within docker
	ginkgo.Context("when install a cluster using docker", func() {
		clusterInstallYamlsPath := "e2e-install-cluster-docker"
		kubeanNamespace := "kubean-system"
		kubeanClusterOpsName := "e2e-install-cluster-docker"
		localKubeConfigPath := "cluster1-config-in-docker"

		// modify hostname
		cmd := tools.RemoteSSHCmdArray([]string{masterSSH, "hostnamectl", "set-hostname", "hello-kubean"})
		_, _ = tools.NewDoCmd("sshpass", cmd...)

		// Create yaml for kuBean CR and related configuration
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

		// Check if the job and related pods have been created
		time.Sleep(30 * time.Second)
		pods, _ := kubeClient.CoreV1().Pods(kubeanNamespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=kubean-%s-job", kubeanClusterOpsName),
		})
		gomega.Expect(len(pods.Items)).NotTo(gomega.Equal(0))
		jobPodName := pods.Items[0].Name

		// Wait for job-related pod status to be succeeded
		for {
			pod, err := kubeClient.CoreV1().Pods(kubeanNamespace).Get(context.Background(), jobPodName, metav1.GetOptions{})
			ginkgo.GinkgoWriter.Printf("* wait for install job using docker related pod[%s] status: %s\n", pod.Name, pod.Status.Phase)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed get job related pod")
			podStatus := string(pod.Status.Phase)
			if podStatus == "Succeeded" || podStatus == "Failed" {
				ginkgo.It("cluster podStatus should be Succeeded", func() {
					gomega.Expect(podStatus).To(gomega.Equal("Succeeded"))
				})
				break
			}
			time.Sleep(1 * time.Minute)
		}

		clusterClientSet, err := kubeanClusterClientSet.NewForConfig(config)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")

		// from Cluster: cluster1 get kubeconfRef: name: cluster1-kubeconf namespace: kubean-system
		cluster1, err := clusterClientSet.KubeanV1alpha1().Clusters().Get(context.Background(), "cluster1", metav1.GetOptions{})
		fmt.Println("Name:", cluster1.Spec.KubeConfRef.Name, "NameSpace:", cluster1.Spec.KubeConfRef.NameSpace)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed to get Cluster")

		// get configmap
		kubeClient, _ := kubernetes.NewForConfig(config)
		cluster1CF, _ := kubeClient.CoreV1().ConfigMaps(cluster1.Spec.KubeConfRef.NameSpace).Get(context.Background(), cluster1.Spec.KubeConfRef.Name, metav1.GetOptions{})
		err1 := os.WriteFile(localKubeConfigPath, []byte(cluster1CF.Data["config"]), 0666)
		gomega.ExpectWithOffset(2, err1).NotTo(gomega.HaveOccurred(), "failed to write localKubeConfigPath")

		// check kube-system pod status
		ginkgo.Context("When fetching kube-system pods status", func() {
			podList, err := kubeClient.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed to check kube-system pod status")
			ginkgo.It("every pod should be in running status", func() {
				for _, pod := range podList.Items {
					fmt.Println(pod.Name, string(pod.Status.Phase))
					gomega.Expect(string(pod.Status.Phase)).To(gomega.Equal("Running"))
				}
			})
		})

		// check hostname after deploy: hostname should be hello-kubean
		cmd = tools.RemoteSSHCmdArray([]string{masterSSH, "hostname"})
		out, _ = tools.NewDoCmd("sshpass", cmd...)
		ginkgo.It("set-hostname to hello-kubean", func() {
			fmt.Println("hostname: ", out.String())
			gomega.Expect(out.String()).Should(gomega.ContainSubstring("hello-kubean"))
		})
	})

	// check docker functions
	ginkgo.Context("Docker: when check docker functions", func() {
		masterCmd := tools.RemoteSSHCmdArray([]string{masterSSH, "docker", "info"})
		out, _ := tools.NewDoCmd("sshpass", masterCmd...)
		ginkgo.It("docker info to check if server running: ", func() {
			gomega.Expect(out.String()).Should(gomega.ContainSubstring("/var/lib/docker"))
			gomega.Expect(out.String()).Should(gomega.ContainSubstring("Cgroup Driver: systemd"))
		})

		masterCmd = tools.RemoteSSHCmdArray([]string{masterSSH, "systemctl", "status", "docker"})
		out1, _ := tools.NewDoCmd("sshpass", masterCmd...)
		ginkgo.It("systemctl status containerd to check if docker running: ", func() {
			gomega.Expect(out1.String()).Should(gomega.ContainSubstring("Active: active (running)"))
		})
	})
})
