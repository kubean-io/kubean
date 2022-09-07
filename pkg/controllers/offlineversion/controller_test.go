package offlineversion

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clientsetfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kubeancomponentsversionv1alpha1 "kubean.io/api/apis/kubeancomponentsversion/v1alpha1"
	kubeanofflineversionv1alpha1 "kubean.io/api/apis/kubeanofflineversion/v1alpha1"
	"kubean.io/api/constants"
	kubeancomponentsversionv1alpha1fake "kubean.io/api/generated/kubeancomponentsversion/clientset/versioned/fake"
	kubeanofflineversionv1alpha1fake "kubean.io/api/generated/kubeanofflineversion/clientset/versioned/fake"
)

func newFakeClient() client.Client {
	sch := scheme.Scheme
	if err := kubeancomponentsversionv1alpha1.AddToScheme(sch); err != nil {
		panic(err)
	}
	if err := kubeanofflineversionv1alpha1.AddToScheme(sch); err != nil {
		panic(err)
	}
	client := fake.NewClientBuilder().WithScheme(sch).WithRuntimeObjects(&kubeancomponentsversionv1alpha1.KuBeanComponentsVersion{}).WithRuntimeObjects(&kubeanofflineversionv1alpha1.KuBeanOfflineVersion{}).Build()
	return client
}

func TestMergeOfflineVersion(t *testing.T) {
	controller := &Controller{
		Client:                     newFakeClient(),
		ClientSet:                  clientsetfake.NewSimpleClientset(),
		OfflineversionClientSet:    kubeanofflineversionv1alpha1fake.NewSimpleClientset(),
		ComponentsversionClientSet: kubeancomponentsversionv1alpha1fake.NewSimpleClientset(),
	}
	tests := []struct {
		name string
		args struct {
			OfflineVersion    kubeanofflineversionv1alpha1.KuBeanOfflineVersion
			ComponentsVersion kubeancomponentsversionv1alpha1.KuBeanComponentsVersion
		}
		updated bool
	}{
		{
			name: "nothing update",
			args: struct {
				OfflineVersion    kubeanofflineversionv1alpha1.KuBeanOfflineVersion
				ComponentsVersion kubeancomponentsversionv1alpha1.KuBeanComponentsVersion
			}{
				OfflineVersion: kubeanofflineversionv1alpha1.KuBeanOfflineVersion{
					Spec: kubeanofflineversionv1alpha1.Spec{},
				},
				ComponentsVersion: kubeancomponentsversionv1alpha1.KuBeanComponentsVersion{
					Status: kubeancomponentsversionv1alpha1.Status{
						Offline: kubeancomponentsversionv1alpha1.OfflineStatus{},
					},
				},
			},
			updated: false,
		},
		{
			name: "update software info",
			args: struct {
				OfflineVersion    kubeanofflineversionv1alpha1.KuBeanOfflineVersion
				ComponentsVersion kubeancomponentsversionv1alpha1.KuBeanComponentsVersion
			}{
				OfflineVersion: kubeanofflineversionv1alpha1.KuBeanOfflineVersion{
					Spec: kubeanofflineversionv1alpha1.Spec{
						Items: []*kubeanofflineversionv1alpha1.SoftwareInfo{
							{
								Name:         "etcd-1",
								VersionRange: []string{"1.1", "1.2"},
							},
						},
					},
				},
				ComponentsVersion: kubeancomponentsversionv1alpha1.KuBeanComponentsVersion{
					Status: kubeancomponentsversionv1alpha1.Status{
						Offline: kubeancomponentsversionv1alpha1.OfflineStatus{},
					},
				},
			},
			updated: true,
		},
		{
			name: "update software info",
			args: struct {
				OfflineVersion    kubeanofflineversionv1alpha1.KuBeanOfflineVersion
				ComponentsVersion kubeancomponentsversionv1alpha1.KuBeanComponentsVersion
			}{
				OfflineVersion: kubeanofflineversionv1alpha1.KuBeanOfflineVersion{
					Spec: kubeanofflineversionv1alpha1.Spec{
						Items: []*kubeanofflineversionv1alpha1.SoftwareInfo{
							{
								Name:         "etcd-1",
								VersionRange: []string{"1.1", "1.2"},
							},
						},
					},
				},
				ComponentsVersion: kubeancomponentsversionv1alpha1.KuBeanComponentsVersion{
					Status: kubeancomponentsversionv1alpha1.Status{
						Offline: kubeancomponentsversionv1alpha1.OfflineStatus{
							Items: []*kubeancomponentsversionv1alpha1.SoftwareInfoStatus{
								{
									Name:         "etcd-1",
									VersionRange: []string{"1.2", "1.3"},
								},
							},
						},
					},
				},
			},
			updated: true,
		},
		{
			name: "add software info but nothing updated",
			args: struct {
				OfflineVersion    kubeanofflineversionv1alpha1.KuBeanOfflineVersion
				ComponentsVersion kubeancomponentsversionv1alpha1.KuBeanComponentsVersion
			}{
				OfflineVersion: kubeanofflineversionv1alpha1.KuBeanOfflineVersion{
					Spec: kubeanofflineversionv1alpha1.Spec{
						Items: []*kubeanofflineversionv1alpha1.SoftwareInfo{
							{
								Name:         "etcd-1",
								VersionRange: []string{"1.1", "1.2"},
							},
						},
					},
				},
				ComponentsVersion: kubeancomponentsversionv1alpha1.KuBeanComponentsVersion{
					Status: kubeancomponentsversionv1alpha1.Status{
						Offline: kubeancomponentsversionv1alpha1.OfflineStatus{
							Items: []*kubeancomponentsversionv1alpha1.SoftwareInfoStatus{
								{
									Name:         "etcd-1",
									VersionRange: []string{"1.1", "1.2", "1.3"},
								},
							},
						},
					},
				},
			},
			updated: false,
		},
		{
			name: "update docker-ce info",
			args: struct {
				OfflineVersion    kubeanofflineversionv1alpha1.KuBeanOfflineVersion
				ComponentsVersion kubeancomponentsversionv1alpha1.KuBeanComponentsVersion
			}{
				OfflineVersion: kubeanofflineversionv1alpha1.KuBeanOfflineVersion{
					Spec: kubeanofflineversionv1alpha1.Spec{
						Docker: []*kubeanofflineversionv1alpha1.DockerInfo{
							{
								OS:           "redhat-7",
								VersionRange: []string{"20.01", "20.02"},
							},
						},
						Items: []*kubeanofflineversionv1alpha1.SoftwareInfo{
							{
								Name:         "etcd-1",
								VersionRange: []string{"1.1", "1.2"},
							},
						},
					},
				},
				ComponentsVersion: kubeancomponentsversionv1alpha1.KuBeanComponentsVersion{
					Status: kubeancomponentsversionv1alpha1.Status{
						Offline: kubeancomponentsversionv1alpha1.OfflineStatus{
							Docker: []*kubeancomponentsversionv1alpha1.DockerInfoStatus{
								{
									OS:           "redhat-8",
									VersionRange: []string{},
								},
							},
							Items: []*kubeancomponentsversionv1alpha1.SoftwareInfoStatus{
								{
									Name:         "etcd-1",
									VersionRange: []string{"1.1", "1.2"},
								},
							},
						},
					},
				},
			},
			updated: true,
		},
		{
			name: "nothing updated",
			args: struct {
				OfflineVersion    kubeanofflineversionv1alpha1.KuBeanOfflineVersion
				ComponentsVersion kubeancomponentsversionv1alpha1.KuBeanComponentsVersion
			}{
				OfflineVersion: kubeanofflineversionv1alpha1.KuBeanOfflineVersion{
					Spec: kubeanofflineversionv1alpha1.Spec{
						Docker: []*kubeanofflineversionv1alpha1.DockerInfo{
							{
								OS:           "redhat-7",
								VersionRange: []string{"20.01", "20.02"},
							},
						},
						Items: []*kubeanofflineversionv1alpha1.SoftwareInfo{
							{
								Name:         "etcd-1",
								VersionRange: []string{"1.1", "1.2"},
							},
						},
					},
				},
				ComponentsVersion: kubeancomponentsversionv1alpha1.KuBeanComponentsVersion{
					Status: kubeancomponentsversionv1alpha1.Status{
						Offline: kubeancomponentsversionv1alpha1.OfflineStatus{
							Docker: []*kubeancomponentsversionv1alpha1.DockerInfoStatus{
								{
									OS:           "redhat-7",
									VersionRange: []string{"20.02", "20.01"},
								},
								{
									OS:           "redhat-8",
									VersionRange: []string{"21.02", "21.01"},
								},
							},
							Items: []*kubeancomponentsversionv1alpha1.SoftwareInfoStatus{
								{
									Name:         "etcd-1",
									VersionRange: []string{"1.2", "1.1"},
								},
							},
						},
					},
				},
			},
			updated: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if updated, _ := controller.MergeOfflineVersionStatus(&test.args.OfflineVersion, &test.args.ComponentsVersion); updated != test.updated {
				t.Fatal()
			}
		})
	}
}

func TestReconcile(t *testing.T) {
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "merge successfully",
			args: func() bool {
				controller := &Controller{
					Client:                     newFakeClient(),
					ClientSet:                  clientsetfake.NewSimpleClientset(),
					OfflineversionClientSet:    kubeanofflineversionv1alpha1fake.NewSimpleClientset(),
					ComponentsversionClientSet: kubeancomponentsversionv1alpha1fake.NewSimpleClientset(),
				}
				offlineVersionData := kubeanofflineversionv1alpha1.KuBeanOfflineVersion{
					TypeMeta: metav1.TypeMeta{
						Kind:       "KuBeanOfflineVersion",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "offlineversion-1",
					},
					Spec: kubeanofflineversionv1alpha1.Spec{
						Docker: []*kubeanofflineversionv1alpha1.DockerInfo{
							{
								OS:           "redhat-7",
								VersionRange: []string{"20.1", "20.2"},
							},
						},
					},
				}

				globalComponentsVersion := kubeancomponentsversionv1alpha1.KuBeanComponentsVersion{
					TypeMeta: metav1.TypeMeta{
						Kind:       "KuBeanComponentsVersion",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: constants.ComponentsversionGlobalName,
					},
				}

				controller.Create(context.Background(), &offlineVersionData)
				controller.OfflineversionClientSet.KubeanV1alpha1().KuBeanOfflineVersions().Create(context.Background(), &offlineVersionData, metav1.CreateOptions{})
				controller.ComponentsversionClientSet.KubeanV1alpha1().KuBeanComponentsVersions().Create(context.Background(), &globalComponentsVersion, metav1.CreateOptions{})

				result, err := controller.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: types.NamespacedName{Name: offlineVersionData.Name}})
				newGlobalComponentsVersion, _ := controller.ComponentsversionClientSet.KubeanV1alpha1().KuBeanComponentsVersions().Get(context.Background(), constants.ComponentsversionGlobalName, metav1.GetOptions{})
				return err == nil && result.RequeueAfter == Loop && len(newGlobalComponentsVersion.Status.Offline.Docker) == 1 && len(newGlobalComponentsVersion.Status.Offline.Docker[0].VersionRange) == 2
			},
			want: true,
		},
		{
			name: "ComponentsversionGlobal not exist",
			args: func() bool {
				controller := &Controller{
					Client:                     newFakeClient(),
					ClientSet:                  clientsetfake.NewSimpleClientset(),
					OfflineversionClientSet:    kubeanofflineversionv1alpha1fake.NewSimpleClientset(),
					ComponentsversionClientSet: kubeancomponentsversionv1alpha1fake.NewSimpleClientset(),
				}
				offlineVersionData := kubeanofflineversionv1alpha1.KuBeanOfflineVersion{
					TypeMeta: metav1.TypeMeta{
						Kind:       "KuBeanOfflineVersion",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "offlineversion-1",
					},
					Spec: kubeanofflineversionv1alpha1.Spec{
						Docker: []*kubeanofflineversionv1alpha1.DockerInfo{
							{
								OS:           "redhat-7",
								VersionRange: []string{"20.1", "20.2"},
							},
						},
					},
				}

				controller.Create(context.Background(), &offlineVersionData)
				controller.OfflineversionClientSet.KubeanV1alpha1().KuBeanOfflineVersions().Create(context.Background(), &offlineVersionData, metav1.CreateOptions{})

				result, _ := controller.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: types.NamespacedName{Name: offlineVersionData.Name}})
				return result.RequeueAfter == Loop
			},
			want: true,
		},
		{
			name: "offlineVersion not found",
			args: func() bool {
				controller := &Controller{
					Client:                     newFakeClient(),
					ClientSet:                  clientsetfake.NewSimpleClientset(),
					OfflineversionClientSet:    kubeanofflineversionv1alpha1fake.NewSimpleClientset(),
					ComponentsversionClientSet: kubeancomponentsversionv1alpha1fake.NewSimpleClientset(),
				}
				result, _ := controller.Reconcile(context.Background(), controllerruntime.Request{NamespacedName: types.NamespacedName{Name: "offlineversion-1"}})
				return result.Requeue == false && result.RequeueAfter == 0
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
