package kubean_ipvs_cluster_e2e

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
	"time"
)

var _ = ginkgo.Describe("Create ipvs cluster", func() {
	ginkgo.Context("Parameters init", func() {
		var masterSSH = fmt.Sprintf("root@%s", tools.Vmipaddr)
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

		ginkgo.It("Create ipvs cluster and all kube-system pods be running", func() {
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
		//check cilium status, should be running
		ginkgo.It("IPVS mode check", func() {
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")

			//cm, _ := cluster1Client.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
			kubeProxyCm, _ := cluster1Client.CoreV1().ConfigMaps(tools.KubeSystemNamespace).Get(context.Background(), "kube-proxy", metav1.GetOptions{})
			gomega.Expect(kubeProxyCm.String()).To(gomega.ContainSubstring("mode: ipvs"))
		})

		ginkgo.It("create service to expose 2 pod on different node", func() {
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")
			daemonSetName := "my-nginx-ds"
			serviceName := "my-nginx-svc"

			myDamonSet := tools.CreateDaemonSet(daemonSetName, tools.DefaultNamespace, nginxImage, cluster1Client)
			// waiting for all pod ready
			gomega.Eventually(func() bool {
				expectedReady := myDamonSet.Status.DesiredNumberScheduled
				actuallyReady := myDamonSet.Status.NumberReady
				if expectedReady == actuallyReady {
					return true
				}
				klog.Info("Waiting for daemonset pods ready...")
				return false
			}, 300*time.Second, 10*time.Second).Should(gomega.BeTrue())

			service1 := tools.ExposeServiceToDaemonset(serviceName, tools.DefaultNamespace, "NodePort", daemonSetName, cluster1Client)
			// wait service ready
			gomega.Eventually(func() bool {
				serviceTmp, _ := cluster1Client.CoreV1().Services(tools.DefaultNamespace).Get(context.Background(), serviceName, metav1.GetOptions{})
				if len(serviceTmp.Spec.Ports) > 0 {
					return true
				}
				return false
			}, 20*time.Second, 5*time.Second).Should(gomega.BeTrue())
			nodePort0 := service1.Spec.Ports[0].NodePort
			clusterPort := service1.Spec.Ports[0].Port
			clusterIP0 := service1.Spec.ClusterIP
			labelSelectorStr := "name=" + daemonSetName
			podList, err1 := cluster1Client.CoreV1().Pods(tools.DefaultNamespace).List(context.Background(), metav1.ListOptions{LabelSelector: labelSelectorStr})
			if err1 != nil {
				klog.Info("Get podlist of service failed: ", err1.Error())
			}
			gomega.Expect(err1).Should(gomega.BeNil())
			fromPodName := podList.Items[0].Name
			klog.Info(nodePort0)
			klog.Info(clusterPort)
			checkString := "Welcome to nginx!"

			//recycling curl the service can scan all node ip of the k8s cluster
			for i := 0; i < 5; i++ {
				// curl the service from runner host
				tools.SvcCurl(tools.Vmipaddr, nodePort0, checkString, 60)

				// curl the service from master node of the k8s cluster
				tools.SvcCurl(clusterIP0, clusterPort, checkString, 60, "node", password, masterSSH)

				//curl the service from a pod of the service pod
				tools.SvcCurl(clusterIP0, clusterPort, checkString, 60, "pod", password, masterSSH, fromPodName, tools.DefaultNamespace)
			}

		})

	})
})
