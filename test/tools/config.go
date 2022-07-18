package tools

import (
	"flag"
)

var Kubeconfig string
var Vmipaddr string

func init() {
	flag.StringVar(&Kubeconfig, "kubeconfig", "", "cluster kubeconfig")
	flag.StringVar(&Vmipaddr, "vmipaddr", "", "vm ip address")
}

func FlagParse() {
	flag.Parse()

}
