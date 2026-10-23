package tools

import (
	"context"
	"fmt"
	"math"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	kubeanClusterClientSet "github.com/kubean-io/kubean-api/generated/cluster/clientset/versioned"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func WaitKubeanJobPodToSuccess(kubeClient *kubernetes.Clientset, podNamespace, podName, expectedStatus string) {
	klog.Info("---- Waiting kubean job-related pod ", podName, " success ----")
	klog.Info("podName: ", podName)
	deadline := time.Now().Add(300 * time.Minute)
	lastStatus := ""
	var lastErr error
	var lastPod *corev1.Pod

	for {
		pod, err := kubeClient.CoreV1().Pods(podNamespace).Get(context.Background(), podName, metav1.GetOptions{})
		if err != nil {
			klog.Info("Get kubean job-related pod error: ", err.Error())
			lastErr = err
		} else {
			lastErr = nil
			lastPod = pod
			lastStatus = string(pod.Status.Phase)
			klog.Info("... podStatus is: ", lastStatus)
			if lastStatus == expectedStatus {
				return
			}
			if lastStatus == PodStatusFailed {
				diagnostics := kubeanJobPodDiagnostics(context.Background(), kubeClient, pod, expectedStatus)
				gomega.ExpectWithOffset(1, lastStatus).To(
					gomega.Equal(expectedStatus),
					diagnostics,
				)
				return
			}
		}

		if time.Now().After(deadline) {
			message := fmt.Sprintf(
				"timed out waiting for kubean job pod %s/%s to reach %s; last status: %q",
				podNamespace,
				podName,
				expectedStatus,
				lastStatus,
			)
			if lastErr != nil {
				message += fmt.Sprintf("; last get error: %v", lastErr)
			}
			if lastPod != nil {
				message += "\n" + kubeanJobPodDiagnostics(context.Background(), kubeClient, lastPod, expectedStatus)
			}
			gomega.ExpectWithOffset(1, lastStatus).To(gomega.Equal(expectedStatus), message)
			return
		}

		time.Sleep(time.Minute)
	}
}

func kubeanJobPodDiagnostics(ctx context.Context, kubeClient *kubernetes.Clientset, pod *corev1.Pod, expectedStatus string) string {
	var diagnostics strings.Builder
	fmt.Fprintf(
		&diagnostics,
		"kubean job pod %s/%s entered %s; expected %s\n",
		pod.Namespace,
		pod.Name,
		pod.Status.Phase,
		expectedStatus,
	)

	appendContainerStatuses(&diagnostics, "init container", pod.Status.InitContainerStatuses)
	appendContainerStatuses(&diagnostics, "container", pod.Status.ContainerStatuses)

	events, err := kubeClient.CoreV1().Events(pod.Namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.uid=%s", pod.UID),
	})
	if err != nil {
		fmt.Fprintf(&diagnostics, "failed to list pod events: %v\n", err)
	} else if len(events.Items) > 0 {
		diagnostics.WriteString("pod events:\n")
		start := 0
		if len(events.Items) > 20 {
			start = len(events.Items) - 20
		}
		for _, event := range events.Items[start:] {
			fmt.Fprintf(
				&diagnostics,
				"- %s %s (x%d): %s\n",
				event.Type,
				event.Reason,
				event.Count,
				event.Message,
			)
		}
	}

	appendContainerLogs(ctx, &diagnostics, kubeClient, pod)
	return strings.TrimSpace(diagnostics.String())
}

func appendContainerStatuses(diagnostics *strings.Builder, label string, statuses []corev1.ContainerStatus) {
	for _, status := range statuses {
		switch {
		case status.State.Terminated != nil:
			terminated := status.State.Terminated
			fmt.Fprintf(
				diagnostics,
				"%s %s terminated: reason=%s exitCode=%d signal=%d message=%q\n",
				label,
				status.Name,
				terminated.Reason,
				terminated.ExitCode,
				terminated.Signal,
				terminated.Message,
			)
		case status.State.Waiting != nil:
			waiting := status.State.Waiting
			fmt.Fprintf(
				diagnostics,
				"%s %s waiting: reason=%s message=%q\n",
				label,
				status.Name,
				waiting.Reason,
				waiting.Message,
			)
		case status.State.Running != nil:
			fmt.Fprintf(diagnostics, "%s %s was still running\n", label, status.Name)
		default:
			fmt.Fprintf(diagnostics, "%s %s has no reported state\n", label, status.Name)
		}
	}
}

func appendContainerLogs(ctx context.Context, diagnostics *strings.Builder, kubeClient *kubernetes.Clientset, pod *corev1.Pod) {
	const (
		logTailLines = int64(300)
		logMaxBytes  = 64 * 1024
	)

	statuses := append(
		append([]corev1.ContainerStatus{}, pod.Status.InitContainerStatuses...),
		pod.Status.ContainerStatuses...,
	)
	for _, status := range statuses {
		tailLines := logTailLines
		options := &corev1.PodLogOptions{
			Container: status.Name,
			TailLines: &tailLines,
		}
		logs, err := kubeClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, options).DoRaw(ctx)
		if err != nil {
			fmt.Fprintf(diagnostics, "failed to read logs for container %s: %v\n", status.Name, err)
			continue
		}
		fmt.Fprintf(
			diagnostics,
			"last %d log lines for container %s:\n%s\n",
			logTailLines,
			status.Name,
			boundedKubeanJobLog(string(logs), logMaxBytes),
		)
	}
}

func normalizeKubeanJobLog(logs string) string {
	logs = strings.ReplaceAll(logs, `\n`, "\n")
	logs = strings.ReplaceAll(logs, `\t`, "\t")
	return strings.TrimSpace(logs)
}

func boundedKubeanJobLog(logs string, maxBytes int) string {
	logs = normalizeKubeanJobLog(logs)
	if len(logs) <= maxBytes {
		return logs
	}

	half := maxBytes / 2
	return fmt.Sprintf(
		"%s\n... %d log bytes omitted ...\n%s",
		logs[:half],
		len(logs)-maxBytes,
		logs[len(logs)-half:],
	)
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

// WaitPodSInKubeSystemBeRunning waits for the current kube-system pods to be Running and Ready.
// The timeout and optional polling interval are specified in seconds.
func WaitPodSInKubeSystemBeRunning(kubeClient *kubernetes.Clientset, timeTotalSecond time.Duration, ops ...time.Duration) {
	klog.Infof("---- Waiting Pods in %s to be Running and Ready ---", KubeSystemNamespace)
	timeInterval := 60 * time.Second
	if len(ops) != 0 {
		timeInterval = ops[0] * time.Second
	}

	err := waitKubeSystemPodsReady(context.Background(), kubeClient, timeTotalSecond*time.Second, timeInterval)
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
}

func waitKubeSystemPodsReady(ctx context.Context, kubeClient kubernetes.Interface, timeout, interval time.Duration) error {
	var lastErr error
	err := wait.PollUntilContextTimeout(ctx, interval, timeout, true, func(ctx context.Context) (bool, error) {
		lastErr = kubeSystemPodsReady(ctx, kubeClient)
		if lastErr != nil {
			klog.Info("Waiting for kube-system pods: ", lastErr)
		}
		return lastErr == nil, nil
	})
	if err != nil {
		return fmt.Errorf("waiting for Running and Ready pods in %s: %w; last observation: %v", KubeSystemNamespace, err, lastErr)
	}
	return nil
}

func kubeSystemPodsReady(ctx context.Context, kubeClient kubernetes.Interface) error {
	podList, err := kubeClient.CoreV1().Pods(KubeSystemNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list pods in %s: %w", KubeSystemNamespace, err)
	}
	if len(podList.Items) == 0 {
		return fmt.Errorf("no pods found in %s", KubeSystemNamespace)
	}

	var diagnostics strings.Builder
	for _, pod := range podList.Items {
		ready := corev1.PodCondition{Status: corev1.ConditionUnknown}
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady {
				ready = condition
				break
			}
		}
		if pod.DeletionTimestamp == nil && pod.Status.Phase == corev1.PodRunning && ready.Status == corev1.ConditionTrue {
			continue
		}

		fmt.Fprintf(&diagnostics, "\n%s/%s (node=%q): phase=%s Ready=%s terminating=%t reason=%q message=%q\n",
			pod.Namespace, pod.Name, pod.Spec.NodeName, pod.Status.Phase, ready.Status,
			pod.DeletionTimestamp != nil, ready.Reason, ready.Message)
		appendContainerStatuses(&diagnostics, "init container", pod.Status.InitContainerStatuses)
		appendContainerStatuses(&diagnostics, "container", pod.Status.ContainerStatuses)
	}
	if diagnostics.Len() != 0 {
		return fmt.Errorf("unhealthy pods in %s:%s", KubeSystemNamespace, diagnostics.String())
	}
	return nil
}

func GenerateClusterClient(localKubeConfigPath string) *kubernetes.Clientset {
	cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
	cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")
	return cluster1Client
}
func WaitPodBeRunning(kubeClient *kubernetes.Clientset, namespace, podName string, timeTotalSecond time.Duration, ops ...time.Duration) *corev1.Pod {
	klog.Infof("---- Waiting Pods in [%s] to be Running ---", namespace)
	var timeInterval time.Duration = 10
	var pod1 *corev1.Pod
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
	err := pingPodFromNode(sshNode, podIP, func(command string) ([]byte, error) {
		args := RemoteSSHCmdArrayByPasswd(password, []string{sshNode, command})
		return exec.Command("sshpass", args...).CombinedOutput()
	})
	gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
}

func pingPodFromNode(sshNode, podIP string, run func(string) ([]byte, error)) error {
	command, err := nodePingCommand(podIP)
	if err != nil {
		return fmt.Errorf("ping pod IP %q from %s: %w", podIP, sshNode, err)
	}

	const maxAttempts = 4
	for attempt := 1; ; attempt++ {
		output, err := run(command)
		if err == nil {
			klog.Infof("%s ping pod IP %s result: %s", sshNode, podIP, output)
			if !strings.Contains(string(output), "1 received") {
				return fmt.Errorf("ping pod IP %s from %s did not report a received packet:\n%s", podIP, sshNode, output)
			}
			return nil
		}

		err = fmt.Errorf("ping pod IP %s from %s failed (attempt %d/%d): %w\n%s", podIP, sshNode, attempt, maxAttempts, err, output)
		klog.Info(err)
		if attempt == maxAttempts {
			return err
		}
		time.Sleep(5 * time.Second)
	}
}

func nodePingCommand(podIP string) (string, error) {
	ip := net.ParseIP(podIP)
	if ip == nil {
		return "", fmt.Errorf("invalid pod IP address %q", podIP)
	}
	if ip.To4() != nil {
		return "ping -c 1 " + ip.String(), nil
	}

	return fmt.Sprintf(`if command -v ping6 >/dev/null 2>&1; then
  echo "using ping6 for IPv6" >&2
  exec ping6 -c 1 %[1]s
elif command -v ping >/dev/null 2>&1; then
  echo "using ping -6 for IPv6" >&2
  exec ping -6 -c 1 %[1]s
else
  echo "IPv6 ping requires ping6 or ping on the target node" >&2
  exit 127
fi`, ip.String()), nil
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

// curlPlace parameter explain:
// 1. local: curl a service on runner node by nodeIP:NodePort
// 2. node: curl a service on k8s cluster node by serviceClusterIP:Port
// 3. pod: curl a service on k8s cluster pod by serviceClusterIP:Port

func SvcCurl(curIP string, curPort int32, checkString string, timeTotalSecond time.Duration, ops ...string) {
	curlPlace := "local"
	var curCmd []string
	loopNum := int(math.Ceil(float64(timeTotalSecond / 5)))
	if len(ops) != 0 {
		curlPlace = ops[0]
		nginxReq := fmt.Sprintf("%s:%d", curIP, curPort)
		if curlPlace == "local" {
			curCmd = RemoteSSHCmdArray([]string{"curl", nginxReq})
		}
		if curlPlace == "node" {
			password := ops[1]
			sshNode := ops[2]
			curCmd = RemoteSSHCmdArrayByPasswd(password, []string{sshNode, "curl", nginxReq})
		}
		if curlPlace == "pod" {
			password := ops[1]
			sshNode := ops[2]
			podName := ops[3]
			podNamespace := ops[4]
			curCmd = RemoteSSHCmdArrayByPasswd(password, []string{sshNode, "kubectl", "exec", "-it", podName, "-n", podNamespace, "--", "curl", nginxReq})

		}
		klog.Info("cur cmd :", curCmd)

		for i := 0; i <= loopNum; i++ {
			curlCmdOut, cmdError := NewDoCmdSoft("sshpass", curCmd...)
			if cmdError == nil {
				gomega.Expect(curlCmdOut.String()).Should(gomega.ContainSubstring(checkString))
				break
			} else {
				klog.Info("Execute cmd failed, retry...")
				time.Sleep(5 * time.Second)
				gomega.Expect(i == loopNum).Should(gomega.BeFalse())
			}
		}
	}
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

func CreatePod(podName, namespace, nodeName, image, kubeconfigFile string) {
	overrideStr := fmt.Sprintf(`{"spec":{"nodeName":"%s"}}`, nodeName)
	klog.Info("...overrideStr is: ", overrideStr)
	//overrideStr := '{"spec":{"nodeName":"node1"}}'
	createCmd := exec.Command("kubectl", "run", podName, "-n", namespace, "--image", image, "--overrides", overrideStr, "--kubeconfig", kubeconfigFile)
	createCmdOut, err1 := DoErrCmd(*createCmd)
	fmt.Println("create nginx1: ", createCmdOut.String(), err1.String())
}

func ExposeServiceToDaemonset(serviceName, serviceNamespace, svcType, sourceSelector string, clientSet *kubernetes.Clientset) *corev1.Service {
	serviceType := corev1.ServiceType(svcType)
	serviceBody := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"name": sourceSelector,
			},
			Type: serviceType,
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Port:     8000,
					Protocol: "TCP",
					TargetPort: intstr.IntOrString{
						IntVal: 80,
					},
					NodePort: 31800,
				},
			},
		},
	}
	service, error1 := clientSet.CoreV1().Services(serviceNamespace).Create(context.Background(), serviceBody, metav1.CreateOptions{})
	if error1 != nil {
		klog.Info("Create service error: ", error1.Error())
	}
	gomega.Expect(error1 == nil).Should(gomega.BeTrue())
	return service
}

func CreateDaemonSet(daemonsetName, namespace, imageName string, client *kubernetes.Clientset) *appsv1.DaemonSet {

	daemonSetsBody := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "DaemonSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: daemonsetName,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": daemonsetName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": daemonsetName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  daemonsetName,
							Image: imageName,
						},
					},
				},
			},
		},
	}

	daemonSet, error1 := client.AppsV1().DaemonSets(namespace).Create(context.Background(), daemonSetsBody, metav1.CreateOptions{})
	if error1 != nil {
		klog.Info("Create daemonset error: ", error1.Error())
	}
	gomega.Expect(error1 == nil).Should(gomega.BeTrue())
	return daemonSet
}

func OperateClusterByYaml(clusterInstallYamlsPath, operatorName string, kindConfig *restclient.Config, args ...string) {
	installYamlPath := fmt.Sprint(GetKuBeanPath(), clusterInstallYamlsPath)
	kindClient, err := kubernetes.NewForConfig(kindConfig)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")
	out := ApplyClusterYamlsInOrder(installYamlPath)
	klog.Info("create cluster result:", out)
	time.Sleep(10 * time.Second)
	kubeanNsName := KubeanNamespace
	if len(args) > 0 {
		kubeanNsName = args[0]
	}
	// Check if the job and related pods have been created
	pods := &corev1.PodList{}
	klog.Info("Wait job related pod to be created")
	labelStr := fmt.Sprintf("job-name=kubean-%s-job", operatorName)
	klog.Info("label is: ", labelStr)
	gomega.Eventually(func() bool {
		pods, _ = kindClient.CoreV1().Pods(kubeanNsName).List(context.Background(), metav1.ListOptions{
			LabelSelector: labelStr,
		})
		if len(pods.Items) > 0 {
			return true
		}
		return false
	}, 120*time.Second, 5*time.Second).Should(gomega.BeTrue())

	jobPodName := pods.Items[0].Name
	WaitKubeanJobPodToSuccess(kindClient, kubeanNsName, jobPodName, PodStatusSucceeded)

}

// ApplyClusterYamlsInOrder applies YAML files from the given directory ensuring all
// non-ClusterOperation resources (Cluster, HostsConfCM, VarsConfCM, etc.) are applied
// before any ClusterOperation. This prevents the kubean operator from copying a stale
// VarsConfCM.
//
// Background: kubectl apply -f <dir> processes files alphabetically, so
// ClusterOperation may be applied before VarsConfCM. The kubean operator reconciles
// ClusterOperation immediately upon creation and copies VarsConfCM at that moment —
// before VarsConfCM has been updated with test-specific configuration. Applying base
// resources first ensures the operator always copies the fully-updated VarsConfCM.
func ApplyClusterYamlsInOrder(installYamlPath string) string {
	entries, err := os.ReadDir(installYamlPath)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(),
		"failed to read yaml directory: "+installYamlPath)

	var baseFiles, clusterOpFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			continue
		}
		fullPath := filepath.Join(installYamlPath, name)
		if isClusterOperationYaml(fullPath) {
			clusterOpFiles = append(clusterOpFiles, fullPath)
		} else {
			baseFiles = append(baseFiles, fullPath)
		}
	}

	var output strings.Builder

	// Step 1: apply Cluster, HostsConfCM, VarsConfCM and any other base resources first.
	if len(baseFiles) > 0 {
		args := []string{"--kubeconfig=" + Kubeconfig, "apply"}
		for _, f := range baseFiles {
			args = append(args, "-f", f)
		}
		cmd := exec.Command("kubectl", args...)
		out, _ := DoCmd(*cmd)
		output.WriteString(out.String())
	}

	// Step 2: apply ClusterOperation last so the operator copies the already-updated VarsConfCM.
	for _, f := range clusterOpFiles {
		cmd := exec.Command("kubectl", "--kubeconfig="+Kubeconfig, "apply", "-f", f)
		out, _ := DoCmd(*cmd)
		output.WriteString(out.String())
	}

	return output.String()
}

// isClusterOperationYaml returns true if the file contains a ClusterOperation resource.
// Detection is based on file content rather than filename because the filename varies
// (e.g. "ClusterOperation.yml", "kubeanClusterOps.yml").
func isClusterOperationYaml(filePath string) bool {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "kind: ClusterOperation")
}
