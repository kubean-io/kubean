package tools

import (
	"flag"
)

var Kubeconfig string

func init() {
	flag.StringVar(&Kubeconfig, "kubeconfig", "", "cluster kubeconfig")
}

func FlagParse() {
	flag.Parse()

}
