package e2e

import (
	"context"
	"fmt"
	"strings"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = ginkgo.Describe("Test Insight Pods Info", func() {
	kubeconfig := insightyml.KubeconfigFile
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	tools.CheckError(err)
	kubeClient, err := kubernetes.NewForConfig(config)
	tools.CheckError(err)
	NamespaceName := "insight-system"

	insightNS, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "kube-system", metav1.GetOptions{})
	tools.CheckError(err)
	tools.UpdateInsightYml(string(insightNS.ObjectMeta.UID))

	ginkgo.Context("When fetching pods status", func() {
		podList, err := kubeClient.CoreV1().Pods(NamespaceName).List(context.TODO(), metav1.ListOptions{})
		tools.CheckError(err)
		ginkgo.It("every pod should be in running status", func() {
			for _, pod := range podList.Items {
				if strings.Contains(pod.ObjectMeta.Name, "vmstorage-insight-victoria-metrics-k8s-stack") || strings.Contains(pod.ObjectMeta.Name, "vmselect-insight-victoria-metrics-k8s-stack") ||
					strings.Contains(pod.ObjectMeta.Name, "node-problem-detector") || strings.Contains(pod.ObjectMeta.Name, "insight-jaeger") {
					continue
				}
				fmt.Println(pod.Name, string(pod.Status.Phase))
				gomega.Expect(string(pod.Status.Phase)).To(gomega.Equal("Running"))
			}
		})
	})
	defer ginkgo.GinkgoRecover()

	ginkgo.Context("When fetching insight service info", func() {
		insightServiceList, err := kubeClient.CoreV1().Services(NamespaceName).Get(context.TODO(), "insight", metav1.GetOptions{})
		tools.CheckError(err)
		ginkgo.It("Insight service type should be NodePort", func() {
			gomega.Expect(string(insightServiceList.Spec.Type)).To(gomega.Equal("NodePort"))
		})

	})

	ginkgo.Context("When fetching insight statefulset info", func() {
		statefulsetList, err := kubeClient.AppsV1().StatefulSets(NamespaceName).List(context.TODO(), metav1.ListOptions{})
		tools.CheckError(err)
		for _, sf := range statefulsetList.Items {
			ginkgo.It("Insight statefulset should be ready", func() {
				gomega.Expect(sf.Status.ReadyReplicas).To(gomega.Equal(sf.Status.Replicas))
			})
		}
	})

	ginkgo.Context("When fetching insight deployment info", func() {
		deploymentList, err := kubeClient.AppsV1().Deployments(NamespaceName).List(context.TODO(), metav1.ListOptions{})
		tools.CheckError(err)
		for _, dm := range deploymentList.Items {
			ginkgo.It("Insight deployment should be ready", func() {
				gomega.Expect(dm.Status.ReadyReplicas).To(gomega.Equal(dm.Status.AvailableReplicas))
			})
		}
	})

	ginkgo.Context("When fetching insight daemonset info", func() {
		daemonsetList, err := kubeClient.AppsV1().DaemonSets(NamespaceName).List(context.TODO(), metav1.ListOptions{})
		tools.CheckError(err)
		for _, ds := range daemonsetList.Items {
			ginkgo.It("Insight daemonset should be ready", func() {
				gomega.Expect(ds.Status.CurrentNumberScheduled).To(gomega.Equal(ds.Status.DesiredNumberScheduled))
			})
		}
	})

})
