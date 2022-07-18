package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	kubeanclusterv1alpha1 "kubean.io/api/apis/kubeancluster/v1alpha1"
	kubeanclusteropsv1alpha1 "kubean.io/api/apis/kubeanclusterops/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// aggregatedScheme aggregates Kubernetes and extended schemes.
var aggregatedScheme = runtime.NewScheme()

func init() {
	_ = scheme.AddToScheme(aggregatedScheme)                   // add Kubernetes schemes
	_ = kubeanclusteropsv1alpha1.AddToScheme(aggregatedScheme) // add clusterOps schemes
	_ = kubeanclusterv1alpha1.AddToScheme(aggregatedScheme)    // add cluster schemes
	//_ = clusterapiv1alpha4.AddToScheme(aggregatedScheme) // add cluster-api v1alpha4 schemes
	//_ = v1alpha2.AddToScheme(aggregatedScheme) // add clusterpedia-api v1alpha2 schemes
}

// NewSchema returns a singleton schema set which aggregated Kubernetes's schemes and extended schemes.
func NewSchema() *runtime.Scheme {
	return aggregatedScheme
}

// NewForConfig creates a new client for the given config.
func NewForConfig(config *rest.Config) (client.Client, error) {
	return client.New(config, client.Options{
		Scheme: aggregatedScheme,
	})
}

// GetCurrentNS fetch namespace the current pod running in. reference to client-go (config *inClusterClientConfig) Namespace() (string, bool, error).
func GetCurrentNS() (string, error) {
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns, nil
	}

	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
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
