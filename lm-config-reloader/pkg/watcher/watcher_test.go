package watcher

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/config"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/fetcher"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/logger"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/provider"
	configSyncer "github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/sync"
	"go.uber.org/zap"

	"golang.org/x/net/context"
)

type MockRemoteProvider struct {
	provider.DefaultRemoteProvider
}

func (mockRemoteProvider *MockRemoteProvider) Fetch(ctx context.Context) (*fetcher.Response, error) {
	return nil, nil
}

type MockConfigSyncer struct {
}

func (mockConfigSyncer MockConfigSyncer) Sync(response *fetcher.Response) error {
	return nil
}

type MockRemoteProviderErr struct {
	provider.DefaultRemoteProvider
}

func (mockRemoteProviderErr *MockRemoteProviderErr) Fetch(ctx context.Context) (*fetcher.Response, error) {
	return nil, fmt.Errorf("error in fetch operation")
}

type MockConfigSyncerErr struct {
}

func (mockConfigSyncerErr MockConfigSyncerErr) Sync(response *fetcher.Response) error {
	return fmt.Errorf("error in sync operation")
}

func TestWatch(t *testing.T) {
	if err := logger.Init("DEBUG"); err != nil {
		t.Error("error occured while initializing the logger", zap.Error(err))
	}
	ctx1, cancel1 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel1()

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	ctx3, cancel3 := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel3()

	tests := []struct {
		name string
		args struct {
			ctx            context.Context
			Reloader       config.Reloader
			RemoteProvider provider.RemoteProvider
			ConfigSyncer   configSyncer.ConfigSyncer
		}
		fields struct {
			watcher RemoteConfigWatcher
		}
		wantErr bool
	}{
		{
			name: "test Watch",
			args: struct {
				ctx            context.Context
				Reloader       config.Reloader
				RemoteProvider provider.RemoteProvider
				ConfigSyncer   configSyncer.ConfigSyncer
			}{
				ctx: ctx1,
				RemoteProvider: &MockRemoteProvider{
					DefaultRemoteProvider: provider.DefaultRemoteProvider{
						Provider:     "mock",
						PullInterval: "1s",
					},
				},
				ConfigSyncer: MockConfigSyncer{},
			},
			fields: struct {
				watcher RemoteConfigWatcher
			}{
				watcher: RemoteConfigWatcher{
					Wg: &sync.WaitGroup{},
				},
			},
			wantErr: true,
		},
		{
			name: "test Watch with fetch operation returning an error",
			args: struct {
				ctx            context.Context
				Reloader       config.Reloader
				RemoteProvider provider.RemoteProvider
				ConfigSyncer   configSyncer.ConfigSyncer
			}{
				ctx: ctx2,
				RemoteProvider: &MockRemoteProviderErr{
					DefaultRemoteProvider: provider.DefaultRemoteProvider{
						Provider:     "mock",
						PullInterval: "1s",
					},
				},
				ConfigSyncer: MockConfigSyncer{},
			},
			fields: struct {
				watcher RemoteConfigWatcher
			}{
				watcher: RemoteConfigWatcher{
					Wg: &sync.WaitGroup{},
				},
			},
			wantErr: true,
		},
		{
			name: "test Watch with sync operation returning an error",
			args: struct {
				ctx            context.Context
				Reloader       config.Reloader
				RemoteProvider provider.RemoteProvider
				ConfigSyncer   configSyncer.ConfigSyncer
			}{
				ctx: ctx3,
				RemoteProvider: &MockRemoteProvider{
					DefaultRemoteProvider: provider.DefaultRemoteProvider{
						Provider:     "mock",
						PullInterval: "1s",
					},
				},
				ConfigSyncer: MockConfigSyncerErr{},
			},
			fields: struct {
				watcher RemoteConfigWatcher
			}{
				watcher: RemoteConfigWatcher{
					Wg: &sync.WaitGroup{},
				},
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.fields.watcher.Wg.Add(1)
			err := test.fields.watcher.Watch(test.args.ctx, test.args.RemoteProvider, test.args.ConfigSyncer)
			if err == nil && test.wantErr {
				t.Errorf("Watch() returned nil, instead of error")
			}
			if err != nil && !test.wantErr {
				t.Errorf("Watch() returned an unexpected error: %+v", err)
			}
		})
	}
}
