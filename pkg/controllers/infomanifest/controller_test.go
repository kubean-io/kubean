package infomanifest

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kubeaninfomanifestv1alpha1 "kubean.io/api/apis/kubeaninfomanifest/v1alpha1"
	kubeanofflineversionv1alpha1 "kubean.io/api/apis/kubeanofflineversion/v1alpha1"
	"kubean.io/api/constants"
	kubeaninfomanifestv1alpha1fake "kubean.io/api/generated/kubeaninfomanifest/clientset/versioned/fake"
)

func newFakeClient() client.Client {
	sch := scheme.Scheme
	if err := kubeaninfomanifestv1alpha1.AddToScheme(sch); err != nil {
		panic(err)
	}
	if err := kubeanofflineversionv1alpha1.AddToScheme(sch); err != nil {
		panic(err)
	}
	client := fake.NewClientBuilder().WithScheme(sch).WithRuntimeObjects(&kubeaninfomanifestv1alpha1.KubeanInfoManifest{}).WithRuntimeObjects(&kubeanofflineversionv1alpha1.KuBeanOfflineVersion{}).Build()
	return client
}

func Test_FetchLatestInfoManifest(t *testing.T) {
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "return empty",
			args: func() bool {
				controller := &Controller{
					Client:                newFakeClient(),
					ClientSet:             clientsetfake.NewSimpleClientset(),
					InfoManifestClientSet: kubeaninfomanifestv1alpha1fake.NewSimpleClientset(),
				}
				_, err := controller.FetchLatestInfoManifest()
				return err != nil && strings.Contains(err.Error(), "not found")
			},
			want: true,
		},
		{
			name: "return empty exclude the global-infomanifest",
			args: func() bool {
				controller := &Controller{
					Client:                newFakeClient(),
					ClientSet:             clientsetfake.NewSimpleClientset(),
					InfoManifestClientSet: kubeaninfomanifestv1alpha1fake.NewSimpleClientset(),
				}
				controller.InfoManifestClientSet.KubeanV1alpha1().KubeanInfoManifests().Create(context.Background(), &kubeaninfomanifestv1alpha1.KubeanInfoManifest{
					TypeMeta: metav1.TypeMeta{
						Kind:       "KubeanInfoManifest",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:              constants.InfoManifestGlobal,
						Labels:            map[string]string{OriginLabel: ""},
						CreationTimestamp: metav1.NewTime(time.Now()),
					},
					Spec: kubeaninfomanifestv1alpha1.Spec{},
				}, metav1.CreateOptions{})
				_, err := controller.FetchLatestInfoManifest()
				return err != nil && strings.Contains(err.Error(), "not found")
			},
			want: true,
		},
		{
			name: "return the latest infomanifest",
			args: func() bool {
				controller := &Controller{
					Client:                newFakeClient(),
					ClientSet:             clientsetfake.NewSimpleClientset(),
					InfoManifestClientSet: kubeaninfomanifestv1alpha1fake.NewSimpleClientset(),
				}
				now := time.Now()
				controller.InfoManifestClientSet.KubeanV1alpha1().KubeanInfoManifests().Create(context.Background(), &kubeaninfomanifestv1alpha1.KubeanInfoManifest{
					TypeMeta: metav1.TypeMeta{
						Kind:       "KubeanInfoManifest",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:              "v1",
						CreationTimestamp: metav1.NewTime(now.Add(time.Second * 100)),
					},
					Spec: kubeaninfomanifestv1alpha1.Spec{},
				}, metav1.CreateOptions{})
				controller.InfoManifestClientSet.KubeanV1alpha1().KubeanInfoManifests().Create(context.Background(), &kubeaninfomanifestv1alpha1.KubeanInfoManifest{
					TypeMeta: metav1.TypeMeta{
						Kind:       "KubeanInfoManifest",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:              "v2",
						CreationTimestamp: metav1.NewTime(now.Add(time.Second * 10000)),
					},
					Spec: kubeaninfomanifestv1alpha1.Spec{},
				}, metav1.CreateOptions{})
				controller.InfoManifestClientSet.KubeanV1alpha1().KubeanInfoManifests().Create(context.Background(), &kubeaninfomanifestv1alpha1.KubeanInfoManifest{
					TypeMeta: metav1.TypeMeta{
						Kind:       "KubeanInfoManifest",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:              constants.InfoManifestGlobal,
						CreationTimestamp: metav1.NewTime(now.Add(time.Second * 10000000)),
					},
					Spec: kubeaninfomanifestv1alpha1.Spec{},
				}, metav1.CreateOptions{})
				controller.InfoManifestClientSet.KubeanV1alpha1().KubeanInfoManifests().Create(context.Background(), &kubeaninfomanifestv1alpha1.KubeanInfoManifest{
					TypeMeta: metav1.TypeMeta{
						Kind:       "KubeanInfoManifest",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:              "v3",
						CreationTimestamp: metav1.NewTime(now.Add(time.Second * 1000)),
					},
					Spec: kubeaninfomanifestv1alpha1.Spec{},
				}, metav1.CreateOptions{})
				result, err := controller.FetchLatestInfoManifest()
				return err == nil && result.Name == "v2"
			},
			want: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.args() != test.want {
				t.Fatal()
			}
		})
	}
}

func Test_ParseConfigMapToLocalService(t *testing.T) {
	controller := &Controller{
		Client:                newFakeClient(),
		ClientSet:             clientsetfake.NewSimpleClientset(),
		InfoManifestClientSet: kubeaninfomanifestv1alpha1fake.NewSimpleClientset(),
	}
	tests := []struct {
		name string
		arg  *corev1.ConfigMap
		want *kubeaninfomanifestv1alpha1.LocalService
	}{
		{
			name: "zero data",
			arg:  &corev1.ConfigMap{},
			want: &kubeaninfomanifestv1alpha1.LocalService{},
		},
		{
			name: "empty string",
			arg:  &corev1.ConfigMap{Data: map[string]string{"localService": ""}},
			want: &kubeaninfomanifestv1alpha1.LocalService{},
		},
		{
			name: "good string data",
			arg:  &corev1.ConfigMap{Data: map[string]string{"localService": "      registryRepo: 'temp-registry.daocloud.io:5000'\n      filesRepo: 'http://temp-registry.daocloud.io:9000'\n      yumRepo:\n        - 'http://temp-registry.daocloud.io:9000/kubean/centos-iso/\\$releasever/os/\\$basearch'\n        - 'http://temp-registry.daocloud.io:9000/centos/\\$releasever/os/\\$basearch'\n      hostsMap:\n        - domain: temp-registry.daocloud.io\n          address: 'a.b.c.d'"}},
			want: &kubeaninfomanifestv1alpha1.LocalService{
				RegistryRepo: "temp-registry.daocloud.io:5000",
				FilesRepo:    "http://temp-registry.daocloud.io:9000",
				YumRepo:      []string{"http://temp-registry.daocloud.io:9000/kubean/centos-iso/\\$releasever/os/\\$basearch", "http://temp-registry.daocloud.io:9000/centos/\\$releasever/os/\\$basearch"},
				HostsMap: []*kubeaninfomanifestv1alpha1.HostsMap{
					{Domain: "temp-registry.daocloud.io", Address: "a.b.c.d"},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, _ := controller.ParseConfigMapToLocalService(test.arg)
			if !reflect.DeepEqual(result, test.want) {
				t.Fatal()
			}
		})
	}
}

func Test_EnsureGlobalInfoManifestBeingLatest(t *testing.T) {
	controller := &Controller{
		Client:                newFakeClient(),
		ClientSet:             clientsetfake.NewSimpleClientset(),
		InfoManifestClientSet: kubeaninfomanifestv1alpha1fake.NewSimpleClientset(),
	}
	tests := []struct {
		name               string
		latestInfoManifest func() *kubeaninfomanifestv1alpha1.KubeanInfoManifest
		want               *kubeaninfomanifestv1alpha1.KubeanInfoManifest
	}{
		{
			name: "not existing global InfoManifest",
			latestInfoManifest: func() *kubeaninfomanifestv1alpha1.KubeanInfoManifest {
				return &kubeaninfomanifestv1alpha1.KubeanInfoManifest{
					TypeMeta: metav1.TypeMeta{
						Kind:       "KubeanInfoManifest",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "v1",
					},
					Spec: kubeaninfomanifestv1alpha1.Spec{
						Components: []*kubeaninfomanifestv1alpha1.SoftwareInfo{{Name: "etcd1", VersionRange: []string{"1"}}},
					},
				}
			},
			want: &kubeaninfomanifestv1alpha1.KubeanInfoManifest{
				TypeMeta: metav1.TypeMeta{
					Kind:       "KubeanInfoManifest",
					APIVersion: "kubean.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:   constants.InfoManifestGlobal,
					Labels: map[string]string{OriginLabel: "v1"},
				},
				Spec: kubeaninfomanifestv1alpha1.Spec{
					Components: []*kubeaninfomanifestv1alpha1.SoftwareInfo{{Name: "etcd1", VersionRange: []string{"1"}}},
				},
			},
		},
		{
			name: "already existing global InfoManifest",
			latestInfoManifest: func() *kubeaninfomanifestv1alpha1.KubeanInfoManifest {
				return &kubeaninfomanifestv1alpha1.KubeanInfoManifest{
					TypeMeta: metav1.TypeMeta{
						Kind:       "KubeanInfoManifest",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "v2",
					},
					Spec: kubeaninfomanifestv1alpha1.Spec{
						Components: []*kubeaninfomanifestv1alpha1.SoftwareInfo{{Name: "etcd2", VersionRange: []string{"2"}}},
					},
				}
			},
			want: &kubeaninfomanifestv1alpha1.KubeanInfoManifest{
				TypeMeta: metav1.TypeMeta{
					Kind:       "KubeanInfoManifest",
					APIVersion: "kubean.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:   constants.InfoManifestGlobal,
					Labels: map[string]string{OriginLabel: "v2"},
				},
				Spec: kubeaninfomanifestv1alpha1.Spec{
					Components: []*kubeaninfomanifestv1alpha1.SoftwareInfo{{Name: "etcd2", VersionRange: []string{"2"}}},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			global, err := controller.EnsureGlobalInfoManifestBeingLatest(test.latestInfoManifest())
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(global, test.want) {
				t.Fatal()
			}
		})
	}
}

func Test_UpdateGlobalLocalService1(t *testing.T) {
	controller := &Controller{
		Client:                newFakeClient(),
		ClientSet:             clientsetfake.NewSimpleClientset(),
		InfoManifestClientSet: kubeaninfomanifestv1alpha1fake.NewSimpleClientset(),
	}
	tests := []struct {
		name string
		arg  func()
		want kubeaninfomanifestv1alpha1.LocalService
	}{
		{
			name: "global not have localService before",
			arg: func() {
				global := &kubeaninfomanifestv1alpha1.KubeanInfoManifest{
					TypeMeta: metav1.TypeMeta{
						Kind:       "KubeanInfoManifest",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   constants.InfoManifestGlobal,
						Labels: map[string]string{OriginLabel: "v2"},
					},
					Spec: kubeaninfomanifestv1alpha1.Spec{
						Components: []*kubeaninfomanifestv1alpha1.SoftwareInfo{{Name: "etcd2", VersionRange: []string{"2"}}},
					},
				}
				configMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      LocalServiceConfigMap,
						Namespace: "default",
					},
					Data: map[string]string{"localService": "      registryRepo: 'temp-registry.daocloud.io:5000'\n      filesRepo: 'http://temp-registry.daocloud.io:9000'\n      yumRepo:\n        - 'http://temp-registry.daocloud.io:9000/kubean/centos-iso/\\$releasever/os/\\$basearch'\n        - 'http://temp-registry.daocloud.io:9000/centos/\\$releasever/os/\\$basearch'\n      hostsMap: \n        - domain: temp-registry.daocloud.io\n          address: 'a.b.c.d'"},
				}
				controller.ClientSet.CoreV1().ConfigMaps("default").Create(context.Background(), configMap, metav1.CreateOptions{})
				controller.Create(context.Background(), global)
				controller.InfoManifestClientSet.KubeanV1alpha1().KubeanInfoManifests().Create(context.Background(), global, metav1.CreateOptions{})
				controller.UpdateGlobalLocalService()
			},
			want: kubeaninfomanifestv1alpha1.LocalService{
				RegistryRepo: "temp-registry.daocloud.io:5000",
				FilesRepo:    "http://temp-registry.daocloud.io:9000",
				YumRepo:      []string{"http://temp-registry.daocloud.io:9000/kubean/centos-iso/\\$releasever/os/\\$basearch", "http://temp-registry.daocloud.io:9000/centos/\\$releasever/os/\\$basearch"},
				HostsMap: []*kubeaninfomanifestv1alpha1.HostsMap{
					{
						Domain:  "temp-registry.daocloud.io",
						Address: "a.b.c.d",
					},
				},
			},
		},
		{
			name: "global already have localService and update it",
			arg: func() {
				global := &kubeaninfomanifestv1alpha1.KubeanInfoManifest{
					TypeMeta: metav1.TypeMeta{
						Kind:       "KubeanInfoManifest",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   constants.InfoManifestGlobal,
						Labels: map[string]string{OriginLabel: "v2"},
					},
					Spec: kubeaninfomanifestv1alpha1.Spec{
						Components: []*kubeaninfomanifestv1alpha1.SoftwareInfo{{Name: "etcd2", VersionRange: []string{"2"}}},
					},
				}
				configMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      LocalServiceConfigMap,
						Namespace: "default",
					},
					Data: map[string]string{"localService": "      registryRepo: 'temp-registry.daocloud.io:5000'\n      filesRepo: 'http://temp-registry.daocloud.io:9000'\n      yumRepo:\n        - 'http://temp-registry.daocloud.io:9000/kubean/centos-iso/\\$releasever/os/\\$basearch'\n        - 'http://temp-registry.daocloud.io:9000/centos/\\$releasever/os/\\$basearch'\n      hostsMap: \n        - domain: temp-registry.daocloud.io\n          address: 'a.b.c.d1'"},
				}
				controller.ClientSet.CoreV1().ConfigMaps("default").Update(context.Background(), configMap, metav1.UpdateOptions{})
				controller.Create(context.Background(), global)
				controller.InfoManifestClientSet.KubeanV1alpha1().KubeanInfoManifests().Create(context.Background(), global, metav1.CreateOptions{})
				controller.UpdateGlobalLocalService()
			},
			want: kubeaninfomanifestv1alpha1.LocalService{
				RegistryRepo: "temp-registry.daocloud.io:5000",
				FilesRepo:    "http://temp-registry.daocloud.io:9000",
				YumRepo:      []string{"http://temp-registry.daocloud.io:9000/kubean/centos-iso/\\$releasever/os/\\$basearch", "http://temp-registry.daocloud.io:9000/centos/\\$releasever/os/\\$basearch"},
				HostsMap: []*kubeaninfomanifestv1alpha1.HostsMap{
					{
						Domain:  "temp-registry.daocloud.io",
						Address: "a.b.c.d1",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.arg()
			global, err := controller.FetchGlobalInfoManifest()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(&global.Spec.LocalService, &test.want) {
				t.Fatal()
			}
		})
	}
}

func Test_UpdateLocalAvailableImage(t *testing.T) {
	controller := &Controller{
		Client:                newFakeClient(),
		ClientSet:             clientsetfake.NewSimpleClientset(),
		InfoManifestClientSet: kubeaninfomanifestv1alpha1fake.NewSimpleClientset(),
	}
	tests := []struct {
		name string
		arg  func()
		want string
	}{
		{
			name: "update local kubespray image with ghcr.io",
			arg: func() {
				global := &kubeaninfomanifestv1alpha1.KubeanInfoManifest{
					TypeMeta: metav1.TypeMeta{
						Kind:       "KubeanInfoManifest",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   constants.InfoManifestGlobal,
						Labels: map[string]string{OriginLabel: "v2"},
					},
					Spec: kubeaninfomanifestv1alpha1.Spec{
						KubeanVersion: "123",
					},
				}
				controller.Create(context.Background(), global)
				controller.InfoManifestClientSet.KubeanV1alpha1().KubeanInfoManifests().Create(context.Background(), global, metav1.CreateOptions{})
				controller.UpdateLocalAvailableImage()
			},
			want: "ghcr.m.daocloud.io/kubean-io/spray-job:123",
		},
		{
			name: "update local kubespray image with local registry",
			arg: func() {
				global := &kubeaninfomanifestv1alpha1.KubeanInfoManifest{
					TypeMeta: metav1.TypeMeta{
						Kind:       "KubeanInfoManifest",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:   constants.InfoManifestGlobal,
						Labels: map[string]string{OriginLabel: "v2"},
					},
					Spec: kubeaninfomanifestv1alpha1.Spec{
						LocalService:  kubeaninfomanifestv1alpha1.LocalService{RegistryRepo: "abc.io"},
						KubeanVersion: "123456",
					},
				}
				controller.Update(context.Background(), global)
				controller.InfoManifestClientSet.KubeanV1alpha1().KubeanInfoManifests().Update(context.Background(), global, metav1.UpdateOptions{})
				controller.UpdateLocalAvailableImage()
			},
			want: "abc.io/kubean-io/spray-job:123456",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.arg()
			global, err := controller.FetchGlobalInfoManifest()
			if err != nil {
				t.Fatal(err)
			}
			if global.Status.LocalAvailable.KubesprayImage != test.want {
				t.Fatal()
			}
		})
	}
}
