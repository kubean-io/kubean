package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetKuBeanPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, "kubean")
	return path[:index] + "kubean/"
}
