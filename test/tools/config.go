package tools

import (
	"flag"
)

var Kubeconfig string
var Vmipaddr string
var Vmipaddr2 string
var ClusterOperationName string
var IsOffline string
var Arch string
var VmPassword string

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

func init() {
	flag.StringVar(&Kubeconfig, "kubeconfig", "", "cluster kubeconfig")
	flag.StringVar(&ClusterOperationName, "clusterOperationName", "", "crd clusteroperation.kubean.io name")
	flag.StringVar(&IsOffline, "isOffline", "", "install k8s cluster offline or online")
	flag.StringVar(&Vmipaddr, "vmipaddr", "", "vm ip address")
	flag.StringVar(&Vmipaddr2, "vmipaddr2", "", "vm worker ip address")
	flag.StringVar(&Arch, "arch", "", "vm os arch")
	flag.StringVar(&VmPassword, "vmPassword", "", "vm login password")

}

func FlagParse() {
	flag.Parse()

}
