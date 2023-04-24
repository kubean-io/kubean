package kubean_cilium_cluster_e2e

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
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var _ = ginkgo.Describe("create cilium clusters one master and one worker", func() {
	ginkgo.Context("precondition: deploy one node cluster using private key file", func() {
		var masterSSH = fmt.Sprintf("root@%s", tools.Vmipaddr)
		var workerSSH = fmt.Sprintf("root@%s", tools.Vmipaddr2)
		localKubeConfigPath := "cluster1.config"
		var offlineConfigs tools.OfflineConfig
		var pod1Name = "nginx1"
		var pod2Name = "nginx2"
		var pod3Name = "nginx3"
		var svc1Name = "nginxsvc1"
		var password = tools.VmPassword
		//kubeanNamespace := tools.KubeanNamespace
		var newKubeanNamespace = "new-kubean-system"
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
		//check kubean deployment status
		ginkgo.It("check kubean deployment status", func() {
			kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new kindConfig set")
			kindClient, err := kubernetes.NewForConfig(kindConfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new kindClient")

			deploymentList, _ := kindClient.AppsV1().Deployments(newKubeanNamespace).List(context.TODO(), metav1.ListOptions{})
			for _, dm := range deploymentList.Items {
				if dm.Name == "kubean" {
					gomega.Expect(dm.Status.AvailableReplicas).To(gomega.Equal(int32(3)))
				}
			}
		})

		ginkgo.It("Create cilium cluster and all kube-system pods be running", func() {
			clusterInstallYamlsPath := "e2e-install-cilium-cluster"
			kubeanClusterOpsName := tools.ClusterOperationName
			kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
			tools.OperateClusterByYaml(clusterInstallYamlsPath, kubeanClusterOpsName, kindConfig, newKubeanNamespace)

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
		ginkgo.It("Cilium pod check", func() {
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")

			podList, _ := cluster1Client.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
			ciliumPodNumber := 0
			for _, pod := range podList.Items {
				if strings.Contains(pod.ObjectMeta.Name, "cilium") {
					gomega.Expect(string(pod.Status.Phase)).To(gomega.Equal("Running"))
					ciliumPodNumber += 1
				}
			}
			gomega.Expect(ciliumPodNumber).To(gomega.Equal(4))
		})

		ginkgo.It("create pod1, pod2, pod3", func() {
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")
			//create pod1 and pod2 on master
			tools.CreatePod(pod1Name, tools.DefaultNamespace, "node1", nginxImage, localKubeConfigPath)
			tools.CreatePod(pod2Name, tools.DefaultNamespace, "node1", nginxImage, localKubeConfigPath)
			//create pod3 on worker
			tools.CreatePod(pod3Name, tools.DefaultNamespace, "node2", nginxImage, localKubeConfigPath)

			pod1 := tools.WaitPodBeRunning(cluster1Client, tools.DefaultNamespace, pod1Name, 1000)
			pod2 := tools.WaitPodBeRunning(cluster1Client, tools.DefaultNamespace, pod2Name, 1000)
			pod3 := tools.WaitPodBeRunning(cluster1Client, tools.DefaultNamespace, pod3Name, 1000)

			klog.Info("Pod1IPs is: ", pod1.Status.PodIP)
			klog.Info("Pod2IPs is: ", pod2.Status.PodIP)
			klog.Info("Pod3IPs is: ", pod3.Status.PodIP)
			klog.Info(len(pod1.Status.PodIP))
			klog.Info(len(pod2.Status.PodIP))
			klog.Info(len(pod3.Status.PodIP))

			//check ip range, pod1, pod2 and pod3 are all in kube_pods_subnet
			podList, _ := cluster1Client.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
			for _, pod := range podList.Items {
				ipSplitArr := strings.Split(pod.Status.PodIP, ".")
				gomega.Expect(len(ipSplitArr)).Should(gomega.Equal(4))

				ipSub1, err := strconv.Atoi(ipSplitArr[0])
				gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "ip split conversion failed")
				ipSub2, err := strconv.Atoi(ipSplitArr[1])
				gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "ip split conversion failed")
				ipSub3, err := strconv.Atoi(ipSplitArr[2])
				gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "ip split conversion failed")
				//192.168.128.1~192.168.143.254
				gomega.Expect(ipSub1).Should(gomega.Equal(192))
				gomega.Expect(ipSub2).Should(gomega.Equal(88))
				gomega.Expect(ipSub3 >= 128).Should(gomega.BeTrue())
				gomega.Expect(ipSub3 <= 143).Should(gomega.BeTrue())
			}

			tools.NodePingPodByPasswd(password, masterSSH, fmt.Sprint(pod2.Status.PodIP))
			tools.NodePingPodByPasswd(password, workerSSH, fmt.Sprint(pod1.Status.PodIP))
			//PodPingPod may fail because of there is no ping cmd, use curl to check.
			tools.PodPingPodByPasswd(password, masterSSH, tools.DefaultNamespace, pod1Name, pod2.Status.PodIP)
			tools.PodPingPodByPasswd(password, masterSSH, tools.DefaultNamespace, pod1Name, pod3.Status.PodIP)

		})

		//check ip range of nginx service, should be in kube_service_addresses
		ginkgo.It("Support cilium: kube_service_addresses ", func() {

			//This case need the nginx pod created in the upper case
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")

			service1Cmd := exec.Command("kubectl", "expose", "pod", pod2Name, "-n", tools.DefaultNamespace, "--port", "18081", "--target-port", "80", "--type", "NodePort", "--name", svc1Name, "--kubeconfig", localKubeConfigPath)
			service1CmdOut, err1 := tools.DoErrCmd(*service1Cmd)
			klog.Info("create service result:", service1CmdOut.String(), err1.String())
			svc, err := cluster1Client.CoreV1().Services(tools.DefaultNamespace).Get(context.Background(), svc1Name, metav1.GetOptions{})
			port := svc.Spec.Ports[0].NodePort
			time.Sleep(10 * time.Second)
			tools.SvcCurl(tools.Vmipaddr, port, "Welcome to nginx!", 60)

			//the pod set was 10.88.0.0/16, so the available svc ip range is 10.88.0.1 ~ 10.88.255.254
			svc1, err := cluster1Client.CoreV1().Services(tools.DefaultNamespace).Get(context.Background(), svc1Name, metav1.GetOptions{})
			klog.Info("svc ip ia: ", svc1.Spec.ClusterIP)
			ipSplitArr := strings.Split(svc1.Spec.ClusterIP, ".")
			gomega.Expect(len(ipSplitArr)).Should(gomega.Equal(4))

			ipSub1, err := strconv.Atoi(ipSplitArr[0])
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "ip split conversion failed")
			ipSub2, err := strconv.Atoi(ipSplitArr[1])
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "ip split conversion failed")
			//10.88.*.*
			gomega.Expect(ipSub1).Should(gomega.Equal(10))
			gomega.Expect(ipSub2).Should(gomega.Equal(88))
		})

	})
})
