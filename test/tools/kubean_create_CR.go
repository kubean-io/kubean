package tools

import (
	"context"
	"fmt"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"kubean.io/api/apis"
	kubeancluster "kubean.io/api/apis/cluster/v1alpha1"
	kubeanclusterops "kubean.io/api/apis/clusteroperation/v1alpha1"
	kubeanClusterClientSet "kubean.io/api/generated/cluster/clientset/versioned"
	kubeanClusterOperationClientSet "kubean.io/api/generated/clusteroperation/clientset/versioned"
)

var hostsYaml = `
	all:
      hosts:
        node1:
          ip: %s
          access_ip: %s
          ansible_host: %s
          ansible_connection: ssh
          ansible_user: %s
          ansible_password: %s
      children:
        kube_control_plane:
          hosts:
            node1:
        kube_node:
          hosts:
            node1:
        etcd:
          hosts:
            node1:
        k8s_cluster:
          children:
            kube_control_plane:
            kube_node:
        calico_rr:
          hosts: {}`

var groupVarsTaml = `
	# k8s-cluster
    kube_version: "v1.23.7"
    cluster_name: kubean.cluster
    container_manager: containerd
    containerd_insecure_registries:
      "10.6.170.10:5000": "http://10.6.170.10:5000"
    k8s_image_pull_policy: IfNotPresent
    kube_network_plugin: calico
    kube_network_plugin_multus: false
    kube_apiserver_port: 6443
    kube_proxy_mode: iptables
    enable_nodelocaldns: false
    etcd_deployment_type: kubeadm
    metrics_server_enabled: true
    local_path_provisioner_enabled: true
 
    # Download Config
    download_run_once: true
    download_container: false
    download_force_cache: true
    download_localhost: true
     
    # offline
    registry_host: "10.6.170.10:5000"
    files_repo: "http://10.6.170.10:8080"
    kube_image_repo: "{{ registry_host }}"
    gcr_image_repo: "{{ registry_host }}"
    github_image_repo: "{{ registry_host }}"
    docker_image_repo: "{{ registry_host }}"
    quay_image_repo: "{{ registry_host }}"
    kubeadm_download_url: "{{ files_repo }}/storage.googleapis.com/kubernetes-release/release/{{ kubeadm_version }}/bin/linux/{{ image_arch }}/kubeadm"
    kubectl_download_url: "{{ files_repo }}/storage.googleapis.com/kubernetes-release/release/{{ kube_version }}/bin/linux/{{ image_arch }}/kubectl"
    kubelet_download_url: "{{ files_repo }}/storage.googleapis.com/kubernetes-release/release/{{ kube_version }}/bin/linux/{{ image_arch }}/kubelet"
    cni_download_url: "{{ files_repo }}/github.com/containernetworking/plugins/releases/download/{{ cni_version }}/cni-plugins-linux-{{ image_arch }}-{{ cni_version }}.tgz"
    crictl_download_url: "{{ files_repo }}/github.com/kubernetes-sigs/cri-tools/releases/download/{{ crictl_version }}/crictl-{{ crictl_version }}-{{ ansible_system | lower }}-{{ image_arch }}.tar.gz"
    etcd_download_url: "{{ files_repo }}/github.com/etcd-io/etcd/releases/download/{{ etcd_version }}/etcd-{{ etcd_version }}-linux-{{ image_arch }}.tar.gz"
    calicoctl_download_url: "{{ files_repo }}/github.com/projectcalico/calico/releases/download/{{ calico_ctl_version }}/calicoctl-linux-{{ image_arch }}"
    calicoctl_alternate_download_url: "{{ files_repo }}/github.com/projectcalico/calicoctl/releases/download/{{ calico_ctl_version }}/calicoctl-linux-{{ image_arch }}"
    calico_crds_download_url: "https://github.com/projectcalico/calico/archive/{{ calico_version }}.tar.gz"
    helm_download_url: "{{ files_repo }}/get.helm.sh/helm-{{ helm_version }}-linux-{{ image_arch }}.tar.gz"
    crun_download_url: "{{ files_repo }}/github.com/containers/crun/releases/download/{{ crun_version }}/crun-{{ crun_version }}-linux-{{ image_arch }}"
    kata_containers_download_url: "{{ files_repo }}/github.com/kata-containers/kata-containers/releases/download/{{ kata_containers_version }}/kata-static-{{ kata_containers_version }}-{{ ansible_architecture }}.tar.xz"
    runc_download_url: "{{ files_repo }}/github.com/opencontainers/runc/releases/download/{{ runc_version }}/runc.{{ image_arch }}"
    containerd_download_url: "{{ files_repo }}/github.com/containerd/containerd/releases/download/v{{ containerd_version }}/containerd-{{ containerd_version }}-linux-{{ image_arch }}.tar.gz"
    nerdctl_download_url: "{{ files_repo }}/github.com/containerd/nerdctl/releases/download/v{{ nerdctl_version }}/nerdctl-{{ nerdctl_version }}-{{ ansible_system | lower }}-{{ image_arch }}.tar.gz"
`

func CreatCR() {
	config, err := clientcmd.BuildConfigFromFlags("", "/home/actions-runner/.kube/kubean-latest-23022-host.config")
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
	kubeClient, err := kubernetes.NewForConfig(config)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")

	kubeanNamespace := "kubean-system"
	kubeanClusterOpsName := "e2e-cluster1-install-cr2"
	hostsConfigMapName := "cluster1-hosts-conf"
	varsConfigMapName := "cluster1-vars-conf"
	kubeClusterName := "cluster1"
	kubeClusterLabelName := "cluster1"

	// 1 create hosts configmap
	hostsYml := fmt.Sprintf(hostsYaml, "10.6.127.41", "10.6.127.41", "10.6.127.41", "root", "root")
	hostsConfigMapObj := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      hostsConfigMapName,
			Namespace: kubeanNamespace,
		},
		Data: map[string]string{
			"hosts.yml": hostsYml,
		},
	}
	configmapClient := kubeClient.CoreV1().ConfigMaps(kubeanNamespace)
	if _, err = configmapClient.Get(context.TODO(), hostsConfigMapName, metav1.GetOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			fmt.Println(err)
			return
		}
		result, err := configmapClient.Create(context.TODO(), hostsConfigMapObj, metav1.CreateOptions{})
		if err != nil {
			panic(err)
		}
		fmt.Printf("Created hostsConfigMapName %q.\n", result.GetObjectMeta().GetName())
	}

	// 2 create vars cnofigmap
	varsConfigMapObj := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      varsConfigMapName,
			Namespace: kubeanNamespace,
		},
		Data: map[string]string{
			"group_vars.yml": groupVarsTaml,
		},
	}
	if _, err = configmapClient.Get(context.TODO(), varsConfigMapName, metav1.GetOptions{}); err != nil {
		if !apierrors.IsNotFound(err) {
			fmt.Println(err)
			return
		}
		result, err := configmapClient.Create(context.TODO(), varsConfigMapObj, metav1.CreateOptions{})
		if err != nil {
			panic(err)
		}
		fmt.Printf("Created varsConfigMapName %q.\n", result.GetObjectMeta().GetName())
	}

	// 3 create Cluster
	clusterClientSet, err := kubeanClusterClientSet.NewForConfig(config)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")
	hostsDataRef := apis.ConfigMapRef{Name: hostsConfigMapName, NameSpace: kubeanNamespace}
	varsDataRef := apis.ConfigMapRef{Name: varsConfigMapName, NameSpace: kubeanNamespace}
	kubeclusterObj := &kubeancluster.Cluster{
		TypeMeta:   metav1.TypeMeta{Kind: "Cluster", APIVersion: "kubean.io/v1alpha1"},
		ObjectMeta: metav1.ObjectMeta{Name: kubeClusterName, Labels: map[string]string{"ClusterName": kubeClusterLabelName}},
		Spec:       kubeancluster.Spec{HostsConfRef: &hostsDataRef, VarsConfRef: &varsDataRef},
	}
	_, err = clusterClientSet.KubeanV1alpha1().Clusters().Get(context.Background(), kubeClusterName, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			fmt.Println(err)
			return
		}
		result, err := clusterClientSet.KubeanV1alpha1().Clusters().Create(context.Background(), kubeclusterObj, metav1.CreateOptions{})
		if err != nil {
			panic(err)
		}
		fmt.Printf("Created kubeClusterName %q.\n", result.GetObjectMeta().GetName())

	}

	// 4 create ClusterOperation
	clusterClientOperationSet, err := kubeanClusterOperationClientSet.NewForConfig(config)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")
	preHookAction :=
		`ansible -i /conf/hosts.yml all -m ping;
		ansible -i /conf/hosts.yml all -m shell -a 'systemctl stop firewalld && systemctl disable firewalld'
		ansible -i /conf/hosts.yml all -m shell -a 'yum install -y ntpdate && ntpdate cn.pool.ntp.org'`
	postHookAction := `ansible -i /conf/hosts.yml node1 -m shell -a 'kubectl get cs'`

	preHook := []kubeanclusterops.HookAction{{ActionType: kubeanclusterops.ShellActionType, Action: preHookAction}}
	postHook := []kubeanclusterops.HookAction{{ActionType: kubeanclusterops.ShellActionType, Action: postHookAction}}
	kubeclusterOpsObj := &kubeanclusterops.ClusterOperation{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterOperation",
			APIVersion: "kubean.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeanClusterOpsName,
			Labels: map[string]string{
				"clusterName": kubeClusterLabelName,
			},
		},
		Spec: kubeanclusterops.Spec{
			Cluster:      kubeClusterName,
			Image:        "ghcr.io/kubean-io/spray-job:v0.0.1",
			BackoffLimit: 0,
			ActionType:   kubeanclusterops.PlaybookActionType,
			Action:       "cluster.yaml",
			PreHook:      preHook,
			PostHook:     postHook,
		},
	}
	result1, err := clusterClientOpsSet.KubeanV1alpha1().ClusterOperations().Create(context.Background(), kubeclusterOpsObj, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created kubeanClusterOpsName %q.\n", result1.GetObjectMeta().GetName())
}

// var _ = ginkgo.Describe("e2e test cluster operation", func() {
// 	CreatCR()
// })
