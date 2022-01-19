package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/config"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/fetcher"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/logger"
	"go.uber.org/zap"

	"github.com/google/go-github/v40/github"
	"golang.org/x/oauth2"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

const (
	baseURLPath = "/api-v3"
)

func setup() (client *github.Client, mux *http.ServeMux, serverURL string, teardown func()) {
	// mux is the HTTP request multiplexer used with the test server.
	mux = http.NewServeMux()

	apiHandler := http.NewServeMux()
	apiHandler.Handle(baseURLPath+"/", http.StripPrefix(baseURLPath, mux))

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(apiHandler)

	// client is the GitHub client being tested and is
	// configured to use test server.
	client = github.NewClient(nil)
	url, _ := url.Parse(server.URL + baseURLPath + "/")
	client.BaseURL = url
	client.UploadURL = url

	return client, mux, server.URL, server.Close
}

func TestCreateGitConfigProvider(t *testing.T) {
	if err := logger.Init("DEBUG"); err != nil {
		t.Error("error occured while initializing the logger", zap.Error(err))
	}
	// cmpOpt := cmp.AllowUnexported()

	tests := []struct {
		name string
		args struct {
			git *config.Git
		}
		wantErr     bool
		wantPayload *GitProvider
	}{
		{
			name: "CreateGitConfigProvider returns GitConfigProvider instance",
			args: struct {
				git *config.Git
			}{
				git: &config.Git{
					Owner:        "test-owner",
					Repo:         "test-repo",
					FilePath:     "test-file-path",
					Ref:          "test-ref",
					PullInterval: "5s",
					Disabled:     false,
				},
			},
			wantErr: false,
			wantPayload: &GitProvider{
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
		{
			name: "CreateGitConfigProvider with error returned by validateGitProviderConfig call",
			args: struct {
				git *config.Git
			}{
				git: &config.Git{
					Owner:        "",
					Repo:         "test-repo",
					FilePath:     "test-file-path",
					Ref:          "test-ref",
					PullInterval: "5s",
					Disabled:     false,
				},
			},
			wantErr:     true,
			wantPayload: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			remoteProvider, err := CreateGitConfigProvider(test.args.git)
			if err == nil && test.wantErr {
				t.Errorf("CreateGitConfigProvider() returned nil, instead of error")
			}
			if err != nil && !test.wantErr {
				t.Errorf("CreateGitConfigProvider() returned an unexpected error: %+v", err)
			}
			if err == nil {
				if gitProvider, ok := remoteProvider.(*GitProvider); ok {
					if !cmp.Equal(test.wantPayload, gitProvider, cmpopts.IgnoreUnexported(GitProvider{})) {
						t.Errorf("CreateGitConfigProvider() returns %v, but expected is %v", gitProvider, test.wantPayload)
						return
					}
				} else {
					t.Error("error occured in asserting remoteProvider to GitProvider")
				}
			}
		})
	}
}
func testMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}
func TestFetch(t *testing.T) {
	if err := logger.Init("DEBUG"); err != nil {
		t.Error("error occured while initializing the logger", zap.Error(err))
	}
	// cmpOpt := cmp.AllowUnexported()
	client, mux, _, teardown := setup()
	defer teardown()
	mux.HandleFunc("/repos/test-owner/test-repo/contents/LICENSE", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
		  "type": "file",
		  "encoding": "base64",
		  "size": 20678,
		  "name": "LICENSE",
		  "path": "LICENSE",
		  "content": "TElDRU5TRQo="
		}`)
	})

	mux.HandleFunc("/repos/test-owner/test-repo/contents/example.txt", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
		  "type": "file",
		  "encoding": "base64",
		  "size": 20678,
		  "name": "example.txt",
		  "path": "example.txt",
		  "content": "TElDRU5TRQo"
		}`)
	})

	tests := []struct {
		name string
		args struct {
			ctx context.Context
		}
		fields struct {
			GitProvider *GitProvider
		}
		wantErr     bool
		wantPayload *fetcher.Response
	}{
		{
			name: "Fetch returns file content",
			args: struct {
				ctx context.Context
			}{
				ctx: context.TODO(),
			},
			wantErr: false,
			fields: struct {
				GitProvider *GitProvider
			}{
				GitProvider: &GitProvider{
					DefaultRemoteProvider: DefaultRemoteProvider{
						Provider:     gitProviderName,
						PullInterval: "5s",
					},
					Git: config.Git{
						Owner:        "test-owner",
						Repo:         "test-repo",
						FilePath:     "LICENSE",
						Ref:          "test-ref",
						PullInterval: "5s",
						AuthRequired: false,
						Disabled:     false,
					},
					client: client,
				},
			},
			wantPayload: &fetcher.Response{FileName: "LICENSE", FileData: decodeBase64("TElDRU5TRQo=")},
		},
		{
			name: "Fetch with wrong file path",
			args: struct {
				ctx context.Context
			}{
				ctx: context.TODO(),
			},
			wantErr: true,
			fields: struct {
				GitProvider *GitProvider
			}{
				GitProvider: &GitProvider{
					DefaultRemoteProvider: DefaultRemoteProvider{
						Provider:     gitProviderName,
						PullInterval: "5s",
					},
					Git: config.Git{
						Owner:        "test-owner",
						Repo:         "test-repo",
						FilePath:     "demo.txt",
						Ref:          "test-ref",
						PullInterval: "5s",
						AuthRequired: false,
						Disabled:     false,
					},
					client: client,
				},
			},
			wantPayload: nil,
		},
		{
			name: "Fetch with wrong base64 encoded content received in response",
			args: struct {
				ctx context.Context
			}{
				ctx: context.TODO(),
			},
			wantErr: true,
			fields: struct {
				GitProvider *GitProvider
			}{
				GitProvider: &GitProvider{
					DefaultRemoteProvider: DefaultRemoteProvider{
						Provider:     gitProviderName,
						PullInterval: "5s",
					},
					Git: config.Git{
						Owner:        "test-owner",
						Repo:         "test-repo",
						FilePath:     "example.txt",
						Ref:          "test-ref",
						PullInterval: "5s",
						AuthRequired: false,
						Disabled:     false,
					},
					client: client,
				},
			},
			wantPayload: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := test.fields.GitProvider.Fetch(test.args.ctx)
			if err == nil && test.wantErr {
				t.Errorf("Fetch() returned nil, instead of error")
			}
			if err != nil && !test.wantErr {
				t.Errorf("Fetch() returned an unexpected error: %+v", err)
			}
			if !cmp.Equal(test.wantPayload, resp) {
				t.Errorf("Fetch() returns %v, but expected is %v", resp, test.wantPayload)
				return
			}
		})
	}
}

func TestConfigureClient(t *testing.T) {
	tests := []struct {
		name   string
		fields struct {
			GitProvider *GitProvider
		}
		wantErr     bool
		wantPayload *GitProvider
	}{
		{
			name: "ConfigureClient without authentication",
			fields: struct {
				GitProvider *GitProvider
			}{
				GitProvider: &GitProvider{},
			},
			wantErr: false,
			wantPayload: &GitProvider{
				client: github.NewClient(nil),
			},
		},
		{
			name: "ConfigureClient with authentication",
			fields: struct {
				GitProvider *GitProvider
			}{
				GitProvider: &GitProvider{
					Git: config.Git{
						AuthRequired: true,
						AccessToken:  "abcd",
					},
				},
			},
			wantErr: false,
			wantPayload: &GitProvider{
				client: github.NewClient(oauth2.NewClient(context.TODO(), oauth2.StaticTokenSource(
					&oauth2.Token{AccessToken: "abcd"},
				))),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.fields.GitProvider.configureClient()
			if !cmp.Equal(test.wantPayload.client.BaseURL.String(), test.fields.GitProvider.client.BaseURL.String()) ||
				!cmp.Equal(test.wantPayload.client.UserAgent, test.fields.GitProvider.client.UserAgent) {
				t.Errorf("configureClient() set client in GitProvider %v, but expected is %v", test.fields.GitProvider, test.wantPayload)
				return
			}
		})
	}
}

func decodeBase64(base64Data string) []byte {
	data, _ := base64.StdEncoding.DecodeString(base64Data)
	return data
}
