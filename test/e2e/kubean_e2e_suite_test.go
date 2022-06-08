package e2e

import (
	//"flag"
	//"fmt"
	"testing"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestInsight(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Kubean E2E installation test Suite")
}
