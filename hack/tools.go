//go:build tools
// +build tools

package tools

import (
	_ "github.com/golang/mock/mockgen"
	_ "github.com/onsi/ginkgo/v2"
	_ "golang.org/x/tools/cmd/goimports"
	_ "k8s.io/code-generator"
)
