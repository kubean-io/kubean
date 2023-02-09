package kubean_calico_nightlye2e

import (
	"context"
	"fmt"
	"github.com/kubean-io/kubean/test/tools"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"strings"
)

var _ = ginkgo.Describe("Calico single stack tunnel: IPIP_ALWAYS", func() {

	var masterSSH = fmt.Sprintf("root@%s", tools.Vmipaddr)
	var workerSSH = fmt.Sprintf("root@%s", tools.Vmipaddr2)
	ginkgo.Context("when install a cluster based on calico single stack", func() {
		var offlineConfigs tools.OfflineConfig
		var pod1Name = "nginx1"
		var pod2Name = "nginx2"
		var password = tools.VmPassword
		testClusterName := tools.TestClusterName
		localKubeConfigPath := "calico-single-stack.config"
		nginxImage := "nginx:alpine"
		offlineFlag := tools.IsOffline
		offlineConfigs = tools.InitOfflineConfig()
		if strings.ToUpper(offlineFlag) == "TRUE" && strings.ToUpper(tools.Arch) == "ARM64" {
			nginxImage = offlineConfigs.NginxImageARM64
		}
		if strings.ToUpper(offlineFlag) == "TRUE" && strings.ToUpper(tools.Arch) == "AMD64" {
			nginxImage = offlineConfigs.NginxImageAMD64
		}
		kubeanClusterOpsName := tools.ClusterOperationName
		kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
		klog.Info("nginx image is: ", nginxImage)
		klog.Info("offlineFlag is: ", offlineFlag)
		klog.Info("arch is: ", tools.Arch)
		ginkgo.It("Create cluster and all kube-system pods be running", func() {
			clusterInstallYamlsPath := "e2e-install-calico-cluster"
			tools.OperateClusterByYaml(clusterInstallYamlsPath, kubeanClusterOpsName, kindConfig)
			tools.SaveKubeConf(kindConfig, testClusterName, localKubeConfigPath)
			cluster1Client := tools.GenerateClusterClient(localKubeConfigPath)
			tools.WaitPodSInKubeSystemBeRunning(cluster1Client, 1800)
			// do sonobuoy check
			if strings.ToUpper(offlineFlag) != "TRUE" {
				tools.DoSonoBuoyCheckByPasswd(password, masterSSH, offlineFlag)
			} else {
				tools.DoSonoBuoyCheckByPasswd(password, masterSSH, offlineFlag, offlineConfigs.SonobuoyImage, offlineConfigs.ConformanceImage, offlineConfigs.SystemdLogImage)
			}
		})

		ginkgo.It("Calico pod check", func() {
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
		})

		ginkgo.It("check calico tunnel valid", func() {
			tunal_mode := tools.OtherLabel
			klog.Info("tunal_mode is: ", tunal_mode)
			poolCmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "calicoctl", "get", "ippools", "--output=go-template=\"{{range .}}{{range .Items}}{{.ObjectMeta.Name}}{{end}}{{end}}\""})
			poolName, _ := tools.NewDoCmd("sshpass", poolCmd...)
			fmt.Println("check poolName: ", poolName.String())

			ipmodeCmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "calicoctl", "get", "ippools", poolName.String(), "--output=custom-columns=IPIPMODE"})
			ipmodeCmdOut, _ := tools.NewDoCmd("sshpass", ipmodeCmd...)
			fmt.Println("check IPIPMODE: ", ipmodeCmdOut.String())

			vxmodeCmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "calicoctl", "get", "ippools", poolName.String(), "--output=custom-columns=VXLANMODE"})
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
			cluster1Client := tools.GenerateClusterClient(localKubeConfigPath)
			tools.WaitPodSInKubeSystemBeRunning(cluster1Client, 1800)
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
