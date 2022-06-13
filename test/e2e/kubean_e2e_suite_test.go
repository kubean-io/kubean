package e2e

import (
	"fmt"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"os"
	"testing"
)

func TestInsight(t *testing.T) {
	fmt.Println("KUBECONFIG:", os.Getenv("KUBECONFIG"))
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Kubean E2E installation test Suite")
}
