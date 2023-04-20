package add_worker_e2e

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

var _ = ginkgo.Describe("e2e add worker node operation", func() {
	ginkgo.Context("precondition: deploy one node cluster using private key file", func() {
		var masterSSH = fmt.Sprintf("root@%s", tools.Vmipaddr)
		var workerSSH = fmt.Sprintf("root@%s", tools.Vmipaddr2)
		localKubeConfigPath := "cluster1.config"
		var offlineConfigs tools.OfflineConfig
		var pod1Name = "nginx1"
		var pod2Name = "nginx2"
		var password = tools.VmPassword
		//kubeanNamespace := tools.KubeanNamespace
		testClusterName := tools.TestClusterName
		nginxImage := tools.NginxAlpha
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
			clusterInstallYamlsPath := "e2e-install-calico-dual-stack-cluster"
			kubeanClusterOpsName := tools.ClusterOperationName
			kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
			tools.OperateClusterByYaml(clusterInstallYamlsPath, kubeanClusterOpsName, kindConfig)
			tools.SaveKubeConf(kindConfig, testClusterName, localKubeConfigPath)
			cluster1Client := tools.GenerateClusterClient(localKubeConfigPath)
			tools.WaitPodSInKubeSystemBeRunning(cluster1Client, 1800)
			// do sonobuoy check
			if strings.ToUpper(offlineFlag) != "TRUE" {
				klog.Info("On line, sonobuoy check")
				tools.DoSonoBuoyCheckByPasswd(password, masterSSH)
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
			tunnelModeConfig := tools.OtherLabel
			klog.Info("tunnelModeConfig: ", tools.OtherLabel)
			klog.Info("tunal_mode is: ", tunnelModeConfig)
			tunnelModeV4Config := strings.Split(tunnelModeConfig, "-")[0]
			tunnelModeV6Config := strings.Split(tunnelModeConfig, "-")[1]
			klog.Info("tunal_mode_v4: ", tunnelModeV4Config)
			klog.Info("tunal_mode_v6: ", tunnelModeV6Config)
			/*poolCmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "calicoctl", "get", "ippools", "--output=go-template=\"{{range .}}{{range .Items}}{{.ObjectMeta.Name}}{{end}}{{end}}\""})
			poolName, _ := tools.NewDoCmd("sshpass", poolCmd...)
			fmt.Println("check poolName: ", poolName.String())*/
			ipType := [...]string{"ipv4", "ipv6"}
			poolName := ""
			tunnelMode := ""
			for _, ip := range ipType {
				if ip == "ipv4" {
					poolName = "default-pool"
					tunnelMode = tunnelModeV4Config
					klog.Info("Check ipv4 tunnelMode...")
				}
				if ip == "ipv6" {
					poolName = "default-pool-ipv6"
					tunnelMode = tunnelModeV6Config
					klog.Info("Check ipv6 tunnelMode...")
				}
				ipmodeCmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "calicoctl", "get", "ippools", poolName, "--output=custom-columns=IPIPMODE"})
				ipmodeCmdOut, _ := tools.NewDoCmd("sshpass", ipmodeCmd...)
				vxmodeCmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "calicoctl", "get", "ippools", poolName, "--output=custom-columns=VXLANMODE"})
				vxmodeCmdOut, _ := tools.NewDoCmd("sshpass", vxmodeCmd...)
				klog.Info("ipmodeCmdOut: ", ipmodeCmdOut.String())
				klog.Info("vxmodeCmdOut: ", vxmodeCmdOut.String())
				if strings.ToUpper(tunnelMode) == "IPIP_ALWAYS" {
					gomega.Expect(ipmodeCmdOut.String()).Should(gomega.ContainSubstring("Always"))
					gomega.Expect(vxmodeCmdOut.String()).Should(gomega.ContainSubstring("Never"))
				}
				if strings.ToUpper(tunnelMode) == "IPIP_CROSSSUBNET" {
					gomega.Expect(ipmodeCmdOut.String()).Should(gomega.ContainSubstring("CrossSubnet"))
					gomega.Expect(vxmodeCmdOut.String()).Should(gomega.ContainSubstring("Never"))
				}
				if strings.ToUpper(tunnelMode) == "VXLAN_ALWAYS" {
					gomega.Expect(ipmodeCmdOut.String()).Should(gomega.ContainSubstring("Never"))
					gomega.Expect(vxmodeCmdOut.String()).Should(gomega.ContainSubstring("Always"))
				}
				if strings.ToUpper(tunnelMode) == "Vxlan_CROSSSUBNET" {
					gomega.Expect(ipmodeCmdOut.String()).Should(gomega.ContainSubstring("Never"))
					gomega.Expect(vxmodeCmdOut.String()).Should(gomega.ContainSubstring("CrossSubnet"))
				}
			}
		})
		ginkgo.It("check nginx pod has ipv4 and ipv6", func() {
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")
			tools.CreatePod(pod1Name, tools.DefaultNamespace, "node1", nginxImage, localKubeConfigPath)
			tools.CreatePod(pod2Name, tools.KubeSystemNamespace, "node2", nginxImage, localKubeConfigPath)

			pod1 := tools.WaitPodBeRunning(cluster1Client, tools.DefaultNamespace, pod1Name, 1000)
			pod2 := tools.WaitPodBeRunning(cluster1Client, tools.KubeSystemNamespace, pod2Name, 1000)
			klog.Info("Pod1IPs is: ", pod1.Status.PodIPs)
			klog.Info("Pod2IPs is: ", pod2.Status.PodIPs)
			klog.Info(len(pod1.Status.PodIPs))
			klog.Info(len(pod2.Status.PodIPs))
			gomega.Expect(len(pod1.Status.PodIPs)).Should(gomega.Equal(2))
			gomega.Expect(len(pod2.Status.PodIPs)).Should(gomega.Equal(2))

			tools.NodePingPodByPasswd(password, masterSSH, fmt.Sprint(pod2.Status.PodIPs[0].IP))
			tools.NodePingPodByPasswd(password, masterSSH, fmt.Sprint(pod2.Status.PodIPs[1].IP))
			tools.NodePingPodByPasswd(password, workerSSH, fmt.Sprint(pod1.Status.PodIPs[0].IP))
			tools.NodePingPodByPasswd(password, workerSSH, fmt.Sprint(pod1.Status.PodIPs[1].IP))
			tools.PodPingPodByPasswd(password, masterSSH, tools.DefaultNamespace, pod1Name, pod2.Status.PodIPs[0].IP)
			tools.PodPingPodByPasswd(password, masterSSH, tools.DefaultNamespace, pod1Name, pod2.Status.PodIPs[1].IP)
			tools.PodPingPodByPasswd(password, masterSSH, tools.KubeSystemNamespace, pod2Name, pod1.Status.PodIPs[0].IP)
			tools.PodPingPodByPasswd(password, masterSSH, tools.KubeSystemNamespace, pod2Name, pod1.Status.PodIPs[1].IP)
		})
	})
})
