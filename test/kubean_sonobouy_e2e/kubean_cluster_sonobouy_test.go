package kubean_sonobouy_e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/kubean-io/kubean/test/tools"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	kubeanClusterClientSet "kubean.io/api/generated/kubeancluster/clientset/versioned"
)

var _ = ginkgo.Describe("e2e test cluster 1 master + 1 worker sonobouy check", func() {

	config, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
	kubeClient, err := kubernetes.NewForConfig(config)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")
	localKubeConfigPath := "cluster1-sonobouy-config"

	defer ginkgo.GinkgoRecover()

	// do cluster installation within docker
	ginkgo.Context("when install a sonobouy cluster using docker", func() {
		clusterInstallYamlsPath := "e2e-install-cluster-sonobouy"
		kubeanNamespace := "kubean-system"
		kubeanClusterOpsName := "e2e-cluster1-install-sonobouy"

		// Create yaml for kuBean CR and related configuration
		installYamlPath := fmt.Sprint(tools.GetKuBeanPath(), clusterInstallYamlsPath)
		cmd := exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", installYamlPath)
		out, _ := tools.DoCmd(*cmd)
		fmt.Println(out.String())

		// Check if the job and related pods have been created
		time.Sleep(30 * time.Second)
		pods, _ := kubeClient.CoreV1().Pods(kubeanNamespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=kubean-%s-job", kubeanClusterOpsName),
		})
		gomega.Expect(len(pods.Items)).NotTo(gomega.Equal(0))
		jobPodName := pods.Items[0].Name

		// Wait for job-related pod status to be succeeded
		for {
			pod, err := kubeClient.CoreV1().Pods(kubeanNamespace).Get(context.Background(), jobPodName, metav1.GetOptions{})
			ginkgo.GinkgoWriter.Printf("* wait for install job using docker related pod[%s] status: %s\n", pod.Name, pod.Status.Phase)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed get job related pod")
			podStatus := string(pod.Status.Phase)
			if podStatus == "Succeeded" || podStatus == "Failed" {
				ginkgo.It("cluster deploy job related pod Status should be Succeeded", func() {
					gomega.Expect(podStatus).To(gomega.Equal("Succeeded"))
				})
				break
			}
			time.Sleep(1 * time.Minute)
		}

		clusterClientSet, err := kubeanClusterClientSet.NewForConfig(config)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")

		// from KuBeanCluster: cluster1 get kubeconfRef: name: cluster1-kubeconf namespace: kubean-system
		cluster1, err := clusterClientSet.KubeanclusterV1alpha1().KuBeanClusters().Get(context.Background(), "cluster1", metav1.GetOptions{})
		fmt.Println("Name:", cluster1.Spec.KubeConfRef.Name, "NameSpace:", cluster1.Spec.KubeConfRef.NameSpace)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed to get KuBeanCluster")

		// get configmap
		kubeClient, err := kubernetes.NewForConfig(config)
		cluster1CF, err := kubeClient.CoreV1().ConfigMaps(cluster1.Spec.KubeConfRef.NameSpace).Get(context.Background(), cluster1.Spec.KubeConfRef.Name, metav1.GetOptions{})
		err1 := os.WriteFile(localKubeConfigPath, []byte(cluster1CF.Data["config"]), 0666)
		gomega.ExpectWithOffset(2, err1).NotTo(gomega.HaveOccurred(), "failed to write localKubeConfigPath")

	})

	time.Sleep(2 * time.Minute)
	// check kube-system pod status
	ginkgo.Context("When fetching kube-system pods status", func() {
		config, err = clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
		kubeClient, err = kubernetes.NewForConfig(config)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")
		podList, err := kubeClient.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed to check kube-system pod status")
		for _, pod := range podList.Items {
			for {
				po, _ := kubeClient.CoreV1().Pods("kube-system").Get(context.Background(), pod.Name, metav1.GetOptions{})
				ginkgo.GinkgoWriter.Printf("* wait for kube-system pod[%s] status: %s\n", po.Name, po.Status.Phase)
				podStatus := string(po.Status.Phase)
				if podStatus == "Running" || podStatus == "Failed" {
					ginkgo.It("every pod in kube-system should be in running status", func() {
						gomega.Expect(podStatus).To(gomega.Equal("Running"))
					})
					break
				}
				time.Sleep(1 * time.Minute)
			}
		}

		// check kube version before upgrade
		nodeList, _ := kubeClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		for _, node := range nodeList.Items {
			ginkgo.It("kube node version should be v1.22.12 before cluster upgrade", func() {
				gomega.Expect(node.Status.NodeInfo.KubeletVersion).To(gomega.Equal("v1.22.12"))
				gomega.Expect(node.Status.NodeInfo.KubeProxyVersion).To(gomega.Equal("v1.22.12"))
			})
		}

	})

	// sonobuoy run --sonobuoy-image docker.m.daocloud.io/sonobuoy/sonobuoy:v0.56.7 --plugin-env e2e.E2E_FOCUS=pods --plugin-env e2e.E2E_DRYRUN=true --wait
	ginkgo.Context("do sonobuoy checking ", func() {
		masterSSH := fmt.Sprintf("root@%s", tools.Vmipaddr)
		cmd := exec.Command("sshpass", "-p", "root", "ssh", masterSSH, "sonobuoy", "run", "--sonobuoy-image", "10.6.170.10:5000/sonobuoy/sonobuoy:v0.56.7", "--plugin-env", "e2e.E2E_FOCUS=pods", "--plugin-env", "e2e.E2E_DRYRUN=true", "--wait")
		out, _ := tools.DoCmd(*cmd)
		fmt.Println(out.String())

		sshcmd := exec.Command("sshpass", "-p", "root", "ssh", masterSSH, "sonobuoy", "status")
		sshout, _ := tools.DoCmd(*sshcmd)
		fmt.Println(sshout.String())

		ginkgo.GinkgoWriter.Printf("sonobuoy status result: %s\n", out.String())
		ginkgo.It("sonobuoy status checking result", func() {
			gomega.Expect(sshout.String()).Should(gomega.ContainSubstring("complete"))
			gomega.Expect(sshout.String()).Should(gomega.ContainSubstring("passed"))
		})
	})

	// cluster upgrade
	ginkgo.Context("do cluster upgrade from v1.22.12 to v1.23.7", func() {
		clusterInstallYamlsPath := "e2e-upgrade-cluster"
		kubeanNamespace := "kubean-system"
		kubeanClusterOpsName := "e2e-upgrade-cluster"

		// Create yaml for kuBean CR and related configuration
		installYamlPath := fmt.Sprint(tools.GetKuBeanPath(), clusterInstallYamlsPath)
		cmd := exec.Command("kubectl", "--kubeconfig="+tools.Kubeconfig, "apply", "-f", installYamlPath)
		out, _ := tools.DoCmd(*cmd)
		fmt.Println(out.String())

		// Check if the job and related pods have been created
		time.Sleep(30 * time.Second)
		config, _ = clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
		kubeClient, _ = kubernetes.NewForConfig(config)
		pods, _ := kubeClient.CoreV1().Pods(kubeanNamespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=kubean-%s-job", kubeanClusterOpsName),
		})
		gomega.Expect(len(pods.Items)).NotTo(gomega.Equal(0))
		jobPodName := pods.Items[0].Name

		// Wait for job-related pod status to be succeeded
		for {
			pod, err := kubeClient.CoreV1().Pods(kubeanNamespace).Get(context.Background(), jobPodName, metav1.GetOptions{})
			ginkgo.GinkgoWriter.Printf("* wait for upgrade job related pod[%s] status: %s\n", pod.Name, pod.Status.Phase)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed get job related pod")
			podStatus := string(pod.Status.Phase)
			if podStatus == "Succeeded" || podStatus == "Failed" {
				ginkgo.It("cluster upgrade job related pod Status should be Succeeded", func() {
					gomega.Expect(podStatus).To(gomega.Equal("Succeeded"))
				})
				break
			}
			time.Sleep(1 * time.Minute)
		}
	})

	time.Sleep(1 * time.Minute)

	// check kube version after upgrade
	ginkgo.Context("When fetching kube-system pods status after upgrade", func() {
		config, err = clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
		kubeClient, err = kubernetes.NewForConfig(config)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")
		podList, err := kubeClient.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed to check kube-system pod status")
		for _, pod := range podList.Items {
			for {
				po, _ := kubeClient.CoreV1().Pods("kube-system").Get(context.Background(), pod.Name, metav1.GetOptions{})
				ginkgo.GinkgoWriter.Printf("* wait for kube-system pod[%s] status: %s\n", po.Name, po.Status.Phase)
				podStatus := string(po.Status.Phase)
				if podStatus == "Running" || podStatus == "Failed" {
					ginkgo.It("every pod in kube-system after upgrade should be in running status", func() {
						gomega.Expect(podStatus).To(gomega.Equal("Running"))
					})
					break
				}
				time.Sleep(1 * time.Minute)
			}
		}
	})

	ginkgo.Context("check kube version after upgrade", func() {
		config, err = clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
		kubeClient, err = kubernetes.NewForConfig(config)
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")
		nodeList, _ := kubeClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		for _, node := range nodeList.Items {
			ginkgo.It("kube node version should be v1.23.7", func() {
				gomega.Expect(node.Status.NodeInfo.KubeletVersion).To(gomega.Equal("v1.23.7"))
				gomega.Expect(node.Status.NodeInfo.KubeProxyVersion).To(gomega.Equal("v1.23.7"))
			})
		}
	})

})
