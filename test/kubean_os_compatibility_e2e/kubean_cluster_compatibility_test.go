package kubean_os_compatibility_e2e

import (
	"context"
	"fmt"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os/exec"
	"time"

	"github.com/kubean-io/kubean/test/tools"
)

var _ = ginkgo.Describe("e2e test compatibility redhat84 1 master + 1 worker", func() {

	kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "clientcmd.BuildConfigFromFlags error")
	kindClient, err := kubernetes.NewForConfig(kindConfig)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "kubernetes.NewForConfig error")
	localKubeConfigPath := "cluster1-config"

	var masterSSH = fmt.Sprintf("root@%s", tools.Vmipaddr)
	var workerSSH = fmt.Sprintf("root@%s", tools.Vmipaddr2)

	// do cluster installation within docker
	ginkgo.Context("when install a redhat84 cluster using docker", func() {

		clusterInstallYamlsPath := "e2e-install-cluster"
		kubeanClusterOpsName := "e2e-cluster1-install"
		testClusterName := "cluster1"
		ginkgo.It("Start create RedHat85 K8S cluster", func() {

			//Create cluster by apply yaml
			installYamlPath := fmt.Sprint(tools.GetKuBeanPath(), clusterInstallYamlsPath)
			cmd := exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", installYamlPath)
			out, _ := tools.DoCmd(*cmd)
			klog.Info("create cluster result:", out.String())
			time.Sleep(10 * time.Second)

			// wait kubean create-cluster pod to success.
			pods, _ := kindClient.CoreV1().Pods(tools.KubeanNamespace).List(context.Background(), metav1.ListOptions{
				LabelSelector: fmt.Sprintf("job-name=kubean-%s-job", kubeanClusterOpsName),
			})
			gomega.Expect(len(pods.Items)).NotTo(gomega.Equal(0))
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
			tools.DoSonoBuoyCheck(masterSSH)

			// do network check:
			// => 1. create nginx1 pod on node1, create nginx2 pod on node2
			pod1Name := "nginx1"
			pod2Name := "nginx2"
			nginx1Cmd := exec.Command("kubectl", "run", pod1Name, "-n", tools.KubeSystemNamespace, "--image", "nginx:alpine", "--kubeconfig", localKubeConfigPath, "--env", "NodeName=node1")
			nginx1CmdOut, err1 := tools.DoErrCmd(*nginx1Cmd)
			klog.Info("create [%s] :", nginx1CmdOut.String(), err1.String())
			nginx2Cmd := exec.Command("kubectl", "run", pod2Name, "-n", tools.DefaultNamespace, "--image", "nginx:alpine", "--kubeconfig", localKubeConfigPath, "--env", "NodeName=node2")
			nginx2CmdOut, err2 := tools.DoErrCmd(*nginx2Cmd)
			klog.Info("create [%s] :", nginx2CmdOut.String(), err2.String())

			// do network check:
			// => 2. wait pod to be Running
			pod1 := tools.WaitPodBeRunning(cluster1Client, tools.KubeSystemNamespace, pod1Name, 1000)
			pod2 := tools.WaitPodBeRunning(cluster1Client, tools.DefaultNamespace, pod2Name, 1000)

			// do network check:
			// ping pod on node
			klog.Info("Node to pod connection check...")
			tools.NodePingPod(masterSSH, pod2.Status.PodIP)
			tools.NodePingPod(workerSSH, pod1.Status.PodIP)

			// do network check:
			//=> 3. pod ping pod
			klog.Info("Pod to pod connection check...")
			tools.PodPingPod(masterSSH, tools.KubeSystemNamespace, pod1Name, pod2.Status.PodIP)
			tools.PodPingPod(masterSSH, tools.DefaultNamespace, pod2Name, pod1.Status.PodIP)

			//service check
			service1Cmd := exec.Command("kubectl", "expose", "pod", pod2Name, "-n", tools.DefaultNamespace, "--port", "18081", "--target-port", "80", "--type", "NodePort", "--name", "nginx2svc", "--kubeconfig", localKubeConfigPath)
			service1CmdOut, err1 := tools.DoErrCmd(*service1Cmd)
			klog.Info("create service result:", service1CmdOut.String(), err1.String())
			svc, err := cluster1Client.CoreV1().Services(tools.DefaultNamespace).Get(context.Background(), "nginx2svc", metav1.GetOptions{})
			port := svc.Spec.Ports[0].NodePort
			time.Sleep(10 * time.Second)
			tools.SvcCurl(tools.Vmipaddr, port, "Welcome to nginx!", 60)
		})
	})

})
