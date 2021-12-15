package reloader

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ghodss/yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/logicmonitor/lm-k8s-webhook/pkg/config"
	corev1 "k8s.io/api/core/v1"
)

func TestSetupConfigReloader(t *testing.T) {
	cmpOpt := cmp.AllowUnexported()
	tests := []struct {
		name string
		args struct {
			ctx              context.Context
			lmconfigFilePath string
		}
		wantErr     bool
		wantPayload config.Config
	}{
		{
			name: "Test SetupConfigReloader with correct file path to watch",
			args: struct {
				ctx              context.Context
				lmconfigFilePath string
			}{
				context.Background(),
				"testdata/config.yaml",
			},
			wantErr: false,
			wantPayload: config.Config{
				MutationConfigProvided: true,
				MutationConfig: config.MutationConfig{
					LMEnvVars: config.LMEnvVars{Resource: []config.ResourceEnv{
						{
							Env: corev1.EnvVar{
								Name:      "SERVICE_ACCOUNT_NAME",
								ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.serviceAccountName"}},
							},
						},
						{
							Env: corev1.EnvVar{
								Name:      "SERVICE_NAMESPACE",
								ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['app-namespace']"}},
							},
						},
						{
							Env: corev1.EnvVar{
								Name:      "SERVICE_NAME",
								ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['app-name']"}},
							},
						},
					}, Operation: []config.OperationEnv{
						{
							Env: corev1.EnvVar{
								Name:  "COMPANY_NAME",
								Value: "ABC Corporation",
							},
						},
						{
							Env: corev1.EnvVar{
								Name:  "OTLP_ENDPOINT",
								Value: "lmotel-svc:4317",
							},
						},
						{
							Env: corev1.EnvVar{
								Name:  "OTEL_JAVAAGENT_ENABLED",
								Value: "true",
							},
						},
					}},
				},
			},
		},
		{
			name: "Test SetupConfigReloader with correct file path to watch",
			args: struct {
				ctx              context.Context
				lmconfigFilePath string
			}{
				context.Background(),
				"testdata/config1.yaml",
			},
			wantErr:     true,
			wantPayload: config.Config{},
		},
	}
	for _, tt := range tests {
		err := SetupConfigReloader(tt.args.ctx, tt.args.lmconfigFilePath)

		if (err != nil) != tt.wantErr {
			t.Errorf("SetupConfigReloader() return error = %v, but expected error = %v", err, tt.wantErr)
			return
		}

		file, err := os.OpenFile(tt.args.lmconfigFilePath, os.O_RDWR|os.O_TRUNC, 0755)
		if err != nil {
			logger.Error(err, "error opening a config file", "path", tt.args.lmconfigFilePath)
			return
		}
		defer file.Close()

		configBytes, err := yaml.Marshal(tt.wantPayload.MutationConfig)
		if err != nil {
			logger.Error(err, "error in marshalling", "path", tt.args.lmconfigFilePath)
			return
		}
		_, err = file.Write(configBytes)
		if err != nil {
			logger.Error(err, "error writing a config file", "path", tt.args.lmconfigFilePath)
			return
		}
		time.Sleep(5 * time.Second)

		if !cmp.Equal(config.GetConfig(), tt.wantPayload, cmpOpt) {
			t.Errorf("updated config = %v, but expected config = %v", config.GetConfig(), tt.wantPayload)
			return
		}
	}
}

func TestReloadConfig(t *testing.T) {
	reloadCh := make(chan bool, 1)
	reloadDoneCh := make(chan bool, 1)

	reloadCh <- true

	tests := []struct {
		name string
		args struct {
			ctx              context.Context
			reloadCh         chan bool
			reloadDoneCh     chan bool
			lmconfigFilePath string
		}
		wantErr bool
	}{
		{
			name: "Test ReloadConfig for reload event",
			args: struct {
				ctx              context.Context
				reloadCh         chan bool
				reloadDoneCh     chan bool
				lmconfigFilePath string
			}{
				context.Background(),
				reloadCh,
				reloadDoneCh,
				"testdata/config.yaml",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		go reloadConfig(tt.args.ctx, tt.args.reloadCh, tt.args.reloadDoneCh, tt.args.lmconfigFilePath)
		done := <-tt.args.reloadDoneCh
		if done != true {
			t.Errorf("reload done channel returned = %v, but expected = %v", done, true)
			return
		}
	}
}

func TestReloadConfigWithContextCancelled(t *testing.T) {
	reloadCh := make(chan bool, 1)
	reloadDoneCh := make(chan bool, 1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name string
		args struct {
			ctx              context.Context
			reloadCh         chan bool
			reloadDoneCh     chan bool
			lmconfigFilePath string
		}
		wantErr error
	}{
		{
			name: "Test ReloadConfig for canceled context",
			args: struct {
				ctx              context.Context
				reloadCh         chan bool
				reloadDoneCh     chan bool
				lmconfigFilePath string
			}{
				ctx,
				reloadCh,
				reloadDoneCh,
				"testdata/config.yaml",
			},
			wantErr: context.Canceled,
		},
	}
	for _, tt := range tests {
		go reloadConfig(tt.args.ctx, tt.args.reloadCh, tt.args.reloadDoneCh, tt.args.lmconfigFilePath)
		if tt.args.ctx.Err() != tt.wantErr {
			t.Errorf("Reload Config returned = %v, but expected = %v", tt.args.ctx.Err(), tt.wantErr)
			return
		}
	}
}
