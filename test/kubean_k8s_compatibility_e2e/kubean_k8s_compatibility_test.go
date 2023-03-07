package kubean_k8s_compatibility_e2e

import (
	"github.com/kubean-io/kubean/test/tools"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var _ = ginkgo.Describe("e2e test compatibility 1 master + 1 worker", func() {
	// do cluster installation within docker
	ginkgo.Context("when install a  cluster using docker", func() {
		ginkgo.It("Create cluster and all kube-system pods be running", func() {
			clusterInstallYamlsPath := "e2e-install-cluster"
			kubeanClusterOpsName := tools.ClusterOperationName
			klog.Info(kubeanClusterOpsName)
			kindConfig, err := clientcmd.BuildConfigFromFlags("", tools.Kubeconfig)
			gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
			tools.OperateClusterByYaml(clusterInstallYamlsPath, kubeanClusterOpsName, kindConfig)
		})
	})
})
