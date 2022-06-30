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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = ginkgo.Describe("e2e test cluster operation", func() {

	config, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
	kubeClient, err := kubernetes.NewForConfig(config)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")

	// ginkgo.Context("when reset a cluster", func() {
	// 	clusterResetYamlsPath := "artifacts/example/e2e-rhel86-cluster/reset"
	// 	kubeanNamespace := "kubean-system"
	// 	kubeanClusterOpsName := "e2e-cluster1-reset"

	// 	// Create yaml for kuBean CR and related configuration
	// 	resetYamlPath := fmt.Sprint(tools.GetKuBeanPath(), clusterResetYamlsPath)
	// 	cmd := exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", resetYamlPath)
	// 	ginkgo.GinkgoWriter.Printf("cmd: %s\n", cmd.String())
	// 	var out, stderr bytes.Buffer
	// 	cmd.Stdout = &out
	// 	cmd.Stderr = &stderr
	// 	if err := cmd.Run(); err != nil {
	// 		ginkgo.GinkgoWriter.Printf("reset cmd error: %s\n", err.Error())
	// 		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), stderr.String())
	// 	}

	// 	// Check if the job and related pods have been created
	// 	time.Sleep(30 * time.Second)
	// 	pods, err := kubeClient.CoreV1().Pods(kubeanNamespace).List(context.Background(), metav1.ListOptions{
	// 		LabelSelector: fmt.Sprintf("job-name=kubean-%s-job", kubeanClusterOpsName),
	// 	})
	// 	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed get pod list")
	// 	gomega.Expect(len(pods.Items)).NotTo(gomega.Equal(0))
	// 	jobPodName := pods.Items[0].Name

	// 	// Wait for job-related pod status to be succeeded
	// 	for {
	// 		pod, err := kubeClient.CoreV1().Pods(kubeanNamespace).Get(context.Background(), jobPodName, metav1.GetOptions{})
	// 		ginkgo.GinkgoWriter.Printf("* wait for reset job related pod[%s] status: %s\n", pod.Name, pod.Status.Phase)
	// 		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed get job related pod")
	// 		podStatus := string(pod.Status.Phase)
	// 		if podStatus == "Succeeded" || podStatus == "Failed" {
	// 			ginkgo.It("cluster podStatus should be Succeeded", func() {
	// 				gomega.Expect(podStatus).To(gomega.Equal("Succeeded"))
	// 			})
	// 			break
	// 		}
	// 		time.Sleep(1 * time.Minute)
	// 	}
	// })

	ginkgo.Context("when install a cluster", func() {
		clusterInstallYamlsPath := "artifacts/example/e2e-rhel86-cluster/install"
		kubeanNamespace := "kubean-system"
		kubeanClusterOpsName := "e2e-cluster1-install"
		remoteAddr := "10.6.127.21"
		remoteUser := "root"
		remotePass := "root"
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

		k8ClusterConfig, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build clusterKubeClient config")
		k8ClusterKubeClient, err := kubernetes.NewForConfig(k8ClusterConfig)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new clusterKubeClient set")
		// Check if pods under kube-system namespace are running
		kubepods, err := k8ClusterKubeClient.CoreV1().Pods("kube-system").List(context.Background(), metav1.ListOptions{})
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed get pods in ns:kube-system")
		for _, kubepod := range kubepods.Items {
			ginkgo.GinkgoWriter.Printf("pod[%s] status: %s\n", kubepod.Name, kubepod.Status.Phase)
			ginkgo.It("cluster kubepod Status should be Running", func() {
				gomega.Expect(string(kubepod.Status.Phase)).To(gomega.Equal("Running"))
			})
		}

		// Check if cluster nodes are ready
		nodes, err := k8ClusterKubeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed get nodes")
		for _, node := range nodes.Items {
			ginkgo.GinkgoWriter.Printf("node[%s] status: %s\n", node.Name, node.Status.Conditions[len(node.Status.Conditions)-1].Type)
			ginkgo.It("cluster kubenode Status should be Ready", func() {
				gomega.Expect(string(node.Status.Conditions[len(node.Status.Conditions)-1].Type)).To(gomega.Equal("Ready"))
			})
		}
	})

	ginkgo.Context("when install nginx service", func() {
		nginxReq := "10.6.127.21:31090"
		localKubeConfigPath := "/root/.kube/cluster1-config"
		nignxNSYml := "test/e2e/nginx-namespace.yaml"
		nignxDSYml := "test/e2e/nginx-deployment.yaml"
		nignxSVCYml := "test/e2e/nginx-service.yaml"

		//deploy nginx demo namespace, deployment and service
		cmd := exec.Command("kubectl", "--kubeconfig="+localKubeConfigPath, "apply", "-f", fmt.Sprint(tools.GetKuBeanPath(), nignxNSYml))
		ginkgo.GinkgoWriter.Printf("cmd: %s\n", cmd.String())
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("nignxNSYml exec Error:", err)
		}
		fmt.Println("nignxNSYml exec: ", string(output))

		cmd = exec.Command("kubectl", "--kubeconfig="+localKubeConfigPath, "apply", "-f", fmt.Sprint(tools.GetKuBeanPath(), nignxDSYml))
		ginkgo.GinkgoWriter.Printf("cmd: %s\n", cmd.String())
		output, err = cmd.Output()
		if err != nil {
			fmt.Println("nignxDSYml exec Error:", err)
		}
		fmt.Println("nignxDSYml exec: ", string(output))

		cmd = exec.Command("kubectl", "--kubeconfig="+localKubeConfigPath, "apply", "-f", fmt.Sprint(tools.GetKuBeanPath(), nignxSVCYml))
		ginkgo.GinkgoWriter.Printf("cmd: %s\n", cmd.String())
		output, err = cmd.Output()
		if err != nil {
			fmt.Println("nignxSVCYml exec Error:", err)
		}
		fmt.Println("nignxSVCYml exec: ", string(output))

		time.Sleep(1 * time.Minute)

		// check nginx svc config
		config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build clusterKubeClient config")
		clusterKubeClient, err := kubernetes.NewForConfig(config)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new clusterKubeClient set")
		nginxSVC, err := clusterKubeClient.CoreV1().Services("nginx").Get(context.TODO(), "nginx-deployment1", metav1.GetOptions{})
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed get nginxSVC")
		ginkgo.It("nginx service type should be NodePort", func() {
			gomega.Expect(string(nginxSVC.Spec.Type)).To(gomega.Equal("NodePort"))
		})

		// check nginx request
		cmd = exec.Command("curl", nginxReq)
		ginkgo.GinkgoWriter.Printf("cmd: %s\n", cmd.String())
		output, err = cmd.Output()
		if err != nil {
			fmt.Println("curl nginx exec Error:", err)
		}
		fmt.Println("curl nginx exec: ", string(output))
		ginkgo.It("nginx service can be request", func() {
			gomega.Expect(string(output)).Should(gomega.ContainSubstring("Welcome to nginx!"))
		})
	})

})
