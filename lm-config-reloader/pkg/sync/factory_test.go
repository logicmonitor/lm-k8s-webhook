package sync

import (
	"testing"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/config"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// getFakeK8sClient returns the dummy kubernetes client object for testing
func getFakeK8sClient() (*config.K8sClient, error) {
	return config.NewK8sClient(nil, func(r *rest.Config) (kubernetes.Interface, error) {
		return testclient.NewSimpleClientset(
			&corev1.ConfigMap{ObjectMeta: v1.ObjectMeta{Name: "test-configmap", Namespace: "default"}, Data: map[string]string{"test-file": "test-data"}}), nil
	})
}

func TestCreateConfigSyncer(t *testing.T) {

	k8sClient, err := getFakeK8sClient()
	if err != nil {
		t.Errorf("Error occured in getting fake k8s client: %v", err)
		return
	}
	tests := []struct {
		name string
		args struct {
			reloader  config.Reloader
			k8sClient *config.K8sClient
		}
		wantErr     bool
		wantPayload ConfigSyncer
	}{
		{
			name: "CreateConfigSyncer for configmap",
			args: struct {
				reloader  config.Reloader
				k8sClient *config.K8sClient
			}{
				reloader: config.Reloader{
					Provider: config.Provider{
						Git: &config.Git{
							Owner:        "test-owner",
							Repo:         "test-repo",
							FilePath:     "test-file-path",
							Ref:          "test-ref",
							PullInterval: "5s",
							AuthRequired: false,
							Disabled:     false,
						},
					},
					Resource: config.Resource{
						"kind":     ConfigMapResourceKind,
						"name":     "test-configmap",
						"fileName": "test-fileName",
					},
				},
				k8sClient: k8sClient,
			},
			wantErr: false,
			wantPayload: configMapConfigSyncer{
				Resource: ConfigMapResource{
					Name:     "test-configmap",
					FileName: "test-fileName",
				},
				k8sClient: k8sClient,
			},
		},
		{
			name: "CreateConfigSyncer for MutatingWebhookConfigurationKind",
			args: struct {
				reloader  config.Reloader
				k8sClient *config.K8sClient
			}{
				reloader: config.Reloader{
					Provider: config.Provider{
						Git: &config.Git{
							Owner:        "test-owner",
							Repo:         "test-repo",
							FilePath:     "test-file-path",
							Ref:          "test-ref",
							PullInterval: "5s",
							AuthRequired: false,
							Disabled:     false,
						},
					},
					Resource: config.Resource{
						"kind": MutatingWebhookConfigurationKind,
						"name": "test-MutatingWebhookConfiguration",
					},
				},
				k8sClient: k8sClient,
			},
			wantErr: false,
			wantPayload: mutatingWebhookConfigSyncer{
				Resource: MutatingWebhookConfigurationResource{
					Name: "test-MutatingWebhookConfiguration",
				},
			},
		},
		{
			name: "CreateConfigSyncer with invalid kind",
			args: struct {
				reloader  config.Reloader
				k8sClient *config.K8sClient
			}{
				reloader: config.Reloader{
					Provider: config.Provider{
						Git: &config.Git{
							Owner:        "test-owner",
							Repo:         "test-repo",
							FilePath:     "test-file-path",
							Ref:          "test-ref",
							PullInterval: "5s",
							AuthRequired: false,
							Disabled:     false,
						},
					},
					Resource: config.Resource{
						"kind": "invalid-kind",
						"name": "test-invalid-resource",
					},
				},
				k8sClient: k8sClient,
			},
			wantErr:     true,
			wantPayload: nil,
		},
	}
	for _, test := range tests {
		syncer, err := CreateConfigSyncer(test.args.reloader, k8sClient)
		if err == nil && test.wantErr {
			t.Errorf("CreateConfigSyncer() returned nil, instead of error")
		}
		if err != nil && !test.wantErr {
			t.Errorf("CreateConfigSyncer() returned an unexpected error: %+v", err)
		}

		if !cmp.Equal(test.wantPayload, syncer, cmpopts.IgnoreUnexported(configMapConfigSyncer{}, mutatingWebhookConfigSyncer{})) {
			t.Errorf("CreateConfigSyncer() returns %v, but expected is %v", syncer, test.wantPayload)
			return
		}
	}
}
