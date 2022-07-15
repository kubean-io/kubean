package main

import (
	"os"

	apiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"

	"github.com/kubean-io/kubean/cmd/kubean-operator/app"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()
	ctx := apiserver.SetupSignalContext()
	if err := app.NewCommand(ctx).Execute(); err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}
}
