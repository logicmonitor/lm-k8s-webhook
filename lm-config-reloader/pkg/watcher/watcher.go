package watcher

import (
	"context"
	"sync"
	"time"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/config"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/logger"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/provider"
	configSyncer "github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/sync"

	"go.uber.org/zap"
)

const gitDefaultPullInterval string = "20s"

// Watcher has the watch method to watch the config
type Watcher interface {
	Watch(context.Context, provider.RemoteProvider, configSyncer.ConfigSyncer) error
}

// RemoteConfigWatcher represents the default config watcher
type RemoteConfigWatcher struct {
	K8sClient *config.K8sClient
	Wg        *sync.WaitGroup
}

// Watch watches the config from the config provider and call the config sync
func (watcher RemoteConfigWatcher) Watch(ctx context.Context, remoteProvider provider.RemoteProvider, configSyncer configSyncer.ConfigSyncer) error {
	defer watcher.Wg.Done()
	pullInterval, _ := remoteProvider.GetParsedPullInterval()

	if pullInterval == 0 {
		remoteProvider.SetPullInterval(gitDefaultPullInterval)
		pullInterval, _ = remoteProvider.GetParsedPullInterval()
	}

	ticker := time.NewTicker(pullInterval)
	for {
		select {
		case <-ticker.C:
			// Fetch config from github repo
			gitResp, err := remoteProvider.Fetch(ctx)
			if err != nil {
				// TODO: Need for limit the failure count
				logger.Logger().Error("error in fetching the config", zap.Error(err))
				continue
			}
			// sync config
			err = configSyncer.Sync(gitResp)
			if err != nil {
				// TODO: Need to limit the failure count
				logger.Logger().Error("error in config sync", zap.Error(err))
			}
		case <-ctx.Done():
			ticker.Stop()
			return ctx.Err()
		}
	}
}
