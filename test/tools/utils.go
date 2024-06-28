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
    containerd_registries_mirrors:
      - prefix: "10.6.170.10:5000"
        mirrors:
          - host: "http://10.6.170.10:5000"
            capabilities: ["pull", "resolve"]
            skip_verify: true
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

	github_url: "{{ files_repo }}/github.com"
	dl_k8s_io_url: "{{ files_repo }}/dl.k8s.io"
	storage_googleapis_url: "{{ files_repo }}/storage.googleapis.com"
	get_helm_url: "{{ files_repo }}/get.helm.sh"
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

var internalSource = `# offline
    registry_host: "10.6.170.10:5000"
    files_repo: "http://10.6.170.10:8080"
    kube_image_repo: "{{ registry_host }}"
    gcr_image_repo: "{{ registry_host }}"
    github_image_repo: "{{ registry_host }}"
    docker_image_repo: "{{ registry_host }}"
    quay_image_repo: "{{ registry_host }}"

    github_url: "{{ files_repo }}/github.com"
    dl_k8s_io_url: "{{ files_repo }}/dl.k8s.io"
    storage_googleapis_url: "{{ files_repo }}/storage.googleapis.com"
    get_helm_url: "{{ files_repo }}/get.helm.sh"
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
