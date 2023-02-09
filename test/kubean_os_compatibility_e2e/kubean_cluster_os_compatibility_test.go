package kubean_os_compatibility_e2e

import (
	"context"
	"fmt"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os/exec"
	"strings"
	"time"

	"github.com/kubean-io/kubean/test/tools"
)

var _ = ginkgo.Describe("e2e test compatibility 1 master + 1 worker", func() {
	var masterSSH = fmt.Sprintf("root@%s", tools.Vmipaddr)
	var workerSSH = fmt.Sprintf("root@%s", tools.Vmipaddr2)
	// do cluster installation within docker
	ginkgo.Context("when install a  cluster using docker", func() {
		var offlineConfigs tools.OfflineConfig
		var pod1Name = "nginx1"
		var pod2Name = "nginx2"
		var svc2Name = "nginxsvc1"
		nginxImage := "nginx:alpine"
		var password = tools.VmPassword
		offlineFlag := tools.IsOffline
		klog.Info("offlineFlag is: ", offlineFlag)
		klog.Info("Arch is: ", tools.Arch)
		offlineConfigs = tools.InitOfflineConfig()
		if strings.ToUpper(offlineFlag) == "TRUE" && strings.ToUpper(tools.Arch) == "ARM64" {
			nginxImage = offlineConfigs.NginxImageARM64
		}
		if strings.ToUpper(offlineFlag) == "TRUE" && strings.ToUpper(tools.Arch) == "AMD64" {
			nginxImage = offlineConfigs.NginxImageAMD64
		}
		klog.Info("nginx images is:", nginxImage)
		localKubeConfigPath := "cluster1-config"
		clusterInstallYamlsPath := "e2e-install-cluster"
		kubeanClusterOpsName := tools.ClusterOperationName
		testClusterName := tools.TestClusterName

		ginkgo.It("Start create K8S cluster", func() {
			kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "clientcmd.BuildConfigFromFlags error")

			tools.OperateClusterByYaml(clusterInstallYamlsPath, kubeanClusterOpsName, kindConfig)

			tools.SaveKubeConf(kindConfig, testClusterName, localKubeConfigPath)
			cluster1Client := tools.GenerateClusterClient(localKubeConfigPath)
			tools.WaitPodSInKubeSystemBeRunning(cluster1Client, 1800)

			// do sonobuoy check
			klog.Info("Sonobuoy check")
			if strings.ToUpper(offlineFlag) != "TRUE" {
				tools.DoSonoBuoyCheckByPasswd(password, masterSSH, offlineFlag)
			} else {
				tools.DoSonoBuoyCheckByPasswd(password, masterSSH, offlineFlag, offlineConfigs.SonobuoyImage, offlineConfigs.ConformanceImage, offlineConfigs.SystemdLogImage)
			}

			// do network check:
			// => 1. create nginx1 pod on node1, create nginx2 pod on node2
			tools.CreatePod(pod1Name, tools.KubeSystemNamespace, "node1", nginxImage, localKubeConfigPath)
			tools.CreatePod(pod2Name, tools.DefaultNamespace, "node2", nginxImage, localKubeConfigPath)

			// do network check:
			// => 2. wait pod to be Running
			pod1 := tools.WaitPodBeRunning(cluster1Client, tools.KubeSystemNamespace, pod1Name, 1000)
			pod2 := tools.WaitPodBeRunning(cluster1Client, tools.DefaultNamespace, pod2Name, 1000)

			// do network check:
			// ping pod on node
			klog.Info("Node to pod connection check...")
			tools.NodePingPodByPasswd(password, masterSSH, pod2.Status.PodIP)
			tools.NodePingPodByPasswd(password, workerSSH, pod1.Status.PodIP)

			// do network check:
			//=> 3. pod ping pod
			klog.Info("Pod to pod connection check...")
			tools.PodPingPodByPasswd(password, masterSSH, tools.KubeSystemNamespace, pod1Name, pod2.Status.PodIP)
			tools.PodPingPodByPasswd(password, masterSSH, tools.DefaultNamespace, pod2Name, pod1.Status.PodIP)

			//service check
			service2Cmd := exec.Command("kubectl", "expose", "pod", pod2Name, "-n", tools.DefaultNamespace, "--port", "18081", "--target-port", "80", "--type", "NodePort", "--name", svc2Name, "--kubeconfig", localKubeConfigPath)
			service2CmdOut, err1 := tools.DoErrCmd(*service2Cmd)
			klog.Info("create service result:", service2CmdOut.String(), err1.String())
			svc, err := cluster1Client.CoreV1().Services(tools.DefaultNamespace).Get(context.Background(), svc2Name, metav1.GetOptions{})
			port := svc.Spec.Ports[0].NodePort
			time.Sleep(10 * time.Second)
			tools.SvcCurl(tools.Vmipaddr, port, "Welcome to nginx!", 60)
		})
	})
})
