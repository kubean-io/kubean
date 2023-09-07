// Copyright 2023 Authors of kubean-io
// SPDX-License-Identifier: Apache-2.0

package clusterops

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/endpoints/responsewriter"

	clusteroperationv1alpha1 "github.com/kubean-io/kubean-api/apis/clusteroperation/v1alpha1"
	clusteroperationv1alpha1fake "github.com/kubean-io/kubean-api/generated/clusteroperation/clientset/versioned/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetfake "k8s.io/client-go/kubernetes/fake"

	"github.com/kubean-io/kubean/pkg/util"
)

func TestCreateHTTPSCAFiles(t *testing.T) {
	certsDir = strings.TrimRight(os.TempDir(), "/")
	if err := CreateHTTPSCAFiles(); err != nil {
		t.Fatal(err)
	}
	if err := CreateHTTPSCAFiles(); err != nil {
		t.Fatal(err)
	}
	if !util.IsExist(filepath.Join(certsDir, certFile)) {
		t.Fatal()
	}
	if !util.IsExist(filepath.Join(certsDir, certKey)) {
		t.Fatal()
	}
}

func TestPrepareWebHookHTTPSServer(t *testing.T) {
	server := prepareWebHookHTTPSServer(nil)
	if server == nil {
		t.Fatal()
	}
}

func TestUpdateClusterOperationWebhook(t *testing.T) {
	certsDir = strings.TrimRight(os.TempDir(), "/")
	ClusterOperationWebhook = "my-webhook-abc"
	CreateHTTPSCAFiles()
	fakeClientSet := clientsetfake.NewSimpleClientset()
	if err := UpdateClusterOperationWebhook(fakeClientSet); err != nil {
		t.Fatal()
	}
	os.Remove(filepath.Join(certsDir, certFile))
	CreateHTTPSCAFiles()
	if err := UpdateClusterOperationWebhook(fakeClientSet); err != nil {
		t.Fatal()
	}
	result, err := fakeClientSet.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(context.Background(), ClusterOperationWebhook, metav1.GetOptions{})
	if err != nil {
		t.Fatal()
	}
	if len(result.Webhooks) == 0 {
		t.Fatal()
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
