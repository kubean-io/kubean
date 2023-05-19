package util

import (
	"fmt"
	"os"
	"strings"

	clusterv1alpha1 "github.com/kubean-io/kubean-api/apis/cluster/v1alpha1"
	manifestv1alpha1 "github.com/kubean-io/kubean-api/apis/manifest/v1alpha1"

	clusteroperationv1alpha1 "github.com/kubean-io/kubean-api/apis/clusteroperation/v1alpha1"
	localartifactsetv1alpha1 "github.com/kubean-io/kubean-api/apis/localartifactset/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

// aggregatedScheme aggregates Kubernetes and extended schemes.
var aggregatedScheme = runtime.NewScheme()

func init() {
	_ = scheme.AddToScheme(aggregatedScheme)                   // add Kubernetes schemes
	_ = clusteroperationv1alpha1.AddToScheme(aggregatedScheme) // add clusterOps schemes
	_ = clusterv1alpha1.AddToScheme(aggregatedScheme)          // add cluster schemes
	_ = localartifactsetv1alpha1.AddToScheme(aggregatedScheme)
	_ = manifestv1alpha1.AddToScheme(aggregatedScheme)
}

// NewSchema returns a singleton schema set which aggregated Kubernetes's schemes and extended schemes.
func NewSchema() *runtime.Scheme {
	return aggregatedScheme
}

//// NewForConfig creates a new client for the given config.
//func NewForConfig(config *rest.Config) (client.Client, error) {
//	return client.New(config, client.Options{
//		Scheme: aggregatedScheme,
//	})
//}

var ServiceAccountNamespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

// GetCurrentNS fetch namespace the current pod running in. reference to client-go (config *inClusterClientConfig) Namespace() (string, bool, error).
func GetCurrentNS() (string, error) {
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns, nil
	}

	if data, err := os.ReadFile(ServiceAccountNamespaceFile); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns, nil
		}
	}
	return "", fmt.Errorf("can not get namespace where pods running in")
}

func GetCurrentNSOrDefault() string {
	ns, err := GetCurrentNS()
	if err != nil {
		return "default"
	}
	return ns
}
