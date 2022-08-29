package tools

import (
	"flag"
)

var Kubeconfig string
var Vmipaddr string
var Vmipaddr2 string

func init() {
	flag.StringVar(&Kubeconfig, "kubeconfig", "", "cluster kubeconfig")
	flag.StringVar(&Vmipaddr, "vmipaddr", "", "vm ip address")
	flag.StringVar(&Vmipaddr2, "vmipaddr2", "", "vm worker ip address")

}

func FlagParse() {
	flag.Parse()

}
