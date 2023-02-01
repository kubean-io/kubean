package kubean_sonobouy_nightlye2e

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"os/exec"
	"strings"
	"time"

	"github.com/kubean-io/kubean/test/tools"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = ginkgo.Describe("e2e test cluster 1 master + 1 worker sonobouy check", func() {
	var masterSSH = fmt.Sprintf("root@%s", tools.Vmipaddr)
	var workerSSH = fmt.Sprintf("root@%s", tools.Vmipaddr2)

	// do cluster installation within docker
	ginkgo.Context("Install a sonobouy cluster using docker", func() {
		var localKubeConfigPath = "cluster1-config"
		var offlineConfigs tools.OfflineConfig
		var pod1Name = "nginx1"
		var pod2Name = "nginx2"
		var password = tools.VmPassword
		testClusterName := tools.TestClusterName
		nginxImage := "nginx:alpine"
		offlineFlag := tools.IsOffline
		offlineConfigs = tools.InitOfflineConfig()
		if strings.ToUpper(offlineFlag) == "TRUE" && strings.ToUpper(tools.Arch) == "ARM64" {
			nginxImage = offlineConfigs.NginxImageARM64
		}
		if strings.ToUpper(offlineFlag) == "TRUE" && strings.ToUpper(tools.Arch) == "AMD64" {
			nginxImage = offlineConfigs.NginxImageAMD64
		}
		klog.Info("nginx image is: ", nginxImage)
		klog.Info("offlineFlag is: ", offlineFlag)
		klog.Info("arch is: ", tools.Arch)

		ginkgo.It("Create cluster", func() {
			clusterInstallYamlsPath := "e2e-install-cluster-sonobouy"
			kubeanClusterOpsName := tools.ClusterOperationName
			installYamlPath := fmt.Sprint(tools.GetKuBeanPath(), clusterInstallYamlsPath)
			kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
			kindClient, err := kubernetes.NewForConfig(kindConfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")

			cmd := exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", installYamlPath)
			out, _ := tools.DoCmd(*cmd)
			klog.Info("create cluster result:", out.String())
			time.Sleep(10 * time.Second)

			// Check if the job and related pods have been created
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
			}, 120*time.Second, 5*time.Second).Should(gomega.BeTrue())

			jobPodName := pods.Items[0].Name
			tools.WaitKubeanJobPodToSuccess(kindClient, tools.KubeanNamespace, jobPodName, tools.PodStatusSucceeded)
			// Save testCluster kubeConfig to local path
			tools.SaveKubeConf(kindConfig, testClusterName, localKubeConfigPath)
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")
			tools.WaitPodSInKubeSystemBeRunning(cluster1Client, 1800)

			// check kube version before upgrade
			nodeList, _ := cluster1Client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			for _, node := range nodeList.Items {
				gomega.Expect(node.Status.NodeInfo.KubeletVersion).To(gomega.Equal(tools.OriginK8Version))
				gomega.Expect(node.Status.NodeInfo.KubeProxyVersion).To(gomega.Equal(tools.OriginK8Version))
			}
			if strings.ToUpper(offlineFlag) != "TRUE" {
				klog.Info("On line, sonobuoy check")
				tools.DoSonoBuoyCheckByPasswd(password, masterSSH)
			}
		})

		// check network configuration:
		// cat /proc/sys/net/ipv4/ip_forward: 1
		// cat /proc/sys/net/ipv4/tcp_tw_recycle: 0
		ginkgo.It("do network configurations checking", func() {
			masterCmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "cat", "/proc/sys/net/ipv4/ip_forward"})
			workerCmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{workerSSH, "cat", "/proc/sys/net/ipv4/ip_forward"})
			masterOut, _ := tools.NewDoCmd("sshpass", masterCmd...)
			klog.Info("out: ", masterOut.String())
			gomega.Expect(masterOut.String()).Should(gomega.ContainSubstring("1"))
			workerOut, _ := tools.NewDoCmd("sshpass", workerCmd...)
			klog.Info("out: ", workerOut.String())
			gomega.Expect(workerOut.String()).Should(gomega.ContainSubstring("1"))

			masterCmd = tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "cat", "/proc/sys/net/ipv4/tcp_tw_recycle"})
			workerCmd = tools.RemoteSSHCmdArrayByPasswd(password, []string{workerSSH, "cat", "/proc/sys/net/ipv4/tcp_tw_recycle"})
			masterOut, _ = tools.NewDoCmd("sshpass", masterCmd...)
			klog.Info("out: ", masterOut.String())
			gomega.Expect(masterOut.String()).Should(gomega.ContainSubstring("0"))
			workerOut, _ = tools.NewDoCmd("sshpass", workerCmd...)
			klog.Info("out: ", workerOut.String())
			gomega.Expect(workerOut.String()).Should(gomega.ContainSubstring("0"))
		})

		ginkgo.It("Support CNI: Calico", func() {
			//4. check calico (calico-node and calico-kube-controller)pod status: pod status should be "Running"
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")

			podList, _ := cluster1Client.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
			calico_pod_number := 0
			for _, pod := range podList.Items {
				if strings.Contains(pod.ObjectMeta.Name, "calico-node") || strings.Contains(pod.ObjectMeta.Name, "calico-kube-controllers") {
					gomega.Expect(string(pod.Status.Phase)).To(gomega.Equal("Running"))
					calico_pod_number += 1
				}
			}
			gomega.Expect(calico_pod_number).To(gomega.Equal(3))

			//5. check folder /opt/cni/bin contains  file "calico" and "calico-ipam" are exist in both master and worker node
			masterCmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "ls", "/opt/cni/bin/"})
			workerCmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{workerSSH, "ls", "/opt/cni/bin/"})
			masterOut, _ := tools.NewDoCmd("sshpass", masterCmd...)
			klog.Info("master cmd out: ", masterOut.String())
			gomega.Expect(masterOut.String()).Should(gomega.ContainSubstring("calico"))
			workerOut, _ := tools.NewDoCmd("sshpass", workerCmd...)
			klog.Info("worker cmd out: ", workerOut.String())
			gomega.Expect(workerOut.String()).Should(gomega.ContainSubstring("calico"))

			// check calicoctl
			masterCmd = tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "calicoctl", "version"})
			masterOut, _ = tools.NewDoCmd("sshpass", masterCmd...)
			klog.Info("check calicoctl: ", masterOut.String())
			gomega.Expect(masterOut.String()).Should(gomega.ContainSubstring("Client Version"))
			gomega.Expect(masterOut.String()).Should(gomega.ContainSubstring("Cluster Version"))
			gomega.Expect(masterOut.String()).Should(gomega.ContainSubstring("kubespray,kubeadm,kdd"))

			//6. check pod connection:
			tools.CreatePod(pod1Name, tools.DefaultNamespace, "node1", nginxImage, localKubeConfigPath)
			tools.CreatePod(pod2Name, tools.KubeSystemNamespace, "node2", nginxImage, localKubeConfigPath)
			pod1 := tools.WaitPodBeRunning(cluster1Client, tools.DefaultNamespace, pod1Name, 1000)
			pod2 := tools.WaitPodBeRunning(cluster1Client, tools.KubeSystemNamespace, pod2Name, 1000)
			tools.NodePingPodByPasswd(password, masterSSH, pod2.Status.PodIP)
			tools.NodePingPodByPasswd(password, workerSSH, pod1.Status.PodIP)
			tools.PodPingPodByPasswd(password, masterSSH, tools.DefaultNamespace, pod1Name, pod2.Status.PodIP)
			tools.PodPingPodByPasswd(password, masterSSH, tools.KubeSystemNamespace, pod2Name, pod1.Status.PodIP)
		})

		ginkgo.It("upgrade cluster Y version", func() {
			clusterInstallYamlsPath := "e2e-upgrade-cluster-y"
			kubeanClusterOpsName := "cluster1-upgrade-y"
			installYamlPath := fmt.Sprint(tools.GetKuBeanPath(), clusterInstallYamlsPath)
			kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
			kindClient, err := kubernetes.NewForConfig(kindConfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")

			cmd := exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", installYamlPath)
			out, _ := tools.DoCmd(*cmd)
			klog.Info("create cluster result:", out.String())
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
			}, 120*time.Second, 5*time.Second).Should(gomega.BeTrue())

			jobPodName := pods.Items[0].Name
			tools.WaitKubeanJobPodToSuccess(kindClient, tools.KubeanNamespace, jobPodName, tools.PodStatusSucceeded)
			tools.WaitPodSInKubeSystemBeRunning(cluster1Client, 1800)

			// kubectl version should be：tools.UpgradeK8Version_Y
			kubectlCmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "kubectl", "version", "--short"})
			kubectlOut, _ := tools.NewDoCmd("sshpass", kubectlCmd...)
			klog.Info(kubectlOut.String())
			gomega.Expect(kubectlOut.String()).Should(gomega.ContainSubstring(tools.UpgradeK8Version_Y))
			// node version should be：tools.UpgradeK8Version_Y
			nodeList, _ := cluster1Client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			for _, node := range nodeList.Items {
				gomega.Expect(node.Status.NodeInfo.KubeletVersion).To(gomega.Equal(tools.UpgradeK8Version_Y))
				gomega.Expect(node.Status.NodeInfo.KubeProxyVersion).To(gomega.Equal(tools.UpgradeK8Version_Y))
			}
		})

		ginkgo.It("upgrade cluster Z version", func() {
			clusterInstallYamlsPath := "e2e-upgrade-cluster-z"
			kubeanClusterOpsName := "cluster1-upgrade-z"

			installYamlPath := fmt.Sprint(tools.GetKuBeanPath(), clusterInstallYamlsPath)
			kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
			kindClient, err := kubernetes.NewForConfig(kindConfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")

			cmd := exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", installYamlPath)
			out, _ := tools.DoCmd(*cmd)
			klog.Info("create cluster result:", out.String())
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
			}, 120*time.Second, 5*time.Second).Should(gomega.BeTrue())

			jobPodName := pods.Items[0].Name
			tools.WaitKubeanJobPodToSuccess(kindClient, tools.KubeanNamespace, jobPodName, tools.PodStatusSucceeded)
			tools.WaitPodSInKubeSystemBeRunning(cluster1Client, 1800)

			kubectlCmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "kubectl", "version", "--short"})
			kubectlOut, _ := tools.NewDoCmd("sshpass", kubectlCmd...)
			klog.Info(kubectlOut.String())
			gomega.Expect(kubectlOut.String()).Should(gomega.ContainSubstring(tools.UpgradeK8Version_Z))

			nodeList, _ := cluster1Client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			for _, node := range nodeList.Items {
				gomega.Expect(node.Status.NodeInfo.KubeletVersion).To(gomega.Equal(tools.UpgradeK8Version_Z))
				gomega.Expect(node.Status.NodeInfo.KubeProxyVersion).To(gomega.Equal(tools.UpgradeK8Version_Z))
			}
		})
	})
})
