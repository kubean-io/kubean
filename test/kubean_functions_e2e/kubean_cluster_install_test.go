package kubean_functions_e2e

import (
	"context"
	"fmt"
	"k8s.io/klog/v2"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/kubean-io/kubean/test/tools"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = ginkgo.Describe("e2e test cluster operation", func() {

	localKubeConfigPath := tools.LocalKubeConfigPath
	var masterSSH = fmt.Sprintf("root@%s", tools.Vmipaddr)

	ginkgo.Context("when install a cluster", func() {
		var offlineConfigs tools.OfflineConfig
		var pod1Name = "nginx1"
		var svc1Name = "nginxsvc1"
		var password = tools.VmPassword
		clusterInstallYamlsPath := "e2e-install-cluster"
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
		kubeanClusterOpsName := tools.ClusterOperationName
		testClusterName := tools.TestClusterName
		ginkgo.It("kubean cluster podStatus should be Succeeded", func() {
			kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
			kindClient, err := kubernetes.NewForConfig(kindConfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")
			// Create yaml for kuBean CR and related configuration
			klog.Info("kubeanClusterOpsName is :", kubeanClusterOpsName)
			installYamlPath := fmt.Sprint(tools.GetKuBeanPath(), clusterInstallYamlsPath)
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
			}, 60*time.Second, 5*time.Second).Should(gomega.BeTrue())

			jobPodName := pods.Items[0].Name
			tools.WaitKubeanJobPodToSuccess(kindClient, tools.KubeanNamespace, jobPodName, tools.PodStatusSucceeded)

			// Save testCluster kubeConfig to local path
			tools.SaveKubeConf(kindConfig, testClusterName, localKubeConfigPath)
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")
			// Wait all pods in kube-syste to be Running
			tools.WaitPodSInKubeSystemBeRunning(cluster1Client, 1800)
		})

		ginkgo.It("set-hostname to node1", func() {
			// hostname after deploy: hostname should be node1
			hostnamecmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "hostname"})
			hostnameout, _ := tools.NewDoCmd("sshpass", hostnamecmd...)

			fmt.Println("hostname: ", hostnameout.String())
			gomega.Expect(hostnameout.String()).Should(gomega.ContainSubstring("node1"))
		})

		ginkgo.It("support iptables", func() {
			getProxModecmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "kubectl", "get", "cm", "-n", "kube-system", "kube-proxy", "-o", "jsonpath='{.data.config\\.conf}{\"\\n\"}'"})
			proxOut, _ := tools.NewDoCmd("sshpass", getProxModecmd...)

			fmt.Println("Prox Mode:", proxOut.String())
			gomega.Expect(proxOut.String()).Should(gomega.ContainSubstring("mode: iptables"))
		})

		ginkgo.It("systemctl status containerd to check if containerd running: ", func() {
			masterCmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "nerdctl", "info"})
			out, _ := tools.NewDoCmd("sshpass", masterCmd...)

			gomega.Expect(out.String()).Should(gomega.ContainSubstring("k8s.io"))
			gomega.Expect(out.String()).Should(gomega.ContainSubstring("Cgroup Driver: systemd"))

			masterCmd = tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "systemctl", "status", "containerd"})
			out1, _ := tools.NewDoCmd("sshpass", masterCmd...)

			gomega.Expect(out1.String()).Should(gomega.ContainSubstring("/etc/systemd/system/containerd.service;"))
			gomega.Expect(out1.String()).Should(gomega.ContainSubstring("Active: active (running)"))
		})

		ginkgo.It("nginx service can be request", func() {

			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")

			//check a test nginx svc for network check
			nginx1Cmd := exec.Command("kubectl", "run", pod1Name, "-n", tools.DefaultNamespace, "--image", nginxImage, "--kubeconfig", localKubeConfigPath, "--env", "NodeName=node1")
			nginx1CmdOut, err1 := tools.DoErrCmd(*nginx1Cmd)
			klog.Info("create %s :", nginx1CmdOut.String(), err1.String())
			tools.WaitPodBeRunning(cluster1Client, tools.DefaultNamespace, pod1Name, 1000)
			service1Cmd := exec.Command("kubectl", "expose", "pod", pod1Name, "-n", tools.DefaultNamespace, "--port", "18081", "--target-port", "80", "--type", "NodePort", "--name", svc1Name, "--kubeconfig", localKubeConfigPath)
			service1CmdOut, err1 := tools.DoErrCmd(*service1Cmd)
			klog.Info("create service result:", service1CmdOut.String(), err1.String())
			svc, err := cluster1Client.CoreV1().Services(tools.DefaultNamespace).Get(context.Background(), svc1Name, metav1.GetOptions{})
			port := svc.Spec.Ports[0].NodePort
			time.Sleep(10 * time.Second)
			tools.SvcCurl(tools.Vmipaddr, port, "Welcome to nginx!", 60)
		})

		ginkgo.It("Support calico: kube_pods_subnet ", func() {
			//This case need the nginx pod created in the upper case
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")

			//the pod set was 192.168.128.0/20, so the available pod ip range is 192.168.128.1 ~ 192.168.143.255
			pod1 := tools.WaitPodBeRunning(cluster1Client, tools.DefaultNamespace, pod1Name, 1000)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed to get pods")

			fmt.Println("pod ip is: ", pod1.Status.PodIP)
			ipSplitArr := strings.Split(pod1.Status.PodIP, ".")
			gomega.Expect(len(ipSplitArr)).Should(gomega.Equal(4))

			ipSub1, err := strconv.Atoi(ipSplitArr[0])
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "ip split conversion failed")
			ipSub2, err := strconv.Atoi(ipSplitArr[1])
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "ip split conversion failed")
			ipSub3, err := strconv.Atoi(ipSplitArr[2])
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "ip split conversion failed")

			gomega.Expect(ipSub1).Should(gomega.Equal(192))
			gomega.Expect(ipSub2).Should(gomega.Equal(168))
			gomega.Expect(ipSub3 >= 128).Should(gomega.BeTrue())
			gomega.Expect(ipSub3 <= 143).Should(gomega.BeTrue())
		})

		ginkgo.It("Support calico: kube_service_addresses ", func() {
			//This case need the nginx pod created in the upper case
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")

			//the pod set was 10.96.0.0/12, so the available svc ip range is 10.96.0.1 ~ 10.111.255.255
			svc1, err := cluster1Client.CoreV1().Services(tools.DefaultNamespace).Get(context.Background(), svc1Name, metav1.GetOptions{})
			klog.Info("svc ip ia: ", svc1.Spec.ClusterIP)
			ipSplitArr := strings.Split(svc1.Spec.ClusterIP, ".")
			gomega.Expect(len(ipSplitArr)).Should(gomega.Equal(4))

			ipSub1, err := strconv.Atoi(ipSplitArr[0])
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "ip split conversion failed")
			ipSub2, err := strconv.Atoi(ipSplitArr[1])
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "ip split conversion failed")

			gomega.Expect(ipSub1).Should(gomega.Equal(10))
			gomega.Expect(ipSub2 >= 96).Should(gomega.BeTrue())
		})
	})
})
