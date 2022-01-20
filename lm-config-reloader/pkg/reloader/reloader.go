package reloader

import (
	"context"
	"sync"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/config"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/logger"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/provider"
	configSync "github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/sync"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/watcher"

	"go.uber.org/zap"
)

// LMReloader holds the reloader config and watcher
type LMReloader struct {
	ReloaderConfig *config.ReloaderConfig
	Watcher        watcher.Watcher
	K8sClient      *config.K8sClient
	Wg             *sync.WaitGroup
}

// SetupProviders setups the providers and starts watching the config from providers
func (lmReloader *LMReloader) SetupProviders(ctx context.Context) error {
	for _, reloader := range lmReloader.ReloaderConfig.Reloaders {
		var remoteProvider provider.RemoteProvider
		var err error
		// Check for Git provider
		if reloader.Provider.Git != nil && !reloader.Provider.Git.Disabled {

			logger.Logger().Info("found git provider config",
				zap.String("owner", reloader.Provider.Git.Owner),
				zap.String("repo", reloader.Provider.Git.Repo),
				zap.String("filepath", reloader.Provider.Git.FilePath),
			)

			remoteProvider, err = provider.CreateGitConfigProvider(reloader.Provider.Git)
			if err != nil {
				return err
			}
		}
		// Create configSyncer for the resource
		syncer, err := configSync.CreateConfigSyncer(reloader, lmReloader.K8sClient)
		if err != nil {
			return err
		}
		lmReloader.Wg.Add(1)
		// nolint
		go lmReloader.Watcher.Watch(ctx, remoteProvider, syncer)
	}
	lmReloader.Wg.Wait()
	logger.Logger().Debug("all gouroutines returned")
	logger.Logger().Info("shutting down all watchers")
	return nil
}
