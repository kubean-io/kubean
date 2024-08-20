// Copyright 2023 Authors of kubean-io
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/kubean-io/kubean-api/apis"
	clusterv1alpha1 "github.com/kubean-io/kubean-api/apis/cluster/v1alpha1"
	clusteroperationv1alpha1 "github.com/kubean-io/kubean-api/apis/clusteroperation/v1alpha1"
	localartifactsetv1alpha1 "github.com/kubean-io/kubean-api/apis/localartifactset/v1alpha1"
	manifestv1alpha1 "github.com/kubean-io/kubean-api/apis/manifest/v1alpha1"
	"github.com/kubean-io/kubean-api/cluster"
	"github.com/kubean-io/kubean-api/constants"
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

func UpdateOwnReference(client kubernetes.Interface, configMapList []*apis.ConfigMapRef, secretList []*apis.SecretRef, belongToReference metav1.OwnerReference) error {
	for _, ref := range configMapList {
		if ref.IsEmpty() {
			continue
		}
		cm, err := client.CoreV1().ConfigMaps(ref.NameSpace).Get(context.Background(), ref.Name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) { // ignore
				continue
			}
			return err // not ignore
		}
		if len(cm.OwnerReferences) != 0 {
			continue // do nothing
		}
		// cm belongs to `Cluster`
		cm.OwnerReferences = append(cm.OwnerReferences, belongToReference)
		if _, err := client.CoreV1().ConfigMaps(ref.NameSpace).Update(context.Background(), cm, metav1.UpdateOptions{}); err != nil {
			return err
		}
	}
	for _, ref := range secretList {
		if ref.IsEmpty() {
			continue
		}
		secret, err := client.CoreV1().Secrets(ref.NameSpace).Get(context.Background(), ref.Name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) { // ignore
				continue
			}
			return err // not ignore
		}
		if len(secret.OwnerReferences) != 0 {
			continue // do nothing
		}
		secret.OwnerReferences = append(secret.OwnerReferences, belongToReference)
		if _, err := client.CoreV1().Secrets(ref.NameSpace).Update(context.Background(), secret, metav1.UpdateOptions{}); err != nil {
			return err
		}
	}
	return nil
}

func GetCurrentRunningPodName() string {
	if name := os.Getenv("HOSTNAME"); name != "" {
		return name
	}
	return "default"
}

func FetchKubeanConfigProperty(client kubernetes.Interface) *cluster.ConfigProperty {
	configData, err := client.CoreV1().ConfigMaps(GetCurrentNSOrDefault()).Get(context.Background(), constants.KubeanConfigMapName, metav1.GetOptions{})
	if err != nil {
		return &cluster.ConfigProperty{}
	}
	jsonData, err := json.Marshal(configData.Data)
	if err != nil {
		return &cluster.ConfigProperty{}
	}
	result := &cluster.ConfigProperty{}
	err = json.Unmarshal(jsonData, result)
	if err != nil {
		return &cluster.ConfigProperty{}
	}
	return result
}
