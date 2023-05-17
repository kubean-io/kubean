package tools

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func GetKuBeanPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, "kubean")
	// return path[:index] + "kubean/"
	return path[:index]
}

type OfflineConfig struct {
	Ip              string `yaml:"ip"`
	RegistryAddr    string `yaml:"registry_addr"`
	MinioAddr       string `yaml:"minio_addr"`
	NginxImageAMD64 string `yaml:"nginx_image_amd64"`
	NginxImageARM64 string `yaml:"nginx_image_arm64"`
}

type KubeanOpsYml struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name   string `yaml:"name"`
		Labels struct {
			ClusterName string `yaml:"clusterName"`
		}
	}
	Spec struct {
		Cluster      string `yaml:"cluster"`
		Image        string `yaml:"image"`
		BackoffLimit int    `yaml:"backoffLimit"`
		ActionType   string `yaml:"actionType"`
		Action       string `yaml:"action"`
	}
}

//go:embed offline_params.yml
var OfflineConfigStr string

func InitOfflineConfig() OfflineConfig {
	configStr := OfflineConfigStr
	var offlineConfig OfflineConfig

	err := yaml.Unmarshal([]byte(configStr), &offlineConfig)
	CheckErr(err, "yaml unmarshal error")
	return offlineConfig
}

func UpdateOpsYml(content string, filePath string) {
	// read in Ops yaml file content
	yamlfileCotent, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("fail to read insight yml file: ", err)
	}
	var kubeanOpsYml KubeanOpsYml
	_ = yaml.Unmarshal(yamlfileCotent, &kubeanOpsYml)
	// modify ops name
	kubeanOpsYml.Metadata.Name = content
	data, _ := yaml.Marshal(kubeanOpsYml)
	// write back to yml file
	_ = os.WriteFile(filePath, data, 0777)
}

func UpdateBackoffLimit(content int, filePath string) {
	// read in Ops yaml file content
	yamlfileCotent, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("fail to read insight yml file: ", err)
	}
	var kubeanOpsYml KubeanOpsYml
	_ = yaml.Unmarshal(yamlfileCotent, &kubeanOpsYml)
	// modify BackoffLimit
	kubeanOpsYml.Spec.BackoffLimit = content
	data, _ := yaml.Marshal(kubeanOpsYml)
	// write back to yml file
	_ = os.WriteFile(filePath, data, 0777)
}

func DoCmd(cmd exec.Cmd) (bytes.Buffer, bytes.Buffer) {
	ginkgo.GinkgoWriter.Printf("cmd: %s\n", cmd.String())
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		ginkgo.GinkgoWriter.Printf("apply cmd error: %s\n", err.Error())
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), stderr.String())
	}
	return out, stderr
}

func NewDoCmd(cmd string, args ...string) (bytes.Buffer, bytes.Buffer) {
	icmd := exec.Command(cmd, args...)
	fmt.Println("NewDoCmd: ", icmd.String())
	var out, stderr bytes.Buffer
	icmd.Stdout = &out
	icmd.Stderr = &stderr
	if err := icmd.Run(); err != nil {
		ginkgo.GinkgoWriter.Printf("apply cmd error: %s\n", err.Error())
		gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), stderr.String())
	}
	return out, stderr
}

func NewDoCmdSoft(cmd string, args ...string) (bytes.Buffer, error) {
	icmd := exec.Command(cmd, args...)
	fmt.Println("NewDoCmd: ", icmd.String())
	var out, stderr bytes.Buffer
	icmd.Stdout = &out
	icmd.Stderr = &stderr

	err := icmd.Run()
	if err != nil {
		ginkgo.GinkgoWriter.Printf("apply cmd error: %s\n", err.Error())
		return out, err
	}
	return out, nil
}

func DoErrCmd(cmd exec.Cmd) (bytes.Buffer, bytes.Buffer) {
	ginkgo.GinkgoWriter.Printf("cmd: %s\n", cmd.String())
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		ginkgo.GinkgoWriter.Printf("apply cmd error: %s\n", err.Error())
	}
	return out, stderr
}

func RemoteSSHCmdArray(subCmd []string) []string {
	var CmdArray = []string{"-p", "root", "ssh", "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no"}
	return append(CmdArray, subCmd...)
}

func RemoteSSHCmdArrayByPasswd(password string, subCmd []string) []string {
	var CmdArray = []string{"-p", password, "ssh", "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no"}
	return append(CmdArray, subCmd...)
}

var groupVarsYaml = `
	kube_version: %s
	container_manager: containerd
	containerd_insecure_registries:
		"10.6.170.10:5000": "http://10.6.170.10:5000"
	k8s_image_pull_policy: IfNotPresent
	kube_network_plugin: %s
	kube_network_plugin_multus: false
	kube_apiserver_port: 6443
	kube_proxy_mode: iptables
	enable_nodelocaldns: false
	etcd_deployment_type: kubeadm
	metrics_server_enabled: true
	auto_renew_certificates: true
	local_path_provisioner_enabled: true
	ntp_enabled: true

	kube_service_addresses: %s
	kube_pods_subnet: %s
	kube_network_node_prefix: %d

	calico_cni_name: calico
	calico_felix_premetheusmetricsenabled: true

	# Download Config
	download_run_once: true
	download_container: false
	download_force_cache: true
	download_localhost: true

	%s
	
	# offline
	registry_host: "10.6.170.10:5000"
	files_repo: "http://10.6.170.10:8080"
	kube_image_repo: "{{ registry_host }}"
	gcr_image_repo: "{{ registry_host }}"
	github_image_repo: "{{ registry_host }}"
	docker_image_repo: "{{ registry_host }}"
	quay_image_repo: "{{ registry_host }}"
	kubeadm_download_url: "{{ files_repo }}/dl.k8s.io/release/{{ kubeadm_version }}/bin/linux/{{ image_arch }}/kubeadm"
	kubectl_download_url: "{{ files_repo }}/dl.k8s.io/release/{{ kube_version }}/bin/linux/{{ image_arch }}/kubectl"
	kubelet_download_url: "{{ files_repo }}/dl.k8s.io/release/{{ kube_version }}/bin/linux/{{ image_arch }}/kubelet"
	cni_download_url: "{{ files_repo }}/github.com/containernetworking/plugins/releases/download/{{ cni_version }}/cni-plugins-linux-{{ image_arch }}-{{ cni_version }}.tgz"
	crictl_download_url: "{{ files_repo }}/github.com/kubernetes-sigs/cri-tools/releases/download/{{ crictl_version }}/crictl-{{ crictl_version }}-{{ ansible_system | lower }}-{{ image_arch }}.tar.gz"
	etcd_download_url: "{{ files_repo }}/github.com/etcd-io/etcd/releases/download/{{ etcd_version }}/etcd-{{ etcd_version }}-linux-{{ image_arch }}.tar.gz"
	calicoctl_download_url: "{{ files_repo }}/github.com/projectcalico/calico/releases/download/{{ calico_ctl_version }}/calicoctl-linux-{{ image_arch }}"
	calicoctl_alternate_download_url: "{{ files_repo }}/github.com/projectcalico/calicoctl/releases/download/{{ calico_ctl_version }}/calicoctl-linux-{{ image_arch }}"
    calico_crds_download_url: "{{ files_repo }}/github.com/projectcalico/calico/archive/{{ calico_version }}.tar.gz"
	helm_download_url: "{{ files_repo }}/get.helm.sh/helm-{{ helm_version }}-linux-{{ image_arch }}.tar.gz"
	crun_download_url: "{{ files_repo }}/github.com/containers/crun/releases/download/{{ crun_version }}/crun-{{ crun_version }}-linux-{{ image_arch }}"
	kata_containers_download_url: "{{ files_repo }}/github.com/kata-containers/kata-containers/releases/download/{{ kata_containers_version }}/kata-static-{{ kata_containers_version }}-{{ ansible_architecture }}.tar.xz"
	runc_download_url: "{{ files_repo }}/github.com/opencontainers/runc/releases/download/{{ runc_version }}/runc.{{ image_arch }}"
	containerd_download_url: "{{ files_repo }}/github.com/containerd/containerd/releases/download/v{{ containerd_version }}/containerd-{{ containerd_version }}-linux-{{ image_arch }}.tar.gz"
	nerdctl_download_url: "{{ files_repo }}/github.com/containerd/nerdctl/releases/download/v{{ nerdctl_version }}/nerdctl-{{ nerdctl_version }}-{{ ansible_system | lower }}-{{ image_arch }}.tar.gz"
	`

func CreatVarsCM(subStr string) {

	var groupVarsYamlString = fmt.Sprintf(groupVarsYaml, "1.23.7", "calico", "10.96.0.0/12", "192.168.128.0/20", 24, subStr)

	config, err := clientcmd.BuildConfigFromFlags("", Kubeconfig)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed build config")
	kubeClient, err := kubernetes.NewForConfig(config)
	gomega.ExpectWithOffset(2, err).NotTo(gomega.HaveOccurred(), "failed new client set")

	kubeanNamespace := "kubean-system"
	varsConfigMapName := "cluster1-vars-conf"
	configmapClient := kubeClient.CoreV1().ConfigMaps(kubeanNamespace)
	// create vars-conf-cm cnofigmap
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
			"group_vars.yml": groupVarsYamlString,
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
}

// var openSource = `# gcr and kubernetes image repo define
// gcr_image_repo: "gcr.m.daocloud.io"
// kube_image_repo: "k8s.m.daocloud.io"

// # docker image repo define
// docker_image_repo: "docker.m.daocloud.io"

// # quay image repo define
// quay_image_repo: "quay.m.daocloud.io"

// # github image repo define (ex multus only use that)
// github_image_repo: "ghcr.m.daocloud.io"

// files_repo: "https://files.m.daocloud.io"

// ## Kubernetes components
// kubeadm_download_url: "{{ files_repo }}/dl.k8s.io/release/{{ kubeadm_version }}/bin/linux/{{ image_arch }}/kubeadm"
// kubectl_download_url: "{{ files_repo }}/dl.k8s.io/release/{{ kube_version }}/bin/linux/{{ image_arch }}/kubectl"
// kubelet_download_url: "{{ files_repo }}/dl.k8s.io/release/{{ kube_version }}/bin/linux/{{ image_arch }}/kubelet"

// ## CNI Plugins
// cni_download_url: "{{ files_repo }}/github.com/containernetworking/plugins/releases/download/{{ cni_version }}/cni-plugins-linux-{{ image_arch }}-{{ cni_version }}.tgz"

// ## cri-tools
// crictl_download_url: "{{ files_repo }}/github.com/kubernetes-sigs/cri-tools/releases/download/{{ crictl_version }}/crictl-{{ crictl_version }}-{{ ansible_system | lower }}-{{ image_arch }}.tar.gz"

// ## [Optional] etcd: only if you **DON'T** use etcd_deployment=host
// etcd_download_url: "{{ files_repo }}/github.com/etcd-io/etcd/releases/download/{{ etcd_version }}/etcd-{{ etcd_version }}-linux-{{ image_arch }}.tar.gz"

// # [Optional] Calico: If using Calico network plugin
// calicoctl_download_url: "{{ files_repo }}/github.com/projectcalico/calico/releases/download/{{ calico_ctl_version }}/calicoctl-linux-{{ image_arch }}"
// calicoctl_alternate_download_url: "{{ files_repo }}/github.com/projectcalico/calicoctl/releases/download/{{ calico_ctl_version }}/calicoctl-linux-{{ image_arch }}"
// # [Optional] Calico with kdd: If using Calico network plugin with kdd datastore
// calico_crds_download_url: "{{ files_repo }}/github.com/projectcalico/calico/archive/{{ calico_version }}.tar.gz"

// # [Optional] Flannel: If using Falnnel network plugin
// flannel_cni_download_url: "{{ files_repo }}/kubernetes/flannel/{{ flannel_cni_version }}/flannel-{{ image_arch }}"

// # [Optional] helm: only if you set helm_enabled: true
// helm_download_url: "{{ files_repo }}/get.helm.sh/helm-{{ helm_version }}-linux-{{ image_arch }}.tar.gz"

// # [Optional] crun: only if you set crun_enabled: true
// crun_download_url: "{{ files_repo }}/github.com/containers/crun/releases/download/{{ crun_version }}/crun-{{ crun_version }}-linux-{{ image_arch }}"

// # [Optional] kata: only if you set kata_containers_enabled: true
// kata_containers_download_url: "{{ files_repo }}/github.com/kata-containers/kata-containers/releases/download/{{ kata_containers_version }}/kata-static-{{ kata_containers_version }}-{{ ansible_architecture }}.tar.xz"

// # [Optional] cri-dockerd: only if you set container_manager: docker
// cri_dockerd_download_url: "{{ files_repo }}/github.com/Mirantis/cri-dockerd/releases/download/v{{ cri_dockerd_version }}/cri-dockerd-{{ cri_dockerd_version }}.{{ image_arch }}.tgz"

// # [Optional] runc,containerd: only if you set container_runtime: containerd
// runc_download_url: "{{ files_repo }}/github.com/opencontainers/runc/releases/download/{{ runc_version }}/runc.{{ image_arch }}"
// containerd_download_url: "{{ files_repo }}/github.com/containerd/containerd/releases/download/v{{ containerd_version }}/containerd-{{ containerd_version }}-linux-{{ image_arch }}.tar.gz"
// nerdctl_download_url: "{{ files_repo }}/github.com/containerd/nerdctl/releases/download/v{{ nerdctl_version }}/nerdctl-{{ nerdctl_version }}-{{ ansible_system | lower }}-{{ image_arch }}.tar.gz"
// cri_dockerd_download_url: "{{ files_repo }}/github.com/Mirantis/cri-dockerd/releases/download/v{{ cri_dockerd_version }}/cri-dockerd-{{ cri_dockerd_version }}.{{ image_arch }}.tgz"
// `

var internalSource = `# offline
    registry_host: "10.6.170.10:5000"
    files_repo: "http://10.6.170.10:8080"
    kube_image_repo: "{{ registry_host }}"
    gcr_image_repo: "{{ registry_host }}"
    github_image_repo: "{{ registry_host }}"
    docker_image_repo: "{{ registry_host }}"
    quay_image_repo: "{{ registry_host }}"
    kubeadm_download_url: "{{ files_repo }}/dl.k8s.io/release/{{ kubeadm_version }}/bin/linux/{{ image_arch }}/kubeadm"
    kubectl_download_url: "{{ files_repo }}/dl.k8s.io/release/{{ kube_version }}/bin/linux/{{ image_arch }}/kubectl"
    kubelet_download_url: "{{ files_repo }}/dl.k8s.io/release/{{ kube_version }}/bin/linux/{{ image_arch }}/kubelet"
    cni_download_url: "{{ files_repo }}/github.com/containernetworking/plugins/releases/download/{{ cni_version }}/cni-plugins-linux-{{ image_arch }}-{{ cni_version }}.tgz"
    crictl_download_url: "{{ files_repo }}/github.com/kubernetes-sigs/cri-tools/releases/download/{{ crictl_version }}/crictl-{{ crictl_version }}-{{ ansible_system | lower }}-{{ image_arch }}.tar.gz"
    etcd_download_url: "{{ files_repo }}/github.com/etcd-io/etcd/releases/download/{{ etcd_version }}/etcd-{{ etcd_version }}-linux-{{ image_arch }}.tar.gz"
    calicoctl_download_url: "{{ files_repo }}/github.com/projectcalico/calico/releases/download/{{ calico_ctl_version }}/calicoctl-linux-{{ image_arch }}"
    calicoctl_alternate_download_url: "{{ files_repo }}/github.com/projectcalico/calicoctl/releases/download/{{ calico_ctl_version }}/calicoctl-linux-{{ image_arch }}"
    calico_crds_download_url: "{{ files_repo }}/github.com/projectcalico/calico/archive/{{ calico_version }}.tar.gz"
    helm_download_url: "{{ files_repo }}/get.helm.sh/helm-{{ helm_version }}-linux-{{ image_arch }}.tar.gz"
    crun_download_url: "{{ files_repo }}/github.com/containers/crun/releases/download/{{ crun_version }}/crun-{{ crun_version }}-linux-{{ image_arch }}"
    kata_containers_download_url: "{{ files_repo }}/github.com/kata-containers/kata-containers/releases/download/{{ kata_containers_version }}/kata-static-{{ kata_containers_version }}-{{ ansible_architecture }}.tar.xz"
    runc_download_url: "{{ files_repo }}/github.com/opencontainers/runc/releases/download/{{ runc_version }}/runc.{{ image_arch }}"
    containerd_download_url: "{{ files_repo }}/github.com/containerd/containerd/releases/download/v{{ containerd_version }}/containerd-{{ containerd_version }}-linux-{{ image_arch }}.tar.gz"
    nerdctl_download_url: "{{ files_repo }}/github.com/containerd/nerdctl/releases/download/v{{ nerdctl_version }}/nerdctl-{{ nerdctl_version }}-{{ ansible_system | lower }}-{{ image_arch }}.tar.gz"
    cri_dockerd_download_url: "{{ files_repo }}/github.com/Mirantis/cri-dockerd/releases/download/v{{ cri_dockerd_version }}/cri-dockerd-{{ cri_dockerd_version }}.{{ image_arch }}.tgz"
`

var kubeServiceAddresses = "10.96.0.0/12"
var kubePodsSubnet = "192.168.128.0/20"

var varsConfCMYml = `apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster1-vars-conf
  namespace: kubean-system
data:
  group_vars.yml: |
    # k8s-cluster
    kube_version: "%s"
    container_manager: docker
    docker_insecure_registries:
    #  "10.6.170.10:5000": "http://10.6.170.10:5000"
      - 10.6.170.10:5000
    k8s_image_pull_policy: IfNotPresent
    kube_network_plugin: %s
    kube_network_plugin_multus: false
    kube_apiserver_port: 6443
    kube_proxy_mode: iptables
    enable_nodelocaldns: false
    etcd_deployment_type: kubeadm
    metrics_server_enabled: true
    auto_renew_certificates: true
    local_path_provisioner_enabled: true
    ntp_enabled: true

    kube_service_addresses: %s
    kube_pods_subnet: %s
    kube_network_node_prefix: %d
    
    %s
    
    calico_cni_name: calico
    calico_felix_premetheusmetricsenabled: true
 
    # Download Config
    download_run_once: true
    download_container: false
    download_force_cache: true
    download_localhost: true
    %s
`

func CreatVarsCMFile(subStr string) string {

	var groupVarsYamlString = fmt.Sprintf(varsConfCMYml, "v1.23.7", "calico", kubeServiceAddresses, kubePodsSubnet, 24, subStr, internalSource)
	return groupVarsYamlString
}

func CheckErr(err error, explain ...string) {
	if err != nil {
		if len(explain) > 0 {
			klog.Fatalf("%s:%s", explain[0], err.Error())
		} else {
			klog.Fatal(err)
		}
	}
}
