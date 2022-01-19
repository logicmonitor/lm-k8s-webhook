package reloader

import (
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/config"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/provider"
	k8sSync "github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/sync"
)

// Validate validates reloader config
func Validate(reloaderConfig *config.ReloaderConfig) error {
	for _, reloader := range reloaderConfig.Reloaders {
		if err := provider.ValidateProviderConfig(reloader.Provider); err != nil {
			return err
		}
		if err := k8sSync.ValidateResourceConfig(reloader.Resource); err != nil {
			return err
		}
		if reloader.ReloadEndpoint != "" {
			if err := k8sSync.ValidateReloadEndpoint(reloader.ReloadEndpoint); err != nil {
				return err
			}
		}
	}
	return nil
}
