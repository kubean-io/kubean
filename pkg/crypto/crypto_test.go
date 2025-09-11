package crypto

import (
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"

	"github.com/kubean-io/kubean-api/constants"
	"github.com/kubean-io/kubean/pkg/util"
)

func TestInitConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		existingCM   *corev1.ConfigMap
		expectError  bool
		expectCreate bool
		expectUpdate bool
		getError     error
		updateError  error
		createError  error
	}{
		{
			name: "private key already exists",
			existingCM: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      constants.KubeanConfigMapName,
					Namespace: util.GetCurrentNSOrDefault(),
				},
				Data: map[string]string{
					PrivateKey: "existing-private-key",
				},
			},
			expectError:  false,
			expectCreate: false,
			expectUpdate: false,
		},
		{
			name: "private key is empty string",
			existingCM: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      constants.KubeanConfigMapName,
					Namespace: util.GetCurrentNSOrDefault(),
				},
				Data: map[string]string{
					PrivateKey: "",
				},
			},
			expectError:  false,
			expectCreate: true,
			expectUpdate: true,
		},
		{
			name: "no private key field",
			existingCM: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      constants.KubeanConfigMapName,
					Namespace: util.GetCurrentNSOrDefault(),
				},
				Data: map[string]string{},
			},
			expectError:  false,
			expectCreate: true,
			expectUpdate: true,
		},
		{
			name:        "configmap not found",
			existingCM:  nil,
			expectError: true,
			getError:    errors.New("configmap not found"),
		},
		{
			name: "update fails",
			existingCM: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      constants.KubeanConfigMapName,
					Namespace: util.GetCurrentNSOrDefault(),
				},
				Data: map[string]string{},
			},
			expectError: true,
			updateError: errors.New("update failed"),
		},
		{
			name: "create pubkey configmap fails",
			existingCM: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      constants.KubeanConfigMapName,
					Namespace: util.GetCurrentNSOrDefault(),
				},
				Data: map[string]string{},
			},
			expectError: true,
			createError: errors.New("create failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake clientset
			var objects []runtime.Object
			if tt.existingCM != nil {
				objects = append(objects, tt.existingCM)
			}
			clientset := fake.NewSimpleClientset(objects...)

			// Mock errors if specified
			if tt.getError != nil {
				clientset.PrependReactor("get", "configmaps", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, tt.getError
				})
			}
			if tt.updateError != nil {
				clientset.PrependReactor("update", "configmaps", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, tt.updateError
				})
			}
			if tt.createError != nil {
				clientset.PrependReactor("create", "configmaps", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, tt.createError
				})
			}

			// Execute test
			err := InitConfiguration(clientset)

			// Verify results
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Verify actions
			actions := clientset.Actions()

			// Should always have a get action
			if len(actions) < 1 || actions[0].GetVerb() != "get" {
				t.Errorf("expected get action")
			}

			if tt.expectUpdate {
				found := false
				for _, action := range actions {
					if action.GetVerb() == "update" {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected update action")
				}
			}

			if tt.expectCreate {
				found := false
				for _, action := range actions {
					if action.GetVerb() == "create" {
						found = true
						break
					}
				}
				if !found && tt.createError == nil {
					t.Errorf("expected create action")
				}
			}
		})
	}
}
