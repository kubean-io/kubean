package tools

import (
	"bytes"
	"context"
	"fmt"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	kubeanClusterClientSet "kubean.io/api/generated/cluster/clientset/versioned"
	"os"
	"os/exec"
	"strings"
	"time"
)

func WaitKubeanJobPodToSuccess(kubeClient *kubernetes.Clientset, podNamespace, podName, expectedStatus string) {
	klog.Info("---- Waiting kubean job-related pod ", podName, " success ----")
	klog.Info("podName: ", podName)
	gomega.Eventually(func() bool {
		pod, err := kubeClient.CoreV1().Pods(podNamespace).Get(context.Background(), podName, metav1.GetOptions{})
		if err != nil {
			klog.Info("Get kubean job-related pod error: ", err.Error())
			return false
		}
		podStatus := string(pod.Status.Phase)
		klog.Info("... podStatus is: ", podStatus)
		if podStatus == PodStatusSucceeded {
			return true
		} else {
			if podStatus == PodStatusFailed {
				cmd := exec.Command("kubectl", "--kubeconfig="+Kubeconfig, "logs", podName, "-n", "kubean-system")
				out, _ := DoCmd(*cmd)
				klog.Info("Get pod log When pod is Error:***")
				klog.Info("pod log string length: ", len(out.String()))
				if len(out.String()) > 10000 {
					klog.Info(out.String()[(len(out.String()) - 10000):len(out.String())])
				} else {
					klog.Info(out.String())
				}
				gomega.Expect(podStatus != PodStatusFailed).To(gomega.BeTrue())
			} else {
				return false
			}
		}
		return false
	}, 300*time.Minute, 1*time.Minute).Should(gomega.BeTrue())
}

func SaveKubeConf(kindConfig *restclient.Config, clusterName, configToSavePath string) {
	clusterClientSet, err := kubeanClusterClientSet.NewForConfig(kindConfig)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")

	cluster1, err := clusterClientSet.KubeanV1alpha1().Clusters().Get(context.Background(), clusterName, metav1.GetOptions{})
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed to get KuBeanCluster")
	klog.Info("****get cluster success")
	fmt.Println("Name:", cluster1.Spec.KubeConfRef.Name, "NameSpace:", cluster1.Spec.KubeConfRef.NameSpace)

	// get kubeconfig from configmap and save to local path
	kubeClient, err := kubernetes.NewForConfig(kindConfig)
	cluster1CF, err := kubeClient.CoreV1().ConfigMaps(cluster1.Spec.KubeConfRef.NameSpace).Get(context.Background(), cluster1.Spec.KubeConfRef.Name, metav1.GetOptions{})
	err1 := os.WriteFile(configToSavePath, []byte(cluster1CF.Data["config"]), 0666)
	gomega.ExpectWithOffset(2, err1).NotTo(gomega.HaveOccurred(), "failed to write localKubeConfigPath")
}

func WaitPodSInKubeSystemBeRunning(kubeClient *kubernetes.Clientset, timeTotalSecond time.Duration, ops ...time.Duration) {
	klog.Info("---- Waiting Pods in %s to be Running ---", KubeSystemNamespace)
	var timeInterval time.Duration = 60
	if len(ops) != 0 {
		timeInterval = ops[0]
	}

	podList, err := kubeClient.CoreV1().Pods(KubeSystemNamespace).List(context.TODO(), metav1.ListOptions{})
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed to get kube-system pods")
	for _, podItem := range podList.Items {
		klog.Info("Waiting ", podItem.Name, "to be Running...")
		gomega.Eventually(func() bool {
			pod, err1 := kubeClient.CoreV1().Pods(KubeSystemNamespace).Get(context.Background(), podItem.Name, metav1.GetOptions{})
			gomega.ExpectWithOffset(2, err1).NotTo(gomega.HaveOccurred(), "Failed get pod", pod.Name)
			podStatus := string(pod.Status.Phase)
			if podStatus == PodStatusRunning {
				return true
			} else {
				klog.Info("    Status is: ", podStatus)
				gomega.Expect(podStatus != PodStatusFailed).To(gomega.BeTrue())
			}
			return false
		}, timeTotalSecond*time.Second, timeInterval*time.Second).Should(gomega.BeTrue())
	}
}

func GenerateClusterClient(localKubeConfigPath string) *kubernetes.Clientset {
	cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
	cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")
	return cluster1Client
}
func WaitPodBeRunning(kubeClient *kubernetes.Clientset, namespace, podName string, timeTotalSecond time.Duration, ops ...time.Duration) *v1.Pod {
	klog.Info("---- Waiting Pods in [%s] to be Running ---", namespace)
	var timeInterval time.Duration = 10
	var pod1 *v1.Pod
	if len(ops) != 0 {
		timeInterval = ops[0]
	}
	gomega.Eventually(func() bool {
		pod1, _ = kubeClient.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
		podStatus := string(pod1.Status.Phase)
		if podStatus == PodStatusRunning {
			return true
		}
		return false
	}, timeTotalSecond*time.Second, timeInterval*time.Second).Should(gomega.BeTrue())
	return pod1
}

func NodePingPodByPasswd(password, sshNode, podIP string) {
	pingCmd := "ping"
	if strings.Contains(podIP, ":") {
		osTypeCmd := RemoteSSHCmdArrayByPasswd(password, []string{sshNode, "cat ", "/etc/redhat-release"})
		osTypeCmdOut, _ := NewDoCmd("sshpass", osTypeCmd...)
		if strings.Contains(osTypeCmdOut.String(), "CentOS") {
			pingCmd = "ping6"
		}
		if strings.Contains(osTypeCmdOut.String(), "Red Hat") && strings.Contains(osTypeCmdOut.String(), "7.") {
			pingCmd = "ping6"
		}
	}

	pingPodIpCmd1 := RemoteSSHCmdArrayByPasswd(password, []string{sshNode, pingCmd, "-c 1", podIP})
	count := 3
	for i := 0; i <= count; i++ {
		pingNginx1IpCmd1Out, cmdError := NewDoCmdSoft("sshpass", pingPodIpCmd1...)
		if cmdError == nil {
			klog.Info(sshNode, "ping pod IP ", podIP, "result: ", pingNginx1IpCmd1Out.String())
			gomega.Expect(pingNginx1IpCmd1Out.String()).Should(gomega.ContainSubstring("1 received"))
			break
		} else {
			klog.Info("excute cmd error, retry...")
			time.Sleep(5 * time.Second)
			gomega.Expect(i == count).Should(gomega.BeFalse())
		}
	}
}

func PodPingPodByPasswd(password, node, podFromNs, podFromName, podToIP string) {
	podsPingCmd1 := RemoteSSHCmdArrayByPasswd(password, []string{node, "kubectl", "exec", "-it", podFromName, "-n", podFromNs, "--", "ping", "-c 1", podToIP})
	count := 3
	for i := 0; i <= count; i++ {
		podsPingCmdOut1, cmdError := NewDoCmdSoft("sshpass", podsPingCmd1...)
		if cmdError == nil {
			fmt.Println("pod ping pod: ", podsPingCmdOut1.String())
			gomega.Expect(podsPingCmdOut1.String()).Should(gomega.ContainSubstring("1 packets received"))
			break
		} else {
			klog.Info("excute cmd error, retry...")
			time.Sleep(5 * time.Second)
			gomega.Expect(i == count).Should(gomega.BeFalse())
		}
	}
}

func SvcCurl(ip string, port int32, checkString string, timeTotalSecond time.Duration, ops ...time.Duration) {
	var timeInterval time.Duration = 5
	var flag bool = false
	if len(ops) != 0 {
		timeInterval = ops[0]
	}
	nginxReq := fmt.Sprintf("%s:%d", ip, port)
	cmd := exec.Command("curl", nginxReq)
	klog.Info("cmd exec: ", cmd)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	gomega.Eventually(func() bool {
		err := cmd.Run()
		if err != nil {
			klog.Info("curl error:", err.Error())
			return false
		}
		flag = strings.Contains(out.String(), checkString)
		return flag
	}, timeTotalSecond*time.Second, timeInterval*time.Second).Should(gomega.BeTrue())
	gomega.Expect(flag).Should(gomega.BeTrue())
}

func DoSonoBuoyCheckByPasswd(password, masterSSH string) {
	subCmd := []string{masterSSH, "sonobuoy", "run", "--sonobuoy-image", "docker.m.daocloud.io/sonobuoy/sonobuoy:v0.56.7", "--plugin-env", "e2e.E2E_FOCUS=pods",
		"--plugin-env", "e2e.E2E_DRYRUN=true", "--wait"}
	klog.Info("sonobuoy check cmd: ", subCmd)
	cmd := RemoteSSHCmdArrayByPasswd(password, subCmd)
	out, _ := NewDoCmd("sshpass", cmd...)
	fmt.Println(out.String())

	sshcmd := RemoteSSHCmdArrayByPasswd(password, []string{masterSSH, "sonobuoy", "status"})
	sshout, _ := NewDoCmd("sshpass", sshcmd...)
	fmt.Println(sshout.String())
	klog.Info("sonobuoy status result:\n", out.String())
	ginkgo.GinkgoWriter.Printf("sonobuoy status result: %s\n", out.String())
	gomega.Expect(sshout.String()).Should(gomega.ContainSubstring("complete"))
	gomega.Expect(sshout.String()).Should(gomega.ContainSubstring("passed"))
}

func DoSonoBuoyCheck(masterSSH string) {
	subCmd := []string{masterSSH, "sonobuoy", "run", "--sonobuoy-image", "docker.m.daocloud.io/sonobuoy/sonobuoy:v0.56.7", "--plugin-env", "e2e.E2E_FOCUS=pods",
		"--plugin-env", "e2e.E2E_DRYRUN=true", "--wait"}
	klog.Info("sonobuoy check cmd: ", subCmd)
	cmd := RemoteSSHCmdArray(subCmd)
	out, _ := NewDoCmd("sshpass", cmd...)
	fmt.Println(out.String())

	sshcmd := RemoteSSHCmdArray([]string{masterSSH, "sonobuoy", "status"})
	sshout, _ := NewDoCmd("sshpass", sshcmd...)
	fmt.Println(sshout.String())
	klog.Info("sonobuoy status result:\n", out.String())
	ginkgo.GinkgoWriter.Printf("sonobuoy status result: %s\n", out.String())
	gomega.Expect(sshout.String()).Should(gomega.ContainSubstring("complete"))
	gomega.Expect(sshout.String()).Should(gomega.ContainSubstring("passed"))
}

func CreatePod(podName, namespace, nodeName, image, kubeconfigFile string) {
	overrideStr := fmt.Sprintf(`{"spec":{"nodeName":"%s"}}`, nodeName)
	klog.Info("...overrideStr is: ", overrideStr)
	//overrideStr := '{"spec":{"nodeName":"node1"}}'
	createCmd := exec.Command("kubectl", "run", podName, "-n", namespace, "--image", image, "--overrides", overrideStr, "--kubeconfig", kubeconfigFile)
	createCmdOut, err1 := DoErrCmd(*createCmd)
	fmt.Println("create nginx1: ", createCmdOut.String(), err1.String())
}

func OperateClusterByYaml(clusterInstallYamlsPath, operatorName string, kindConfig *restclient.Config) {
	installYamlPath := fmt.Sprint(GetKuBeanPath(), clusterInstallYamlsPath)
	kindClient, err := kubernetes.NewForConfig(kindConfig)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")
	cmd := exec.Command("kubectl", "--kubeconfig="+Kubeconfig, "apply", "-f", installYamlPath)
	out, _ := DoCmd(*cmd)
	klog.Info("create cluster result:", out.String())
	time.Sleep(10 * time.Second)

	// Check if the job and related pods have been created
	pods := &v1.PodList{}
	klog.Info("Wait job related pod to be created")
	labelStr := fmt.Sprintf("job-name=kubean-%s-job", operatorName)
	klog.Info("label is: ", labelStr)
	gomega.Eventually(func() bool {
		pods, _ = kindClient.CoreV1().Pods(KubeanNamespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: labelStr,
		})
		if len(pods.Items) > 0 {
			return true
		}
		return false
	}, 120*time.Second, 5*time.Second).Should(gomega.BeTrue())

	jobPodName := pods.Items[0].Name
	WaitKubeanJobPodToSuccess(kindClient, KubeanNamespace, jobPodName, PodStatusSucceeded)

}
