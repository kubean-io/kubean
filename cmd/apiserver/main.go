package main

import (
	"log"
	_ "net/http/pprof"

	apiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/component-base/logs"

	"github.com/daocloud/kubean/cmd/apiserver/app"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	ctx := apiserver.SetupSignalContext()

	cmd := app.NewAPIServerCommand(ctx)

	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
