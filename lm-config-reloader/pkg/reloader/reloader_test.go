package reloader

import (
	"context"
	"sync"
	"testing"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/config"
	"go.uber.org/zap"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/logger"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/provider"

	configSyncer "github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/sync"

	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

// getFakeK8sClient returns the dummy kubernetes client object for testing
func getFakeK8sClient() (*config.K8sClient, error) {
	return config.NewK8sClient(nil, func(r *rest.Config) (kubernetes.Interface, error) {
		return testclient.NewSimpleClientset(), nil
	})
}

type MockWatcher struct {
	Wg *sync.WaitGroup
}

func (mockWatcher MockWatcher) Watch(ctx context.Context, remoteProvider provider.RemoteProvider, configSyncer configSyncer.ConfigSyncer) error {
	mockWatcher.Wg.Done()
	return nil
}

func TestSetupProviders(t *testing.T) {
	if err := logger.Init("DEBUG"); err != nil {
		t.Error("error occured while initializing the logger", zap.Error(err))
	}
	k8sClient, err := getFakeK8sClient()
	if err != nil {
		t.Errorf("Error occured in getting fake k8s client: %v", err)
		return
	}
	var wg sync.WaitGroup
	tests := []struct {
		name string
		args struct {
			ctx context.Context
		}
		fields struct {
			lmreloader *LMReloader
		}
		wantErr bool
	}{
		{
			name: "setup providers",
			args: struct{ ctx context.Context }{
				ctx: context.Background(),
			},
			fields: struct{ lmreloader *LMReloader }{
				lmreloader: &LMReloader{
					ReloaderConfig: &config.ReloaderConfig{
						Reloaders: []config.Reloader{
							{
								Provider: config.Provider{
									Git: &config.Git{
										Owner:        "test-owner",
										Repo:         "test-repo",
										FilePath:     "test-file-path",
										Ref:          "test-ref",
										PullInterval: "5s",
									},
								},
								Resource: config.Resource{
									"kind":     "ConfigMap",
									"name":     "configMap-name",
									"fileName": "configMap-file-name",
								},
							},
						},
					},
					Watcher: MockWatcher{
						Wg: &wg,
					},
					K8sClient: k8sClient,
					Wg:        &wg,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.fields.lmreloader.SetupProviders(test.args.ctx)
			if err == nil && test.wantErr {
				t.Errorf("SetupProviders() returned nil, instead of error")
			}
			if err != nil && !test.wantErr {
				t.Errorf("SetupProviders() returned an unexpected error: %+v", err)
			}
		})
	}
}
