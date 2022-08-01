package reloader

import (
	"context"

	"github.com/fsnotify/fsnotify"
	lmk8swebhookconfig "github.com/logicmonitor/lm-k8s-webhook/pkg/config"
	"github.com/spf13/viper"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var logger = log.Log.WithName("reloader")

// SetupConfigReloader starts watching for update event on the config file
func SetupConfigReloader(ctx context.Context, lmconfigFilePath string) error {
	reload := make(chan bool, 1)
	reloadDone := make(chan bool, 1)
	v := viper.New()
	v.SetConfigFile(lmconfigFilePath)
	err := v.ReadInConfig()
	if err != nil {
		logger.Error(err, "lmconfigFilePath", lmconfigFilePath)
		return err
	}
	v.WatchConfig()
	go reloadConfig(ctx, reload, reloadDone, lmconfigFilePath)
	v.OnConfigChange(func(e fsnotify.Event) {
		logger.Info("Config file changed", "event", e)
		reload <- true
		<-reloadDone
		logger.Info("Config file reload success")
	})
	return nil
}

// reloadConfig hotreloads the config on receiving the config update
func reloadConfig(ctx context.Context, reloadCh <-chan bool, reloadDone chan<- bool, lmconfigFilePath string) {
	logger.Info("Checking for reload request")
	for {
		select {
		case <-reloadCh:
			logger.Info("Reloading the config")
			err := lmk8swebhookconfig.LoadConfig(lmconfigFilePath)
			if err != nil {
				logger.Error(err, "Error while loading the config file", "lmconfigFilePath", lmconfigFilePath)
			}
			reloadDone <- true
		case <-ctx.Done():
			logger.Error(ctx.Err(), "reload config is shut down")
			return
		default:
			continue
		}
	}
}
