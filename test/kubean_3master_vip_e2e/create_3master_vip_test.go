package kubean_3master_vip_e2e

import (
	"fmt"
	"strings"
	"time"

	"github.com/kubean-io/kubean/test/tools"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func GetVipNode(password string, nodes []string, vipAdd string) (string, error) {
	allNodesNotMatched := true
	for _, node := range nodes {
		var rootNode string
		newMasterCmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{node, "ip a"})
		out, _ := tools.NewDoCmd("sshpass", newMasterCmd...)
		if strings.Contains(out.String(), vipAdd) {
			fmt.Println("vip address", vipAdd)
			rootNode = node
			vipNode := strings.Replace(rootNode, "root@", "", -1)
			klog.Info("the node where vip is located is ", vipNode)
			allNodesNotMatched = false
			return rootNode, nil
		} else {
			klog.Info("have not matched node ", node)
		}
	}
	gomega.Expect(allNodesNotMatched == false).Should(gomega.BeTrue())
	return "", fmt.Errorf("no node matched with vip address %s", vipAdd)
}

var _ = ginkgo.Describe("Create 3master cluster", func() {

	ginkgo.Context("Parameters init", func() {
		var masterSSH = fmt.Sprintf("root@%s", tools.Vmipaddr)
		var master2SSH = fmt.Sprintf("root@%s", tools.Vmipaddr2)
		var master3SSH = fmt.Sprintf("root@%s", tools.Vmipaddr3)
		localKubeConfigPath := "cluster1.config"
		var offlineConfigs tools.OfflineConfig
		var password = tools.VmPassword
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

		// create cluster with 3 master nodes and kube-vip component
		ginkgo.It("Create 3master cluster and all kube-system pods be running", func() {
			clusterInstallYamlsPath := tools.E2eInstallClusterYamlFolder
			kubeanClusterOpsName := tools.ClusterOperationName
			kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
			tools.OperateClusterByYaml(clusterInstallYamlsPath, kubeanClusterOpsName, kindConfig)

			tools.SaveKubeConf(kindConfig, testClusterName, localKubeConfigPath)
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")
			tools.WaitPodSInKubeSystemBeRunning(cluster1Client, 1800)
			// do sonobuoy check
			if strings.ToUpper(offlineFlag) != "TRUE" {
				klog.Info("On line, sonobuoy check")
				tools.DoSonoBuoyCheckByPasswd(password, masterSSH)
			}
		})

		// after reboot the node where vip is located, find the node that matches vip
		ginkgo.It("reboot node and check which node with vip address", func() {

			// check which node has vip address before vip drifts
			vipAdd := "10.6.178.220"
			nodes := []string{masterSSH, master2SSH, master3SSH}
			var rootNode string
			rootNode, _ = GetVipNode(password, nodes, vipAdd)
			klog.Info("before drift ,the node where vip is located is ", rootNode)

			// get nodes excluding the vip node
			var matchedNodes []string
			for _, node := range nodes {
				if node != rootNode {
					matchedNodes = append(matchedNodes, node)
				}
			}
			fmt.Println("matchedNodes is", matchedNodes)

			// reboot vip node
			newMasterCmd := tools.RemoteSSHCmdArrayByPasswd(password, []string{rootNode, "nohup", "reboot", ">", "/dev/null", "2>&1", "&"})
			out, _ := tools.NewDoCmd("sshpass", newMasterCmd...)
			if out.Len() == 0 {
				klog.Info("reboot ok")
			} else {
				klog.Info(out.String())
			}
			time.Sleep(30 * time.Second)

			// check which node has vip after the node is rebooted
			rootNodeAfterReboot, _ := GetVipNode(password, matchedNodes, vipAdd)
			klog.Info("after drift ,the node where vip is located is ", rootNodeAfterReboot)
		})
	})
})
