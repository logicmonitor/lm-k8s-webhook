package main

import (
	"context"
	"flag"
	"os"
	"sync"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/internal/version"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/config"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/logger"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/reloader"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/watcher"

	goruntime "runtime"

	"go.uber.org/zap"
	"k8s.io/client-go/rest"
)

var (
	lmReloaderFilePath string
	logLevel           string
)

func main() {
	flag.StringVar(&lmReloaderFilePath, "lmreloader-file-path", "/etc/lmreloader/config/lmreloaderconfig.yaml", "File path of lmreloader")
	flag.StringVar(&logLevel, "log-level", "debug", "log level")
	flag.Parse()

	var reloaderCfg *config.ReloaderConfig
	var k8sRestConfig *rest.Config
	var err error

	// initialize the logger
	if err := logger.Init(logLevel); err != nil {
		panic(err)
	}

	v := version.Get()

	logger.Logger().Info("Starting the LM-Config-Reloader",
		zap.String("lm-config-reloader-version", v.LMConfigReloader),
		zap.String("build-date", v.BuildDate),
		zap.String("go-version", v.Go),
		zap.String("go-arch", goruntime.GOARCH),
		zap.String("go-os", goruntime.GOOS),
	)

	// load config
	reloaderCfg, err = config.LoadConfig(lmReloaderFilePath)
	if err != nil {
		logger.Logger().Error("error in loading config file", zap.Error(err))
		os.Exit(1)
	}

	// validate config
	if err = reloader.Validate(reloaderCfg); err != nil {
		logger.Logger().Error("error in validating config", zap.Error(err))
		os.Exit(1)
	}

	// create K8s client
	if k8sRestConfig, err = config.K8sRestConfig(); err != nil {
		logger.Logger().Error("error in getting K8sRestConfig", zap.Error(err))
		os.Exit(1)
	}

	k8sClient, err := config.NewK8sClient(k8sRestConfig, config.NewK8sClientSet)
	if err != nil {
		logger.Logger().Error("error in getting k8s client", zap.Error(err))
		os.Exit(1)
	}

	var wg sync.WaitGroup

	lmReloader := &reloader.LMReloader{
		ReloaderConfig: reloaderCfg,
		Watcher: watcher.RemoteConfigWatcher{
			K8sClient: k8sClient,
			Wg:        &wg,
		},
		K8sClient: k8sClient,
		Wg:        &wg,
	}

	// setup reloaders
	if err = lmReloader.SetupProviders(context.Background()); err != nil {
		logger.Logger().Error("error in setup of the config providers", zap.Error(err))
		os.Exit(1)
	}
	logger.Logger().Info("exiting the main process")
}
