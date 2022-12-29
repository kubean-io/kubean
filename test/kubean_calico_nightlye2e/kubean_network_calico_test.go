package kubean_calico_nightlye2e

import (
	"context"
	"fmt"
	"github.com/kubean-io/kubean/test/tools"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os/exec"
	"strings"
	"time"
)

/*var _, currentFile, _, _ = runtime.Caller(0)
var basepath = filepath.Dir(currentFile)*/

var _ = ginkgo.Describe("Calico single stack tunnel: IPIP_ALWAYS", func() {

	localKubeConfigPath := "calico-single-stack.config"
	var masterSSH = fmt.Sprintf("root@%s", tools.Vmipaddr)
	var workerSSH = fmt.Sprintf("root@%s", tools.Vmipaddr2)
	ginkgo.Context("when install a cluster based on calico single stack", func() {
		var offlineConfigs tools.OfflineConfig
		var pod1Name = "nginx1"
		var pod2Name = "nginx2"
		var password = tools.VmPassword
		//var svc1Name = "nginxsvc1"
		//kubeanNamespace := tools.KubeanNamespace
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

		ginkgo.It("Create cluster and all kube-system pods be running", func() {
			clusterInstallYamlsPath := "e2e-install-calico-cluster"
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
			// Wait all pods in kube-syste to be Running
			tools.WaitPodSInKubeSystemBeRunning(cluster1Client, 1800)
		})

		ginkgo.It("Calico pod check", func() {
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")

			podList, _ := cluster1Client.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
			calico_pod_number := 0
			for _, pod := range podList.Items {
				if strings.Contains(pod.ObjectMeta.Name, "calico-node") || strings.Contains(pod.ObjectMeta.Name, "calico_kube_controller") {
					gomega.Expect(string(pod.Status.Phase)).To(gomega.Equal("Running"))
					calico_pod_number += 1
				}
			}
			gomega.Expect(calico_pod_number).To(gomega.Equal(2))
		})

		ginkgo.It("check calico tunnel valid", func() {
			tunal_mode := tools.OtherLabel
			klog.Info("tunal_mode is: ", tunal_mode)
			poolCmd := tools.RemoteSSHCmdArray([]string{masterSSH, "calicoctl", "get", "ippools", "--output=go-template=\"{{range .}}{{range .Items}}{{.ObjectMeta.Name}}{{end}}{{end}}\""})
			poolName, _ := tools.NewDoCmd("sshpass", poolCmd...)
			fmt.Println("check poolName: ", poolName.String())

			ipmodeCmd := tools.RemoteSSHCmdArray([]string{masterSSH, "calicoctl", "get", "ippools", poolName.String(), "--output=custom-columns=IPIPMODE"})
			ipmodeCmdOut, _ := tools.NewDoCmd("sshpass", ipmodeCmd...)
			fmt.Println("check IPIPMODE: ", ipmodeCmdOut.String())

			vxmodeCmd := tools.RemoteSSHCmdArray([]string{masterSSH, "calicoctl", "get", "ippools", poolName.String(), "--output=custom-columns=VXLANMODE"})
			vxmodeCmdOut, _ := tools.NewDoCmd("sshpass", vxmodeCmd...)
			fmt.Println("check VXLANMODE: ", vxmodeCmdOut.String())

			if strings.ToUpper(tunal_mode) == "IPIP_ALWAYS" {
				gomega.Expect(ipmodeCmdOut.String()).Should(gomega.ContainSubstring("Always"))
				gomega.Expect(vxmodeCmdOut.String()).Should(gomega.ContainSubstring("Never"))
			}
			if strings.ToUpper(tunal_mode) == "IPIP_CROSSSUBNET" {
				gomega.Expect(ipmodeCmdOut.String()).Should(gomega.ContainSubstring("CrossSubnet"))
				gomega.Expect(vxmodeCmdOut.String()).Should(gomega.ContainSubstring("Never"))
			}
			if strings.ToUpper(tunal_mode) == "VXLAN_ALWAYS" {
				gomega.Expect(ipmodeCmdOut.String()).Should(gomega.ContainSubstring("Never"))
				gomega.Expect(vxmodeCmdOut.String()).Should(gomega.ContainSubstring("Always"))
			}
			if strings.ToUpper(tunal_mode) == "Vxlan_CROSSSUBNET" {
				gomega.Expect(ipmodeCmdOut.String()).Should(gomega.ContainSubstring("Never"))
				gomega.Expect(vxmodeCmdOut.String()).Should(gomega.ContainSubstring("CrossSubnet"))
			}
		})

		ginkgo.It("check pod on different node connection", func() {
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")
			tools.CreatePod(pod1Name, tools.DefaultNamespace, "node1", nginxImage, localKubeConfigPath)
			tools.CreatePod(pod2Name, tools.KubeSystemNamespace, "node2", nginxImage, localKubeConfigPath)

			pod1 := tools.WaitPodBeRunning(cluster1Client, tools.DefaultNamespace, pod1Name, 1000)
			pod2 := tools.WaitPodBeRunning(cluster1Client, tools.KubeSystemNamespace, pod2Name, 1000)
			tools.NodePingPodByPasswd(password, masterSSH, pod2.Status.PodIP)
			tools.NodePingPodByPasswd(password, workerSSH, pod1.Status.PodIP)
			tools.PodPingPodByPasswd(password, masterSSH, tools.DefaultNamespace, pod1Name, pod2.Status.PodIP)
			tools.PodPingPodByPasswd(password, masterSSH, tools.KubeSystemNamespace, pod2Name, pod1.Status.PodIP)
		})
	})
})
