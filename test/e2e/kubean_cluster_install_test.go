package e2e

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/daocloud/kubean/test/tools"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = ginkgo.Describe("e2e test cluster operation", func() {

	config, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
	kubeClient, err := kubernetes.NewForConfig(config)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")

	ginkgo.Context("when reset a cluster", func() {
		clusterResetYamlsPath := "artifacts/example/e2e-rhel86-cluster/reset"
		kubeanNamespace := "kubean-system"
		kubeanClusterOpsName := "e2e-cluster1-reset"

		// Create yaml for kuBean CR and related configuration
		resetYamlPath := fmt.Sprint(tools.GetKuBeanPath(), clusterResetYamlsPath)
		cmd := exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", resetYamlPath)
		ginkgo.GinkgoWriter.Printf("cmd: %s\n", cmd.String())
		var out, stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			ginkgo.GinkgoWriter.Printf("reset cmd error: %s\n", err.Error())
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), stderr.String())
		}

		// Check if the job and related pods have been created
		time.Sleep(30 * time.Second)
		pods, err := kubeClient.CoreV1().Pods(kubeanNamespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=kubean-%s-job", kubeanClusterOpsName),
		})
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed get pod list")
		gomega.Expect(len(pods.Items)).NotTo(gomega.Equal(0))
		jobPodName := pods.Items[0].Name

		// Wait for job-related pod status to be succeeded
		for {
			pod, err := kubeClient.CoreV1().Pods(kubeanNamespace).Get(context.Background(), jobPodName, metav1.GetOptions{})
			ginkgo.GinkgoWriter.Printf("* wait for reset job related pod[%s] status: %s\n", pod.Name, pod.Status.Phase)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed get job related pod")
			podStatus := string(pod.Status.Phase)
			if podStatus == "Succeeded" || podStatus == "Failed" {
				ginkgo.It("cluster podStatus should be Succeeded", func() {
					gomega.Expect(podStatus).To(gomega.Equal("Succeeded"))
				})
				break
			}
			time.Sleep(1 * time.Minute)
		}
	})

	ginkgo.Context("when install a cluster", func() {
		clusterInstallYamlsPath := "artifacts/example/e2e-rhel86-cluster/install"
		kubeanNamespace := "kubean-system"
		kubeanClusterOpsName := "e2e-cluster1-install"
		remoteAddr := "10.6.127.11"
		remoteUser := "root"
		remotePass := "dangerous"
		remoteKubeConfigPath := "/root/.kube/config"
		localKubeConfigPath := "/root/.kube/cluster1-config"

		// Create yaml for kuBean CR and related configuration
		installYamlPath := fmt.Sprint(tools.GetKuBeanPath(), clusterInstallYamlsPath)
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
		pods, err := kubeClient.CoreV1().Pods(kubeanNamespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=kubean-%s-job", kubeanClusterOpsName),
		})
		gomega.Expect(len(pods.Items)).NotTo(gomega.Equal(0))
		jobPodName := pods.Items[0].Name

		// Wait for job-related pod status to be succeeded
		for {
			pod, err := kubeClient.CoreV1().Pods(kubeanNamespace).Get(context.Background(), jobPodName, metav1.GetOptions{})
			ginkgo.GinkgoWriter.Printf("* wait for install job related pod[%s] status: %s\n", pod.Name, pod.Status.Phase)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed get job related pod")
			podStatus := string(pod.Status.Phase)
			if podStatus == "Succeeded" || podStatus == "Failed" {
				ginkgo.It("cluster podStatus should be Succeeded", func() {
					gomega.Expect(podStatus).To(gomega.Equal("Succeeded"))
				})
				break
			}
			time.Sleep(1 * time.Minute)
		}

		// Copy kubeconfig of the cluster to local
		scpCmd := exec.Command(
			"sshpass",
			"-p", remotePass,
			"scp",
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-q", remoteUser+"@"+remoteAddr+":"+remoteKubeConfigPath,
			localKubeConfigPath)
		ginkgo.GinkgoWriter.Printf("scp cmd: %s\n", scpCmd.String())
		scpCmd.Stdout = &out
		scpCmd.Stderr = &stderr
		if err := scpCmd.Run(); err != nil {
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), stderr.String())
		}
		sedCmd := exec.Command("sed", "-i", "s/127.0.0.1:.*/"+remoteAddr+":6443/", localKubeConfigPath)
		ginkgo.GinkgoWriter.Printf("sed cmd: %s\n", sedCmd.String())
		sedCmd.Stdout = &out
		sedCmd.Stderr = &stderr
		if err := sedCmd.Run(); err != nil {
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), stderr.String())
		}

		// Check if pods under kube-system namespace are running
		k8ClusterConfig, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build clusterKubeClient config")
		k8ClusterKubeClient, err := kubernetes.NewForConfig(k8ClusterConfig)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new clusterKubeClient set")
		kubepods, err := k8ClusterKubeClient.CoreV1().Pods("kube-system").List(context.Background(), metav1.ListOptions{})
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed get pods in ns:kube-system")
		for _, kubepod := range kubepods.Items {
			ginkgo.GinkgoWriter.Printf("pod[%s] status: %s\n", kubepod.Name, kubepod.Status.Phase)
			ginkgo.It("cluster kubepod Status should be Running", func() {
				gomega.Expect(string(kubepod.Status.Phase)).To(gomega.Equal("Running"))
			})
		}

		// Check if cluster nodes are ready
		nodes, err := kubeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed get nodes")
		for _, node := range nodes.Items {
			ginkgo.GinkgoWriter.Printf("node[%s] status: %s\n", node.Name, node.Status.Conditions[len(node.Status.Conditions)-1].Type)
			ginkgo.It("cluster kubenode Status should be Ready", func() {
				gomega.Expect(string(node.Status.Conditions[len(node.Status.Conditions)-1].Type)).To(gomega.Equal("Ready"))
			})
		}
	})

	ginkgo.Context("when install nginx service", func() {
		localKubeConfigPath := "/root/.kube/cluster1-config"

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
								Image:           "docker.m.daocloud.io/nginx:alpine",
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

		time.Sleep(30 * time.Second)
		// check nginx request
		nginxReq := "10.6.127.11:30090"
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
	})
})
