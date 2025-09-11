package crypto

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	klog "k8s.io/klog/v2"

	"github.com/kubean-io/kubean-api/constants"
	"github.com/kubean-io/kubean/pkg/util"
)

const (
	PrivateKey = "sk"
	PublicKey  = "pk"
)

func InitConfiguration(clientset kubernetes.Interface) error {
	kubeanConfig, err := clientset.CoreV1().ConfigMaps(util.GetCurrentNSOrDefault()).Get(context.Background(), constants.KubeanConfigMapName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if privateKey, ok := kubeanConfig.Data[PrivateKey]; ok && privateKey != "" {
		return nil
	}

	sk, pk, err := util.GenerateRSAKeyPairB64()
	if err != nil {
		return err
	}

	klog.Infof("inject %s into %s", PrivateKey, constants.KubeanConfigMapName)
	kubeanConfig.Data[PrivateKey] = sk
	_, err = clientset.CoreV1().ConfigMaps(util.GetCurrentNSOrDefault()).Update(context.Background(), kubeanConfig, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	kubeanPubkey := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.KubeanPubKeyConfigMapName,
			Namespace: util.GetCurrentNSOrDefault(),
		},
		Data: map[string]string{
			PublicKey: pk,
		},
	}
	klog.Infof("create %s", kubeanPubkey.Name)
	_, err = clientset.CoreV1().ConfigMaps(util.GetCurrentNSOrDefault()).Create(context.Background(), kubeanPubkey, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
