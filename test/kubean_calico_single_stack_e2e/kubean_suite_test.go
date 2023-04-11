package kubean_calico_single_stack_e2e

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
