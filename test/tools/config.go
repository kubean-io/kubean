package tools

import (
	"flag"
)

var Kubeconfig string
var Vmipaddr string
var Vmipaddr2 string

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
)

func init() {
	flag.StringVar(&Kubeconfig, "kubeconfig", "", "cluster kubeconfig")
	flag.StringVar(&Vmipaddr, "vmipaddr", "", "vm ip address")
	flag.StringVar(&Vmipaddr2, "vmipaddr2", "", "vm worker ip address")

}

func FlagParse() {
	flag.Parse()

}
