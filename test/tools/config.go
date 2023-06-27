package tools

import (
	"flag"
)

var Kubeconfig string
var Vmipaddr string
var Vmipaddr2 string
var Vmipaddr3 string
var Vipadd string
var ClusterOperationName string
var IsOffline string
var Arch string
var VmPassword string
var OtherLabel string

// k8s const
const (
	PodStatusSucceeded = "Succeeded"
	PodStatusFailed    = "Failed"
	PodStatusRunning   = "Running"
)

// kubean_const
const (
	KubeanNamespace     = "kubean-system"
	KubeSystemNamespace = "kube-system"
	DefaultNamespace    = "default"
	TestClusterName     = "cluster1"
	LocalKubeConfigPath = "cluster1-config"
)

const (
	OriginK8Version             = "v1.25.3"
	UpgradeK8Version_Y          = "v1.26.0"
	UpgradeK8Version_Z          = "v1.26.5"
	NginxAlpha                  = "release-ci.daocloud.io/kubean/nginx:alpine"
	E2eInstallClusterYamlFolder = "e2e-install-cluster"
)

func init() {
	flag.StringVar(&Kubeconfig, "kubeconfig", "", "cluster kubeconfig")
	flag.StringVar(&ClusterOperationName, "clusterOperationName", "", "crd clusteroperation.kubean.io name")
	flag.StringVar(&IsOffline, "isOffline", "", "install k8s cluster offline or online")
	flag.StringVar(&Vmipaddr, "vmipaddr", "", "vm ip address")
	flag.StringVar(&Vmipaddr2, "vmipaddr2", "", "vm ip address")
	flag.StringVar(&Vmipaddr3, "vmipaddr3", "", "vm ip address")
	flag.StringVar(&Vipadd, "vipadd", "", "vip address")
	flag.StringVar(&Arch, "arch", "", "vm os arch")
	flag.StringVar(&VmPassword, "vmPassword", "", "vm login password")
	flag.StringVar(&OtherLabel, "otherLabel", "", "for not general label")
}

func FlagParse() {
	flag.Parse()

}
