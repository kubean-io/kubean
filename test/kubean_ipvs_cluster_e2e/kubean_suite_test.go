package kubean_ipvs_cluster_e2e_test

import (
	"testing"

	"github.com/kubean-io/kubean/test/tools"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func init() {
	testing.Init()
	tools.FlagParse()
}

func TestKuBean(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Test KuBean Suite")
}
