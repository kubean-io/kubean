package main

import (
	"os"

	"github.com/daocloud/kubean/cmd/kubean-operator/app"
	apiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"
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
