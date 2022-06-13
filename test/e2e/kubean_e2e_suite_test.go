package e2e

import (
	"testing"
	"os"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)
fmt.Println("KUBECONFIG:", os.Getenv("KUBECONFIG"))

func TestInsight(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Kubean E2E installation test Suite")
}
