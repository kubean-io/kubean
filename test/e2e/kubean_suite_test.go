package e2e

import (
	"github.com/kubean-io/kubean/test/tools"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"testing"
)

func init() {
	testing.Init()
	tools.FlagParse()
}

func TestKuBean(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Test KuBean Suite")
}
