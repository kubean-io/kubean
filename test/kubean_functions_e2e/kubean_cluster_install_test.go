package kubean_functions_e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/kubean-io/kubean/test/tools"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	kubeanClusterClientSet "kubean.io/api/generated/cluster/clientset/versioned"
)

var _ = ginkgo.Describe("e2e test cluster operation", func() {

	config, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
	kubeClient, err := kubernetes.NewForConfig(config)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")
	localKubeConfigPath := "cluster1-config"
	var masterSSH = fmt.Sprintf("root@%s", tools.Vmipaddr)

	defer ginkgo.GinkgoRecover()

	ginkgo.Context("when install a cluster", func() {
		clusterInstallYamlsPath := "e2e-install-cluster"
		kubeanNamespace := "kubean-system"
		kubeanClusterOpsName := "e2e-cluster1-install"

		// Create yaml for kuBean CR and related configuration
		installYamlPath := fmt.Sprint(tools.GetKuBeanPath(), clusterInstallYamlsPath)
		// do cluster deploy in containerd mode
		cmd := exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", installYamlPath)
		ginkgo.GinkgoWriter.Printf("cmd: %s\n", cmd.String())
		var out, stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			ginkgo.GinkgoWriter.Printf("apply cmd error: %s\n", err.Error())
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), stderr.String())
		}

		// Check if the job and related pods have been created
		time.Sleep(30 * time.Second)
		pods, _ := kubeClient.CoreV1().Pods(kubeanNamespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=kubean-%s-job", kubeanClusterOpsName),
		})
		gomega.Expect(len(pods.Items)).NotTo(gomega.Equal(0))
		jobPodName := pods.Items[0].Name

		// Wait for kubean job-related pod status to be succeeded
		for {
			pod, err := kubeClient.CoreV1().Pods(kubeanNamespace).Get(context.Background(), jobPodName, metav1.GetOptions{})
			ginkgo.GinkgoWriter.Printf("* wait for install job related pod[%s] status: %s\n", pod.Name, pod.Status.Phase)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed get job related pod")
			podStatus := string(pod.Status.Phase)
			if podStatus == "Succeeded" || podStatus == "Failed" {
				ginkgo.It("kubean cluster podStatus should be Succeeded", func() {
					gomega.Expect(podStatus).To(gomega.Equal("Succeeded"))
				})
				break
			}
			time.Sleep(1 * time.Minute)
		}

		clusterClientSet, err := kubeanClusterClientSet.NewForConfig(config)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")

		// from Cluster: cluster1 get kubeconfRef: name: cluster1-kubeconf namespace: kubean-system
		cluster1, err := clusterClientSet.KubeanV1alpha1().Clusters().Get(context.Background(), "cluster1", metav1.GetOptions{})
		fmt.Println("Name:", cluster1.Spec.KubeConfRef.Name, "NameSpace:", cluster1.Spec.KubeConfRef.NameSpace)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed to get Cluster")

		// get configmap
		kubeClient, err := kubernetes.NewForConfig(config)
		cluster1CF, err := kubeClient.CoreV1().ConfigMaps(cluster1.Spec.KubeConfRef.NameSpace).Get(context.Background(), cluster1.Spec.KubeConfRef.Name, metav1.GetOptions{})
		err1 := os.WriteFile(localKubeConfigPath, []byte(cluster1CF.Data["config"]), 0666)
		gomega.ExpectWithOffset(2, err1).NotTo(gomega.HaveOccurred(), "failed to write localKubeConfigPath")

	})

	// check kube-system pod status
	ginkgo.Context("When fetching kube-system pods status", func() {
		config, err = clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
		kubeClient, err = kubernetes.NewForConfig(config)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")

		podList, err := kubeClient.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed to check kube-system pod status")
		ginkgo.It("every pod in kube-system should be in running status", func() {
			for _, pod := range podList.Items {
				fmt.Println(pod.Name, string(pod.Status.Phase))
				gomega.Expect(string(pod.Status.Phase)).To(gomega.Equal("Running"))
			}
		})

		// check hostname after deploy: hostname should be node1
		hostnamecmd := tools.RemoteSSHCmdArray([]string{masterSSH, "hostname"})
		hostnameout, _ := tools.NewDoCmd("sshpass", hostnamecmd...)
		ginkgo.It("set-hostname to node1", func() {
			fmt.Println("hostname: ", hostnameout.String())
			gomega.Expect(hostnameout.String()).Should(gomega.ContainSubstring("node1"))
		})

	})

	// check containerd functions
	ginkgo.Context("Containerd: when check containerd functions", func() {
		masterCmd := tools.RemoteSSHCmdArray([]string{masterSSH, "nerdctl", "info"})
		out, _ := tools.NewDoCmd("sshpass", masterCmd...)
		ginkgo.It("nerdctl info to check if server running: ", func() {
			gomega.Expect(out.String()).Should(gomega.ContainSubstring("k8s.io"))
			gomega.Expect(out.String()).Should(gomega.ContainSubstring("Cgroup Driver: systemd"))
		})

		masterCmd = tools.RemoteSSHCmdArray([]string{masterSSH, "systemctl", "status", "containerd"})
		out1, _ := tools.NewDoCmd("sshpass", masterCmd...)
		ginkgo.It("systemctl status containerd to check if containerd running: ", func() {
			gomega.Expect(out1.String()).Should(gomega.ContainSubstring("/etc/systemd/system/containerd.service;"))
			gomega.Expect(out1.String()).Should(gomega.ContainSubstring("Active: active (running)"))
		})
	})

	ginkgo.Context("Containerd: when install nginx service", func() {
		config, err = clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
		kubeClient, err = kubernetes.NewForConfig(config)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")

		//Create Depployment
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "nginx-deployment",
			},
			Spec: appsv1.DeploymentSpec{
				//Replicas: int32Ptr(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "nginx",
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "nginx",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:            "nginx",
								Image:           "nginx:alpine",
								ImagePullPolicy: "IfNotPresent",
								Ports: []corev1.ContainerPort{
									{
										Name:          "http",
										Protocol:      corev1.ProtocolTCP,
										ContainerPort: 80,
									},
								},
							},
						},
					},
				},
			},
		}
		fmt.Println("Creating nginx deployment...")
		deploymentName := deployment.ObjectMeta.Name
		deploymentClient := kubeClient.AppsV1().Deployments(corev1.NamespaceDefault)
		if _, err = deploymentClient.Get(context.TODO(), deploymentName, metav1.GetOptions{}); err != nil {
			if !apierrors.IsNotFound(err) {
				fmt.Println(err)
				return
			}
			result, err := deploymentClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
			if err != nil {
				panic(err)
			}
			fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
		}

		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: "nginx-svc",
				Labels: map[string]string{
					"app": "nginx",
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": "nginx",
				},
				Type: corev1.ServiceTypeNodePort,
				Ports: []corev1.ServicePort{
					{
						Name:     "http",
						Port:     80,
						Protocol: corev1.ProtocolTCP,
						NodePort: 30090,
					},
				},
			},
		}
		fmt.Println("Creating nginx service...")
		service, err = kubeClient.CoreV1().Services("default").Create(context.TODO(), service, metav1.CreateOptions{})
		fmt.Printf("Created service %q.\n", service.GetObjectMeta().GetName())

		time.Sleep(2 * time.Minute)
		// check nginx request, such as: nginxReq := "10.6.127.41:30090"
		nginxReq := fmt.Sprintf("%s:30090", tools.Vmipaddr)
		cmd := exec.Command("curl", nginxReq)
		ginkgo.GinkgoWriter.Printf("cmd: %s\n", cmd.String())
		var out, stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		fmt.Println("curl nginx exec: ", out.String())
		if err := cmd.Run(); err != nil {
			ginkgo.GinkgoWriter.Printf("curl cmd error: %s\n", err.Error())
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), stderr.String())
		}

		ginkgo.It("nginx service can be request", func() {
			gomega.Expect(out.String()).Should(gomega.ContainSubstring("Welcome to nginx!"))
		})

		ginkgo.It("check pod ip is in kube_pods_subnet", func() {
			//the pod set was 192.168.128.0/20, so the available pod ip range is 192.168.128.1 ~ 192.168.143.255
			pods, err := kubeClient.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed to get pods")
			//gomega.Expect(len(pods.Items) > 0).Should(gomega.BeTrue())
			gomega.Expect(len(pods.Items)).NotTo(gomega.Equal(0))

			podName := pods.Items[0].Name
			pod, err := kubeClient.CoreV1().Pods(corev1.NamespaceDefault).Get(context.Background(), podName, metav1.GetOptions{})
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed to get pod")
			fmt.Println("pod ip is: ", pod.Status.PodIP)
			ipSplitArr := strings.Split(pod.Status.PodIP, ".")
			gomega.Expect(len(ipSplitArr)).Should(gomega.Equal(4))

			ipSub1, err := strconv.Atoi(ipSplitArr[0])
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "ip split conversion failed")
			ipSub2, err := strconv.Atoi(ipSplitArr[1])
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "ip split conversion failed")
			ipSub3, err := strconv.Atoi(ipSplitArr[2])
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "ip split conversion failed")

			gomega.Expect(ipSub1).Should(gomega.Equal(192))
			gomega.Expect(ipSub2).Should(gomega.Equal(168))
			gomega.Expect(ipSub3 >= 128).Should(gomega.BeTrue())
			gomega.Expect(ipSub3 <= 143).Should(gomega.BeTrue())
		})
	})
})
