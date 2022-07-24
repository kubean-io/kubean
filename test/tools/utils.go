package tools

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

func GetKuBeanPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, "kubean")
	// return path[:index] + "kubean/"
	return path[:index]
}

type KubeanOpsYml struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name   string `yaml:"name"`
		Labels struct {
			ClusterName string `yaml:"clusterName"`
		}
	}
	Spec struct {
		KuBeanCluster string `yaml:"kuBeanCluster"`
		Image         string `yaml:"image"`
		BackoffLimit  int    `yaml:"backoffLimit"`
		ActionType    string `yaml:"actionType"`
		Action        string `yaml:"action"`
	}
}

func UpdateOpsYml(content string, filePath string) {
	// read in Ops yaml file content
	yamlfileCotent, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("fail to read insight yml file: ", err)
	}
	var kubeanOpsYml KubeanOpsYml
	_ = yaml.Unmarshal(yamlfileCotent, &kubeanOpsYml)
	// modify ops name
	kubeanOpsYml.Metadata.Name = content
	data, _ := yaml.Marshal(kubeanOpsYml)
	// write back to yml file
	_ = ioutil.WriteFile(filePath, data, 0777)
}
