package provider

import (
	"testing"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/config"
	// "github.com/google/go-cmp/cmp"
)

func TestValidateProviderConfig(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			provider config.Provider
		}
		wantErr bool
	}{
		{
			name: "ValidateProviderConfig for git provider without validation error",
			args: struct {
				provider config.Provider
			}{
				provider: config.Provider{
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
			},
			wantErr: false,
		},
		{
			name: "ValidateProviderConfig for git provider with validation error",
			args: struct {
				provider config.Provider
			}{
				provider: config.Provider{
					Git: &config.Git{
						Owner:        "test-owner",
						Repo:         "test-repo",
						FilePath:     "LICENSE",
						Ref:          "test-ref",
						PullInterval: "5",
						AuthRequired: false,
						Disabled:     false,
					},
				},
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		err := ValidateProviderConfig(test.args.provider)
		if err == nil && test.wantErr {
			t.Errorf("ValidateProviderConfig() returned nil, instead of error")
		}
		if err != nil && !test.wantErr {
			t.Errorf("ValidateProviderConfig() returned an unexpected error: %+v", err)
		}
	}
}

func TestValidateGitProviderConfig(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			gitProvider GitProvider
		}
		wantErr bool
	}{
		{
			name: "validateGitProviderConfig without validation error",
			args: struct {
				gitProvider GitProvider
			}{
				gitProvider: GitProvider{
					DefaultRemoteProvider: DefaultRemoteProvider{
						Provider:     gitProviderName,
						PullInterval: "5s",
					},
					Git: config.Git{
						Owner:        "test-owner",
						Repo:         "test-repo",
						FilePath:     "test-file-path",
						Ref:          "test-ref",
						PullInterval: "5s",
						AuthRequired: false,
						Disabled:     false,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "validateGitProviderConfig with validation error in Owner",
			args: struct {
				gitProvider GitProvider
			}{
				gitProvider: GitProvider{
					DefaultRemoteProvider: DefaultRemoteProvider{
						Provider:     gitProviderName,
						PullInterval: "5s",
					},
					Git: config.Git{
						Owner:        "",
						Repo:         "test-repo",
						FilePath:     "test-file-path",
						Ref:          "test-ref",
						PullInterval: "5s",
						AuthRequired: false,
						Disabled:     false,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "validateGitProviderConfig with validation error in Repo",
			args: struct {
				gitProvider GitProvider
			}{
				gitProvider: GitProvider{
					DefaultRemoteProvider: DefaultRemoteProvider{
						Provider:     gitProviderName,
						PullInterval: "5s",
					},
					Git: config.Git{
						Owner:        "test-owner",
						Repo:         "",
						FilePath:     "test-file-path",
						Ref:          "test-ref",
						PullInterval: "5s",
						AuthRequired: false,
						Disabled:     false,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "validateGitProviderConfig with validation error in FilePath",
			args: struct {
				gitProvider GitProvider
			}{
				gitProvider: GitProvider{
					DefaultRemoteProvider: DefaultRemoteProvider{
						Provider:     gitProviderName,
						PullInterval: "5s",
					},
					Git: config.Git{
						Owner:        "test-owner",
						Repo:         "test-repo",
						FilePath:     "",
						Ref:          "test-ref",
						PullInterval: "5s",
						AuthRequired: false,
						Disabled:     false,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "validateGitProviderConfig with validation error in PullInterval",
			args: struct {
				gitProvider GitProvider
			}{
				gitProvider: GitProvider{
					DefaultRemoteProvider: DefaultRemoteProvider{
						Provider:     gitProviderName,
						PullInterval: "5",
					},
					Git: config.Git{
						Owner:        "test-owner",
						Repo:         "test-repo",
						FilePath:     "test-file-path",
						Ref:          "test-ref",
						PullInterval: "5",
						AuthRequired: false,
						Disabled:     false,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "validateGitProviderConfig with validation error in AccessToken",
			args: struct {
				gitProvider GitProvider
			}{
				gitProvider: GitProvider{
					DefaultRemoteProvider: DefaultRemoteProvider{
						Provider:     gitProviderName,
						PullInterval: "5s",
					},
					Git: config.Git{
						Owner:        "test-owner",
						Repo:         "test-repo",
						FilePath:     "test-file-path",
						Ref:          "test-ref",
						PullInterval: "5s",
						AuthRequired: true,
						AccessToken:  "",
						Disabled:     false,
					},
				},
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		err := validateGitProviderConfig(test.args.gitProvider)
		if err == nil && test.wantErr {
			t.Errorf("validateGitProviderConfig() returned nil, instead of error")
		}
		if err != nil && !test.wantErr {
			t.Errorf("validateGitProviderConfig() returned an unexpected error: %+v", err)
		}
	}
}
