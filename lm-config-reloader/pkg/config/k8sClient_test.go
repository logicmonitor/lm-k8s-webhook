package config

import (
	"fmt"
	"os"
	"testing"

	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func TestNewK8sClient(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			k8sRestConfig *rest.Config
			k8sClientSet  func(*rest.Config) (kubernetes.Interface, error)
		}
		wantErr bool
	}{
		{
			name: "Get a new k8s client",
			args: struct {
				k8sRestConfig *rest.Config
				k8sClientSet  func(*rest.Config) (kubernetes.Interface, error)
			}{
				k8sRestConfig: nil,
				k8sClientSet: func(r *rest.Config) (kubernetes.Interface, error) {
					return testclient.NewSimpleClientset(), nil
				},
			},
			wantErr: false,
		},
		{
			name: "Get a new k8s client",
			args: struct {
				k8sRestConfig *rest.Config
				k8sClientSet  func(*rest.Config) (kubernetes.Interface, error)
			}{
				k8sRestConfig: nil,
				k8sClientSet: func(r *rest.Config) (kubernetes.Interface, error) {
					return testclient.NewSimpleClientset(), fmt.Errorf("Cannot create k8s clientSet")
				},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewK8sClient(test.args.k8sRestConfig, test.args.k8sClientSet)

			if err == nil && test.wantErr {
				t.Errorf("NewK8sClient() returned nil, instead of error")
			}
			if err != nil && !test.wantErr {
				t.Errorf("NewK8sClient() returned an unexpected error: %+v", err)
			}
		})

	}
}

func TestK8sRestConfig(t *testing.T) {

	t.Run("Test K8sRestConfig with KUBECONFIG set", func(t *testing.T) {
		os.Setenv("KUBECONFIG", "testdata/.kube/config")
		defer os.Unsetenv("KUBECONFIG")
		_, err := K8sRestConfig()
		if err != nil {
			t.Errorf("K8sRestConfig() returned an unexpected error: %+v", err)
		}
	})

	t.Run("Test K8sRestConfig without KUBECONFIG set", func(t *testing.T) {
		os.Setenv("HOME", "testdata")
		defer os.Unsetenv("HOME")

		_, err := K8sRestConfig()
		if err != nil {
			t.Errorf("K8sRestConfig() returned an unexpected error: %+v", err)
		}
	})
}
