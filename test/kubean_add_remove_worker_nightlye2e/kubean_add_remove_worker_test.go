package add_worker_e2e

import (
	"github.com/kubean-io/kubean/test/tools"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var _ = ginkgo.Describe("e2e add worker node operation", func() {
	localKubeConfigPath := "cluster1.config"

	ginkgo.Context("precondition: deploy one node cluster using private key file", func() {
		var offlineConfigs tools.OfflineConfig
		testClusterName := tools.TestClusterName
		offlineFlag := tools.IsOffline
		offlineConfigs = tools.InitOfflineConfig()

		klog.Info("offlineFlag is: ", offlineFlag)
		klog.Info("arch is: ", tools.Arch)
		klog.Info("arch is: ", offlineConfigs)

		ginkgo.It("Create cluster and all kube-system pods be running", func() {
			clusterInstallYamlsPath := "e2e-install-1node-cluster-prikey"
			kubeanClusterOpsName := tools.ClusterOperationName
			kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
			tools.OperateClusterByYaml(clusterInstallYamlsPath, kubeanClusterOpsName, kindConfig)
			tools.SaveKubeConf(kindConfig, testClusterName, localKubeConfigPath)
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Client")
			// Wait all pods in kube-syste to be Running
			tools.WaitPodSInKubeSystemBeRunning(cluster1Client, 1800)
		})

		ginkgo.It("Add a worker to cluster", func() {
			clusterInstallYamlsPath := "add-worker-node"
			kubeanClusterOpsName := "cluster1-add-worker"
			kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
			tools.OperateClusterByYaml(clusterInstallYamlsPath, kubeanClusterOpsName, kindConfig)
			tools.SaveKubeConf(kindConfig, testClusterName, localKubeConfigPath)
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			tools.WaitPodSInKubeSystemBeRunning(cluster1Client, 1800)
		})

		ginkgo.It("Remove a worker after add it", func() {
			clusterInstallYamlsPath := "remove-worker-node"
			kubeanClusterOpsName := "cluster1-remove-worker"
			kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
			tools.OperateClusterByYaml(clusterInstallYamlsPath, kubeanClusterOpsName, kindConfig)
			tools.SaveKubeConf(kindConfig, testClusterName, localKubeConfigPath)
			cluster1Config, err := clientcmd.BuildConfigFromFlags("", localKubeConfigPath)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "Failed new cluster1Config set")
			cluster1Client, err := kubernetes.NewForConfig(cluster1Config)
			tools.WaitPodSInKubeSystemBeRunning(cluster1Client, 1800)
		})
	})
})
