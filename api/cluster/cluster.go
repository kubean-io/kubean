package cluster

import (
	"github.com/kubean-io/kubean-api/constants"
	"k8s.io/klog/v2"
	"strconv"
)

type ConfigProperty struct {
	ClusterOperationsBackEndLimit string `json:"CLUSTER_OPERATIONS_BACKEND_LIMIT"`
	SprayJobImageRegistry         string `json:"SPRAY_JOB_IMAGE_REGISTRY"`
}

func (config *ConfigProperty) GetClusterOperationsBackEndLimit() int {
	value, _ := strconv.Atoi(config.ClusterOperationsBackEndLimit)
	if value <= 0 {
		klog.Warningf("GetClusterOperationsBackEndLimit and use default value %d", constants.DefaultClusterOperationsBackEndLimit)
		return constants.DefaultClusterOperationsBackEndLimit
	}
	if value >= constants.MaxClusterOperationsBackEndLimit {
		klog.Warningf("GetClusterOperationsBackEndLimit and use max value %d", constants.MaxClusterOperationsBackEndLimit)
		return constants.MaxClusterOperationsBackEndLimit
	}
	return value
}
