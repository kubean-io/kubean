// Copyright 2023 Authors of kubean-io
// SPDX-License-Identifier: Apache-2.0

package clusterops

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	clusteroperationv1alpha1 "github.com/kubean-io/kubean-api/apis/clusteroperation/v1alpha1"
	"github.com/kubean-io/kubean-api/constants"
	clusterOperationClientSet "github.com/kubean-io/kubean-api/generated/clusteroperation/clientset/versioned"

	admissionv1 "k8s.io/api/admission/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/kubean-io/kubean/pkg/util"
)

var (
	certsDir          = "/etc/webhook/certs"
	certKey           = "tls.key"
	certFile          = "tls.crt"
	Organization      = "kubean.io"
	DefaultEffectTime = 10 * 365 * 24 * time.Hour

	WebHookPath                = "/mutate-webhook"
	WebhookSVCNamespace, _     = os.LookupEnv("WEBHOOK_SERVICE_NAMESPACE")
	ClusterOperationWebhook, _ = os.LookupEnv("MUTATE_WEBHOOK_CONFIG")
	WebhookSVCName, _          = os.LookupEnv("WEBHOOK_SERVICE_NAME")
	FailurePolicy, _           = os.LookupEnv("WEBHOOK_FAILURE_POLICY")
	dnsNames                   = []string{
		WebhookSVCName,
		WebhookSVCName + "." + WebhookSVCNamespace,
		WebhookSVCName + "." + WebhookSVCNamespace + "." + "svc",
		WebhookSVCName + "." + WebhookSVCNamespace + "." + "svc.cluster.local",
	}
	commonName = WebhookSVCName + "." + WebhookSVCNamespace + "." + "svc"
)

func RunWithLeaseLock(ctx context.Context, client kubernetes.Interface, KubeanClusterOpsSet clusterOperationClientSet.Interface) error {
	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("webhook RunWithLeaseLock but error %v", r)
		}
	}()
	if err := CreateHTTPSCAFiles(); err != nil {
		klog.ErrorS(err, "CreateHTTPSCAFiles")
		return err
	}
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      "lease-for-kubean-webhook",
			Namespace: util.GetCurrentNSOrDefault(),
		},
		Client: client.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: util.GetCurrentRunningPodName(), // podName as ID
		},
	}
	var server *http.Server
	serverLock := &sync.Mutex{}
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   time.Second * 60,
		RenewDeadline:   time.Second * 30,
		RetryPeriod:     time.Second * 20,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				klog.Warningf("webhook OnStartedLeading on %s", util.GetCurrentRunningPodName())
				serverLock.Lock()
				if server != nil {
					server.Close()
				}
				server = prepareWebHookHTTPSServer(KubeanClusterOpsSet)
				serverLock.Unlock()
				if err := UpdateClusterOperationWebhook(client); err != nil {
					klog.ErrorS(err, "UpdateClusterOperationWebhook")
					return
				}
				StartWebHookHTTPSServer(server) // block
			},
			OnNewLeader: func(identity string) {
				klog.Warningf("webhook OnNewLeader on %s", identity)
				if identity != util.GetCurrentRunningPodName() { // other pod is new leader, then close https server
					serverLock.Lock()
					defer serverLock.Unlock()
					if server != nil {
						server.Close()
						server = nil
					}
				}
			},
			OnStoppedLeading: func() {
			},
		},
	})
	return nil
}

func CreateHTTPSCAFiles() error {
	if certsDir == "" {
		return errors.New("empty certsDir")
	}
	if util.IsExist(filepath.Join(certsDir, certFile)) && util.IsExist(filepath.Join(certsDir, certKey)) {
		// need not create CA files for https.
		return nil
	}
	if err := os.MkdirAll(certsDir, 0o666); err != nil {
		return err
	}
	serverCertPEM, serverPrivateKeyPEM, err := util.NewCertManager(
		[]string{Organization},
		DefaultEffectTime,
		dnsNames,
		commonName,
	).GenerateSelfSignedCerts()
	if err != nil {
		klog.ErrorS(err, "failed to generate certs")
		return err
	}
	err = util.WriteFile(filepath.Join(certsDir, certFile), serverCertPEM)
	if err != nil {
		klog.ErrorS(err, "failed to write tls.cert")
		return err
	}

	err = util.WriteFile(filepath.Join(certsDir, certKey), serverPrivateKeyPEM)
	if err != nil {
		klog.ErrorS(err, "failed to write tls.key")
		return err
	}
	return nil
}

type PingHandler struct{}

func (p PingHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK) // call WriteHandler before Write
	writer.Write([]byte("pong"))
}

type AdmissionReviewHandler struct {
	KubeanClusterOpsSet clusterOperationClientSet.Interface
}

func (handler AdmissionReviewHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	klog.Warning("receive webhook request")
	admissionReviewReq := admissionv1.AdmissionReview{}
	if err := json.NewDecoder(request.Body).Decode(&admissionReviewReq); err != nil {
		klog.ErrorS(err, "parse http body to AdmissionReview")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(fmt.Sprint(err, "parse http body to AdmissionReview")))
		return
	}
	if admissionReviewReq.Request == nil || len(admissionReviewReq.Request.Object.Raw) == 0 { // for validate create ,so Object.Raw is not empty
		klog.Error("parse http body to AdmissionReview but no object")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("parse http body to AdmissionReview but no object"))
		return
	}
	clusterOperation := clusteroperationv1alpha1.ClusterOperation{}
	if err := json.Unmarshal(admissionReviewReq.Request.Object.Raw, &clusterOperation); err != nil {
		klog.ErrorS(err, "parse AdmissionReview.Object.Raw in ClusterOperation but failed")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(fmt.Sprint(err, "parse AdmissionReview.Object.Raw in ClusterOperation but failed")))
		return
	}
	if clusterOperation.Spec.Cluster == "" {
		klog.Error("parse AdmissionReview.Object.Raw in ClusterOperation but spec.Cluster is empty")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("parse AdmissionReview.Object.Raw in ClusterOperation but spec.Cluster is empty"))
		return
	}
	klog.Warningf("receive webhook request for clusterOperation %s", clusterOperation.Name)
	requirement, err := labels.NewRequirement(constants.KubeanClusterHasCompleted, selection.DoesNotExist, []string{}) // only when the operation has succeed or failed , then has this label
	if err != nil {                                                                                                    // todo
		klog.Error(err)
		return
	}
	selector := labels.NewSelector()
	selector.Add(*requirement)
	opsList, err := handler.KubeanClusterOpsSet.KubeanV1alpha1().ClusterOperations().List(context.Background(), metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		klog.ErrorS(err, "fetch ClusterOperations but failed")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(fmt.Sprint(err, "fetch ClusterOperations but failed")))
		return
	}
	admissionReviewResponse := admissionv1.AdmissionReview{
		TypeMeta: admissionReviewReq.TypeMeta,
		Response: &admissionv1.AdmissionResponse{
			UID: admissionReviewReq.Request.UID,
		},
	}
	// allow default
	admissionReviewResponse.Response.Allowed = true

	for _, ops := range opsList.Items {
		if ops.Status.Status == clusteroperationv1alpha1.FailedStatus || ops.Status.Status == clusteroperationv1alpha1.SucceededStatus {
			continue // ignore
		}
		if ops.Name != clusterOperation.Name && ops.Spec.Cluster == clusterOperation.Spec.Cluster &&
			(ops.Status.Status == "" || ops.Status.Status == clusteroperationv1alpha1.RunningStatus) { // belongs to the same cluster and still running
			// not allow
			admissionReviewResponse.Response.Allowed = false
			admissionReviewResponse.Response.Result = &metav1.Status{
				Message: fmt.Sprintf("clusterOperation %s has not completed which belongs to Cluster %s", ops.Name, clusterOperation.Spec.Cluster),
				Reason:  metav1.StatusReasonNotAcceptable,
				Code:    http.StatusNotAcceptable,
			}
			break
		}
	}
	httpResult, _ := json.Marshal(admissionReviewResponse)
	writer.WriteHeader(http.StatusOK)
	writer.Write(httpResult)
}

func prepareWebHookHTTPSServer(KubeanClusterOpsSet clusterOperationClientSet.Interface) *http.Server {
	mux := http.NewServeMux()
	mux.Handle(WebHookPath, AdmissionReviewHandler{KubeanClusterOpsSet: KubeanClusterOpsSet})
	mux.Handle("/ping", PingHandler{})
	server := &http.Server{
		Addr:    ":10443",
		Handler: mux,
	}
	return server
}

func StartWebHookHTTPSServer(server *http.Server) {
	certPath := filepath.Join(certsDir, certFile)
	keyPath := filepath.Join(certsDir, certKey)
	if err := server.ListenAndServeTLS(certPath, keyPath); err != nil {
		klog.ErrorS(err, "start https server for webhook but failed")
	}
}

func UpdateClusterOperationWebhook(clientSet kubernetes.Interface) error {
	if ClusterOperationWebhook == "" {
		return errors.New("ClusterOperationWebhook empty")
	}
	caCertData, err := os.ReadFile(filepath.Join(certsDir, certFile))
	if err != nil {
		return err
	}
	m, err := clientSet.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(context.Background(), ClusterOperationWebhook, metav1.GetOptions{})
	if err == nil && len(m.Webhooks) > 0 && string(m.Webhooks[0].ClientConfig.CABundle) == string(caCertData) {
		// need not update mutating-webhook
		return nil
	}
	mutateConfig := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: ClusterOperationWebhook,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name: Organization + ".mutate-webhook",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					CABundle: caCertData, // CA bundle created earlier
					Service: &admissionregistrationv1.ServiceReference{
						Name:      WebhookSVCName,
						Namespace: WebhookSVCNamespace,
						Path:      &WebHookPath,
						Port: func() *int32 {
							httpsPort := int32(443)
							return &httpsPort
						}(),
					},
				},
				Rules: []admissionregistrationv1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1.OperationType{
							admissionregistrationv1.Create, // only for create , not for update or delete
						},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"kubean.io"},
							APIVersions: []string{"v1alpha1"},
							Resources:   []string{"clusteroperations"},
						},
					},
				},
				FailurePolicy: func() *admissionregistrationv1.FailurePolicyType {
					policy := admissionregistrationv1.FailurePolicyType(FailurePolicy)
					return &policy
				}(),
				TimeoutSeconds: func() *int32 {
					timeout := int32(10)
					return &timeout
				}(),
				AdmissionReviewVersions: []string{"v1", "v1beta1"},
				SideEffects: func() *admissionregistrationv1.SideEffectClass {
					se := admissionregistrationv1.SideEffectClassNone
					return &se
				}(),
			},
		},
	}
	if err != nil && apierrors.IsNotFound(err) { // create
		if _, err := clientSet.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(context.Background(), mutateConfig, metav1.CreateOptions{}); err != nil {
			klog.Error(err)
			return err
		}
		return nil
	}
	mutateConfig.ResourceVersion = m.ResourceVersion // update
	if _, err := clientSet.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(context.Background(), mutateConfig, metav1.UpdateOptions{}); err != nil {
		klog.Error(err)
		return err
	}
	return nil
}
