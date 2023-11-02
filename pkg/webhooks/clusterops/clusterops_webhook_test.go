// Copyright 2023 Authors of kubean-io
// SPDX-License-Identifier: Apache-2.0

package clusterops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apiserver/pkg/endpoints/responsewriter"
	clientsetfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	k8stesting "k8s.io/client-go/testing"

	clusteroperationv1alpha1 "github.com/kubean-io/kubean-api/apis/clusteroperation/v1alpha1"
	clusteroperationv1alpha1fake "github.com/kubean-io/kubean-api/generated/clusteroperation/clientset/versioned/fake"

	"github.com/kubean-io/kubean/pkg/util"
)

func fetchTestingFake(obj interface{ RESTClient() rest.Interface }) *k8stesting.Fake {
	// https://stackoverflow.com/questions/69740891/mocking-errors-with-client-go-fake-client
	return reflect.Indirect(reflect.ValueOf(obj)).FieldByName("Fake").Interface().(*k8stesting.Fake)
}

func removeReactorFromTestingTake(obj interface{ RESTClient() rest.Interface }, verb, resource string) {
	if fakeObj := fetchTestingFake(obj); fakeObj != nil {
		newReactionChain := make([]k8stesting.Reactor, 0)
		fakeObj.Lock()
		defer fakeObj.Unlock()
		for i := range fakeObj.ReactionChain {
			reaction := fakeObj.ReactionChain[i]
			if simpleReaction, ok := reaction.(*k8stesting.SimpleReactor); ok && simpleReaction.Verb == verb && simpleReaction.Resource == resource {
				continue // ignore
			}
			newReactionChain = append(newReactionChain, reaction)
		}
		fakeObj.ReactionChain = newReactionChain
	}
}

func TestCreateHTTPSCAFiles(t *testing.T) {
	certsDir = strings.TrimRight(os.TempDir(), "/")
	if err := CreateHTTPSCAFilesFromSecret(&corev1.Secret{}); err != nil {
		t.Fatal(err)
	}
	if err := CreateHTTPSCAFilesFromSecret(&corev1.Secret{}); err != nil {
		t.Fatal(err)
	}
	if !util.IsExist(filepath.Join(certsDir, certFile)) {
		t.Fatal()
	}
	if !util.IsExist(filepath.Join(certsDir, certKey)) {
		t.Fatal()
	}

	mux := http.NewServeMux()
	mux.Handle("/ping", PingHandler{})
	server := &http.Server{
		Addr:    ":10443",
		Handler: mux,
	}
	go func() {
		time.Sleep(time.Second * 2)
		server.Close()
	}()
	err := errors.AggregateGoroutines(func() error {
		return StartWebHookHTTPSServer(server)
	})
	if err != nil && !strings.Contains(err.Error(), "tls: failed to find any PEM data") {
		// other err
		t.Fatal(err)
	}
}

func TestPrepareWebHookHTTPSServer(t *testing.T) {
	server := PrepareWebHookHTTPSServer(nil)
	if server == nil {
		t.Fatal()
	}
}

func TestCreateHTTPSCAFilesFromSecret(t *testing.T) {
	certsDir = strings.TrimRight(os.TempDir(), "/")
	removeCrtData := func() {
		if certsDir != "" {
			os.Remove(filepath.Join(certsDir, certKey))
			os.Remove(filepath.Join(certsDir, certFile))
		}
	}
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "empty certsDir path",
			args: func() bool {
				certsDir = ""
				secret := &corev1.Secret{Data: map[string][]byte{
					"crt": []byte(""),
					"key": []byte(""),
				}}
				err := CreateHTTPSCAFilesFromSecret(secret)
				return err != nil
			},
			want: true,
		},
		{
			name: "good case",
			args: func() bool {
				certsDir = strings.TrimRight(os.TempDir(), "/")
				removeCrtData()
				secret := &corev1.Secret{Data: map[string][]byte{
					"crt": []byte(""),
					"key": []byte(""),
				}}
				CreateHTTPSCAFilesFromSecret(secret)
				err := CreateHTTPSCAFilesFromSecret(secret)
				return err == nil && util.IsExist(filepath.Join(certsDir, certFile)) && util.IsExist(filepath.Join(certsDir, certKey))
			},
			want: true,
		},
		{
			name: "bad base64 crt data",
			args: func() bool {
				certsDir = strings.TrimRight(os.TempDir(), "/")
				removeCrtData()
				secret := &corev1.Secret{
					Data: map[string][]byte{
						"crt": []byte("123"),
						"key": []byte(""),
					},
				}
				err := CreateHTTPSCAFilesFromSecret(secret)
				return err != nil
			},
			want: true,
		},
		{
			name: "bad base64 key data",
			args: func() bool {
				certsDir = strings.TrimRight(os.TempDir(), "/")
				removeCrtData()
				secret := &corev1.Secret{
					Data: map[string][]byte{
						"crt": []byte(""),
						"key": []byte("123"),
					},
				}
				err := CreateHTTPSCAFilesFromSecret(secret)
				return err != nil
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

func TestEnsureCASecretExist(t *testing.T) {
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "good case",
			args: func() bool {
				fakeClientSet := clientsetfake.NewSimpleClientset()
				err := EnsureCASecretExist(fakeClientSet)
				data, _ := fakeClientSet.CoreV1().Secrets(util.GetCurrentNSOrDefault()).Get(context.Background(), CAStoreSecret, metav1.GetOptions{})
				return err == nil && len(data.Data) == 2
			},
			want: true,
		},
		{
			name: "create secret unsuccessfully",
			args: func() bool {
				fakeClientSet := clientsetfake.NewSimpleClientset()
				fetchTestingFake(fakeClientSet.CoreV1()).PrependReactor("create", "secrets", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, fmt.Errorf("create secret but error")
				})
				defer removeReactorFromTestingTake(fakeClientSet.CoreV1(), "create", "secrets")
				err := EnsureCASecretExist(fakeClientSet)
				return err != nil && strings.Contains(err.Error(), "create secret but error")
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

func TestUpdateClusterOperationWebhook(t *testing.T) {
	certsDir = strings.TrimRight(os.TempDir(), "/")
	tests := []struct {
		name string
		arg  func() bool
		want bool
	}{
		{
			name: "create webhook",
			arg: func() bool {
				certsDir = strings.TrimRight(os.TempDir(), "/")
				ClusterOperationWebhook = "my-webhook-abc"
				CreateHTTPSCAFilesFromSecret(&corev1.Secret{})
				fakeClientSet := clientsetfake.NewSimpleClientset()
				EnsureCASecretExist(fakeClientSet)
				UpdateClusterOperationWebhook(fakeClientSet)
				os.Remove(filepath.Join(certsDir, certFile))
				CreateHTTPSCAFilesFromSecret(&corev1.Secret{})
				UpdateClusterOperationWebhook(fakeClientSet)
				result, err := fakeClientSet.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(context.Background(), ClusterOperationWebhook, metav1.GetOptions{})
				return len(result.Webhooks) != 0 && err == nil
			},
			want: true,
		},
		{
			name: "empty webhook name",
			arg: func() bool {
				certsDir = strings.TrimRight(os.TempDir(), "/")
				ClusterOperationWebhook = ""
				defer func() {
					ClusterOperationWebhook = "kubean-admission-webhook"
				}()
				fakeClientSet := clientsetfake.NewSimpleClientset()
				err := UpdateClusterOperationWebhook(fakeClientSet)
				return err != nil && err.Error() == "ClusterOperationWebhook empty"
			},
			want: true,
		},
		{
			name: "get secret unsuccessfully",
			arg: func() bool {
				certsDir = strings.TrimRight(os.TempDir(), "/")
				fakeClientSet := clientsetfake.NewSimpleClientset()
				EnsureCASecretExist(fakeClientSet)
				fetchTestingFake(fakeClientSet.CoreV1()).PrependReactor("get", "secrets", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, fmt.Errorf("this is error")
				})
				defer removeReactorFromTestingTake(fakeClientSet.CoreV1(), "get", "secrets")
				err := UpdateClusterOperationWebhook(fakeClientSet)
				return err != nil && err.Error() == "this is error"
			},
			want: true,
		},
		{
			name: "create webhook unsuccessfully",
			arg: func() bool {
				certsDir = strings.TrimRight(os.TempDir(), "/")
				fakeClientSet := clientsetfake.NewSimpleClientset()
				EnsureCASecretExist(fakeClientSet)
				fetchTestingFake(fakeClientSet.AdmissionregistrationV1()).PrependReactor("create", "validatingwebhookconfigurations", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, fmt.Errorf("create validatingwebhookconfigurations but error")
				})
				defer removeReactorFromTestingTake(fakeClientSet.AdmissionregistrationV1(), "get", "validatingwebhookconfigurations")
				err := UpdateClusterOperationWebhook(fakeClientSet)
				return err != nil && err.Error() == "create validatingwebhookconfigurations but error"
			},
			want: true,
		},
		{
			name: "update webhook unsuccessfully",
			arg: func() bool {
				certsDir = strings.TrimRight(os.TempDir(), "/")
				fakeClientSet := clientsetfake.NewSimpleClientset()
				EnsureCASecretExist(fakeClientSet)
				fetchTestingFake(fakeClientSet.AdmissionregistrationV1()).PrependReactor("update", "validatingwebhookconfigurations", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, fmt.Errorf("update validatingwebhookconfigurations but error")
				})
				defer removeReactorFromTestingTake(fakeClientSet.AdmissionregistrationV1(), "update", "validatingwebhookconfigurations")
				UpdateClusterOperationWebhook(fakeClientSet)
				fakeClientSet.CoreV1().Secrets(util.GetCurrentNSOrDefault()).Delete(context.Background(), CAStoreSecret, metav1.DeleteOptions{})
				EnsureCASecretExist(fakeClientSet)
				err := UpdateClusterOperationWebhook(fakeClientSet)
				return err != nil && err.Error() == "update validatingwebhookconfigurations but error"
			},
			want: true,
		},
		{
			name: "unnecessarily update webhook when secret is the same",
			arg: func() bool {
				certsDir = strings.TrimRight(os.TempDir(), "/")
				fakeClientSet := clientsetfake.NewSimpleClientset()
				EnsureCASecretExist(fakeClientSet)
				fetchTestingFake(fakeClientSet.AdmissionregistrationV1()).PrependReactor("update", "validatingwebhookconfigurations", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, fmt.Errorf("update validatingwebhookconfigurations but error")
				})
				defer removeReactorFromTestingTake(fakeClientSet.AdmissionregistrationV1(), "update", "validatingwebhookconfigurations")
				UpdateClusterOperationWebhook(fakeClientSet)
				err := UpdateClusterOperationWebhook(fakeClientSet)
				return err == nil
			},
			want: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.arg() != test.want {
				t.Fatal()
			}
		})
	}
}

type FakeResponseWriter struct {
	code   int
	result string
	responsewriter.FakeResponseWriter
}

func (fw *FakeResponseWriter) WriteHeader(code int) {
	fw.code = code
}

func (fw *FakeResponseWriter) Write(bs []byte) (int, error) {
	fw.result = string(bs)
	return len(bs), nil
}

func TestAdmissionReviewHandlerHttp(t *testing.T) {
	clusterOperationClientSet := clusteroperationv1alpha1fake.NewSimpleClientset()
	handler := AdmissionReviewHandler{clusterOperationClientSet}
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "PingHandler",
			args: func() bool {
				handler := PingHandler{}
				response := &FakeResponseWriter{}
				request, _ := http.NewRequest("", "", strings.NewReader("{}"))
				handler.ServeHTTP(response, request)
				return response.code == http.StatusOK && response.result == "pong"
			},
			want: true,
		},
		{
			name: "bad body parse",
			args: func() bool {
				response := &FakeResponseWriter{}
				request, _ := http.NewRequest("", "", strings.NewReader("{"))
				handler.ServeHTTP(response, request)
				return response.code == http.StatusBadRequest
			},
			want: true,
		},
		{
			name: "empty object",
			args: func() bool {
				response := &FakeResponseWriter{}
				request, _ := http.NewRequest("", "", strings.NewReader("{}"))
				handler.ServeHTTP(response, request)
				return response.code == http.StatusBadRequest
			},
			want: true,
		},
		{
			name: "bad clusterOperation json body",
			args: func() bool {
				response := &FakeResponseWriter{}
				raw, _ := json.Marshal([]string{"1", "2"})
				admissionReview := admissionv1.AdmissionReview{Request: &admissionv1.AdmissionRequest{Object: runtime.RawExtension{Raw: raw}}}
				admissionReviewBytes, _ := json.Marshal(admissionReview)
				request, _ := http.NewRequest("", "", bytes.NewReader(admissionReviewBytes))
				handler.ServeHTTP(response, request)
				return response.code == http.StatusBadRequest && strings.Contains(response.result, "parse AdmissionReview.Object.Raw in ClusterOperation but failed")
			},
			want: true,
		},
		{
			name: "empty spec.Cluster",
			args: func() bool {
				response := &FakeResponseWriter{}
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "my_kubean_ops_cluster",
					},
				}
				raw, _ := json.Marshal(clusterOps)
				admissionReview := admissionv1.AdmissionReview{Request: &admissionv1.AdmissionRequest{Object: runtime.RawExtension{Raw: raw}}}
				admissionReviewBytes, _ := json.Marshal(admissionReview)
				request, _ := http.NewRequest("", "", bytes.NewReader(admissionReviewBytes))
				handler.ServeHTTP(response, request)
				return response.code == http.StatusBadRequest && strings.Contains(response.result, "spec.Cluster is empty")
			},
			want: true,
		},
		{
			name: "allow",
			args: func() bool {
				response := &FakeResponseWriter{}
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "my_kubean_ops_cluster",
					},
					Spec: clusteroperationv1alpha1.Spec{
						Cluster: "abc_cluster",
					},
				}
				raw, _ := json.Marshal(clusterOps)
				admissionReview := admissionv1.AdmissionReview{Request: &admissionv1.AdmissionRequest{Object: runtime.RawExtension{Raw: raw}}}
				admissionReviewBytes, _ := json.Marshal(admissionReview)
				request, _ := http.NewRequest("", "", bytes.NewReader(admissionReviewBytes))
				handler.ServeHTTP(response, request)
				admissionReviewResponse := &admissionv1.AdmissionReview{}
				json.Unmarshal([]byte(response.result), admissionReviewResponse)
				return response.code == http.StatusOK && admissionReviewResponse.Response.Allowed
			},
			want: true,
		},
		{
			name: "list clusterOperations unsuccessfully",
			args: func() bool {
				response := &FakeResponseWriter{}
				clusterOps := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "my_kubean_ops_cluster",
					},
					Spec: clusteroperationv1alpha1.Spec{
						Cluster: "abc_cluster",
					},
				}
				raw, _ := json.Marshal(clusterOps)
				admissionReview := admissionv1.AdmissionReview{Request: &admissionv1.AdmissionRequest{Object: runtime.RawExtension{Raw: raw}}}
				admissionReviewBytes, _ := json.Marshal(admissionReview)
				request, _ := http.NewRequest("", "", bytes.NewReader(admissionReviewBytes))
				fetchTestingFake(clusterOperationClientSet.KubeanV1alpha1()).PrependReactor("list", "clusteroperations", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, fmt.Errorf("list clusteroperations but error")
				})
				defer removeReactorFromTestingTake(clusterOperationClientSet.KubeanV1alpha1(), "list", "clusteroperations")
				handler.ServeHTTP(response, request)
				admissionReviewResponse := &admissionv1.AdmissionReview{}
				json.Unmarshal([]byte(response.result), admissionReviewResponse)
				return response.code == http.StatusBadRequest && strings.Contains(response.result, "fetch ClusterOperations but failed")
			},
			want: true,
		},
		{
			name: "not allow",
			args: func() bool {
				response := &FakeResponseWriter{}
				clusterOps1 := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "my_kubean_ops_cluster_1",
					},
					Spec: clusteroperationv1alpha1.Spec{
						Cluster: "abc_cluster",
					},
				}
				clusterOperationClientSet.KubeanV1alpha1().ClusterOperations().Create(context.Background(), clusterOps1, metav1.CreateOptions{})
				clusterOps2 := &clusteroperationv1alpha1.ClusterOperation{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterOperation",
						APIVersion: "kubean.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "my_kubean_ops_cluster_2",
					},
					Spec: clusteroperationv1alpha1.Spec{
						Cluster: "abc_cluster",
					},
				}
				raw, _ := json.Marshal(clusterOps2)
				admissionReview := admissionv1.AdmissionReview{Request: &admissionv1.AdmissionRequest{Object: runtime.RawExtension{Raw: raw}}}
				admissionReviewBytes, _ := json.Marshal(admissionReview)
				request, _ := http.NewRequest("", "", bytes.NewReader(admissionReviewBytes))
				handler.ServeHTTP(response, request)
				admissionReviewResponse := &admissionv1.AdmissionReview{}
				json.Unmarshal([]byte(response.result), admissionReviewResponse)
				return response.code == http.StatusOK && !admissionReviewResponse.Response.Allowed && admissionReviewResponse.Response.Result.Reason == metav1.StatusReasonNotAcceptable
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

func TestWaitForCASecretExist(t *testing.T) {
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "good case",
			args: func() bool {
				fakeClientSet := clientsetfake.NewSimpleClientset()
				go func() {
					time.Sleep(time.Second)
					EnsureCASecretExist(fakeClientSet)
				}()
				data := WaitForCASecretExist(fakeClientSet)
				return data != nil
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

func TestCreateHTTPSCASecretWithLock(t *testing.T) {
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "good case",
			args: func() bool {
				fakeClientSet := clientsetfake.NewSimpleClientset()
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(time.Second * 2)
					cancel()
				}()
				err := CreateHTTPSCASecretWithLock(ctx, fakeClientSet)
				return err == nil
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
