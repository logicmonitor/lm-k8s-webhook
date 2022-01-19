package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			configFilePath string
		}
		wantErr     bool
		wantPayload *ReloaderConfig
	}{
		{
			name: "Test LoadConfig",
			args: struct {
				configFilePath string
			}{
				configFilePath: "testdata/reload-config.yaml",
			},
			wantErr: false,
			wantPayload: &ReloaderConfig{
				Reloaders: []Reloader{
					{
						Provider: Provider{
							Git: &Git{
								Owner:        "test-owner",
								Repo:         "test-repo",
								FilePath:     "test-file-path",
								Ref:          "test-ref",
								AuthRequired: true,
								AccessToken:  "test-access-token",
								PullInterval: "5s",
							},
						},
						Resource: Resource{
							"kind":     "ConfigMap",
							"name":     "test-configmap",
							"fileName": "test-configMap-file-path",
						},
					},
					{
						Provider: Provider{
							Git: &Git{
								Owner:        "test-owner",
								Repo:         "test-repo",
								FilePath:     "test-file-path",
								Ref:          "test-ref",
								AuthRequired: true,
								AccessToken:  "test-access-token",
								PullInterval: "5s",
							},
						},
						Resource: Resource{
							"kind": "MutatingWebhookConfiguration",
							"name": "test-mutating-webhook-configuration",
						},
					},
				},
			},
		},
		{
			name:        "LoadConfig with incorrect file path",
			args:        struct{ configFilePath string }{configFilePath: "testdata/config1.yaml"},
			wantErr:     true,
			wantPayload: nil,
		},

		{
			name:        "LoadConfig with incorrect file content",
			args:        struct{ configFilePath string }{configFilePath: "testdata/reload-config-with-error.yaml"},
			wantErr:     true,
			wantPayload: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reloaderCfg, err := LoadConfig(test.args.configFilePath)
			if err == nil && test.wantErr {
				t.Errorf("LoadConfig() returned nil, instead of error")
			}
			if err != nil && !test.wantErr {
				t.Errorf("LoadConfig() returned an unexpected error: %+v", err)
			}
			if !cmp.Equal(test.wantPayload, reloaderCfg) {
				t.Errorf("LoadConfig() returns %v, but expected is %v", reloaderCfg, test.wantPayload)
				return
			}
		})
	}
}
