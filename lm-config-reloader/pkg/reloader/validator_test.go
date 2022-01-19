package reloader

import (
	"testing"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/config"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			reloaderConfig *config.ReloaderConfig
		}
		wantErr bool
	}{
		{
			name: "with correct config",
			args: struct {
				reloaderConfig *config.ReloaderConfig
			}{
				reloaderConfig: &config.ReloaderConfig{
					Reloaders: []config.Reloader{
						{
							Provider: config.Provider{
								Git: &config.Git{
									Owner:        "test-owner",
									Repo:         "test-repo",
									FilePath:     "LICENSE",
									Ref:          "test-ref",
									PullInterval: "5s",
									AuthRequired: false,
									Disabled:     false,
								},
							},
							Resource: config.Resource{
								"kind":     "ConfigMap",
								"name":     "configMap-name",
								"fileName": "configMap-file-name",
							},
							ReloadEndpoint: "http://127.0.0.1:3030/reload",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "with invalid reload endpoint format",
			args: struct {
				reloaderConfig *config.ReloaderConfig
			}{
				reloaderConfig: &config.ReloaderConfig{
					Reloaders: []config.Reloader{
						{
							Provider: config.Provider{
								Git: &config.Git{
									Owner:        "test-owner",
									Repo:         "test-repo",
									FilePath:     "LICENSE",
									Ref:          "test-ref",
									PullInterval: "5s",
									AuthRequired: false,
									Disabled:     false,
								},
							},
							Resource: config.Resource{
								"kind":     "ConfigMap",
								"name":     "configMap-name",
								"fileName": "configMap-file-name",
							},
							ReloadEndpoint: "invalid-url",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "with invalid provider config",
			args: struct {
				reloaderConfig *config.ReloaderConfig
			}{
				reloaderConfig: &config.ReloaderConfig{
					Reloaders: []config.Reloader{
						{
							Provider: config.Provider{
								Git: &config.Git{
									Owner:        "",
									Repo:         "test-repo",
									FilePath:     "LICENSE",
									Ref:          "test-ref",
									PullInterval: "5s",
									AuthRequired: false,
									Disabled:     false,
								},
							},
							Resource: config.Resource{
								"kind":     "ConfigMap",
								"name":     "configMap-name",
								"fileName": "configMap-file-name",
							},
							ReloadEndpoint: "http://127.0.0.1:3030/reload",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "with invalid resource config",
			args: struct {
				reloaderConfig *config.ReloaderConfig
			}{
				reloaderConfig: &config.ReloaderConfig{
					Reloaders: []config.Reloader{
						{
							Provider: config.Provider{
								Git: &config.Git{
									Owner:        "test-owner",
									Repo:         "test-repo",
									FilePath:     "LICENSE",
									Ref:          "test-ref",
									PullInterval: "5s",
									AuthRequired: false,
									Disabled:     false,
								},
							},
							Resource: config.Resource{
								"kind":     "Invalid-kind",
								"name":     "configMap-name",
								"fileName": "configMap-file-name",
							},
							ReloadEndpoint: "http://127.0.0.1:3030/reload",
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		err := Validate(test.args.reloaderConfig)
		if err == nil && test.wantErr {
			t.Errorf("Validate() returned nil, instead of error")
		}
		if err != nil && !test.wantErr {
			t.Errorf("Validate() returned an unexpected error: %+v", err)
		}
	}
}
