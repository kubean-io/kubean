package kubean_os_compatibility_e2e

import (
	"context"
	"fmt"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os/exec"
	"strings"
	"time"

	"github.com/kubean-io/kubean/test/tools"
)

var _ = ginkgo.Describe("e2e test compatibility redhat84 1 master + 1 worker", func() {
	var masterSSH = fmt.Sprintf("root@%s", tools.Vmipaddr)
	var workerSSH = fmt.Sprintf("root@%s", tools.Vmipaddr2)
	// do cluster installation within docker
	ginkgo.Context("when install a redhat84 cluster using docker", func() {
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
		localKubeConfigPath := "cluster1-config"
		clusterInstallYamlsPath := "e2e-install-cluster"
		kubeanClusterOpsName := tools.ClusterOperationName
		testClusterName := tools.TestClusterName

		ginkgo.It("Start create RedHat85 K8S cluster", func() {
			kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "clientcmd.BuildConfigFromFlags error")
			kindClient, err := kubernetes.NewForConfig(kindConfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "kubernetes.NewForConfig error")

			//Create cluster by apply yaml
			installYamlPath := fmt.Sprint(tools.GetKuBeanPath(), clusterInstallYamlsPath)
			cmd := exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", installYamlPath)
			out, _ := tools.DoCmd(*cmd)
			klog.Info("create cluster result:", out.String())
			time.Sleep(10 * time.Second)

			// wait kubean create-cluster pod to success.
			pods := &v1.PodList{}
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

			// Wait all pods in kube-syste to be Running
			tools.WaitPodSInKubeSystemBeRunning(cluster1Client, 1800)

			// do sonobuoy check
			if strings.ToUpper(offlineFlag) != "TRUE" {
				klog.Info("On line, sonobuoy check")
				tools.DoSonoBuoyCheckByPasswd(password, masterSSH)
			}

			// do network check:
			// => 1. create nginx1 pod on node1, create nginx2 pod on node2
			nginx1Cmd := exec.Command("kubectl", "run", pod1Name, "-n", tools.KubeSystemNamespace, "--image", nginxImage, "--kubeconfig", localKubeConfigPath, "--env", "NodeName=node1")
			nginx1CmdOut, err1 := tools.DoErrCmd(*nginx1Cmd)
			klog.Info("create [%s] :", nginx1CmdOut.String(), err1.String())
			nginx2Cmd := exec.Command("kubectl", "run", pod2Name, "-n", tools.DefaultNamespace, "--image", nginxImage, "--kubeconfig", localKubeConfigPath, "--env", "NodeName=node2")
			nginx2CmdOut, err2 := tools.DoErrCmd(*nginx2Cmd)
			klog.Info("create [%s] :", nginx2CmdOut.String(), err2.String())

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
