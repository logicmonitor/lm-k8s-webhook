package config

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func TestLoadConfig(t *testing.T) {

	cmpOpt := cmp.AllowUnexported()

	tests := []struct {
		name string
		args struct {
			configFilePath string
		}
		wantErr     bool
		wantPayload Config
	}{
		{
			name:    "load config with correct file path",
			args:    struct{ configFilePath string }{configFilePath: "testdata/config.yaml"},
			wantErr: false,
			wantPayload: Config{
				MutationConfigProvided: true,
				MutationConfig: MutationConfig{
					LMEnvVars: LMEnvVars{Resource: []corev1.EnvVar{
						{
							Name:      "SERVICE_ACCOUNT_NAME",
							ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.serviceAccountName"}},
						},
						{
							Name:      "SERVICE_NAMESPACE",
							ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['app-namespace']"}},
						},
						{
							Name:      "SERVICE_NAME",
							ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['app-name']"}},
						},
					}, Operation: []corev1.EnvVar{
						{
							Name:  "COMPANY_NAME",
							Value: "ABC Corporation",
						},
						{
							Name:  "OTLP_ENDPOINT",
							Value: "lmotel-svc:4317",
						},
						{
							Name:  "OTEL_JAVAAGENT_ENABLED",
							Value: "true",
						},
					}},
				},
			},
		},

		{
			name:        "load config with incorrect file path",
			args:        struct{ configFilePath string }{configFilePath: "testdata/config1.yaml"},
			wantErr:     true,
			wantPayload: Config{},
		},

		{
			name:        "load config with incorrect file content",
			args:        struct{ configFilePath string }{configFilePath: "testdata/config_with_error.yaml"},
			wantErr:     true,
			wantPayload: Config{},
		},
	}

	for _, tt := range tests {
		cfg = Config{}
		err := LoadConfig(tt.args.configFilePath)

		if (err != nil) != tt.wantErr {
			t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			return
		}

		if !cmp.Equal(cfg, tt.wantPayload, cmpOpt) {
			t.Errorf("LoadConfig() returned config = %v, but expected config = %v", cfg, tt.wantPayload)
			return
		}
	}
}

func TestGetConfig(t *testing.T) {
	cmpOpt := cmp.AllowUnexported()

	tests := []struct {
		name string
		args struct {
			configFilePath string
		}
		wantErr     bool
		wantPayload Config
	}{
		{
			name:    "Test GetConfig",
			args:    struct{ configFilePath string }{configFilePath: "testdata/config.yaml"},
			wantErr: false,
			wantPayload: Config{
				MutationConfigProvided: true,
				MutationConfig: MutationConfig{
					LMEnvVars: LMEnvVars{Resource: []corev1.EnvVar{
						{
							Name:      "SERVICE_ACCOUNT_NAME",
							ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.serviceAccountName"}},
						},
						{
							Name:      "SERVICE_NAMESPACE",
							ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['app-namespace']"}},
						},
						{
							Name:      "SERVICE_NAME",
							ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['app-name']"}},
						},
					}, Operation: []corev1.EnvVar{
						{
							Name:  "COMPANY_NAME",
							Value: "ABC Corporation",
						},
						{
							Name:  "OTLP_ENDPOINT",
							Value: "lmotel-svc:4317",
						},
						{
							Name:  "OTEL_JAVAAGENT_ENABLED",
							Value: "true",
						},
					}},
				},
			},
		},
	}

	for _, tt := range tests {
		cfg = Config{}
		err := LoadConfig(tt.args.configFilePath)

		if (err != nil) != tt.wantErr {
			t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)
			return
		}

		if !cmp.Equal(GetConfig(), tt.wantPayload, cmpOpt) {
			t.Errorf("GetConfig() returned config = %v, but expected config = %v", cfg, tt.wantPayload)
			return
		}
	}
}

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

	for _, tt := range tests {
		_, err := NewK8sClient(tt.args.k8sRestConfig, tt.args.k8sClientSet)

		if (err != nil) != tt.wantErr {
			t.Errorf("NewK8sClient() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
	}
}
