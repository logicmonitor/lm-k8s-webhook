package main

import (
	"context"
	"flag"
	"os"
	"strconv"

	"github.com/logicmonitor/lm-k8s-webhook/internal/version"
	lmconfig "github.com/logicmonitor/lm-k8s-webhook/pkg/config"
	"github.com/logicmonitor/lm-k8s-webhook/pkg/handler"
	"github.com/logicmonitor/lm-k8s-webhook/pkg/reloader"

	"sigs.k8s.io/controller-runtime/pkg/healthz"

	goruntime "runtime"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("Setup")
)

const (
	webhookCertName = "tls.crt"
	webhookKeyName  = "tls.key"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricAddr string
	var webhookPort string
	var webhookCertDir string
	var probeAddr string
	var lmconfigFilePath string
	var k8sRestConfig *rest.Config

	flag.StringVar(&metricAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&webhookPort, "webhook-bind-port", "9443", "The port webhook will listen on.")
	flag.StringVar(&webhookCertDir, "webhook-cert-dir", "/etc/lmwebhook/certs", "webhook certificate directory.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&lmconfigFilePath, "lmconfig-file-path", "/etc/lmwebhook/config/lmconfig.yaml", "File path of lmconfig")

	var ctx context.Context
	ctx = context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := zap.Options{
		// Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)

	v := version.Get()

	logger.Info("Starting the LM-K8s-Webhook",
		"lm-k8s-webhook-version", v.LMK8sWebhook,
		"build-date", v.BuildDate,
		"go-version", v.Go,
		"go-arch", goruntime.GOARCH,
		"go-os", goruntime.GOOS,
	)

	setupLog.Info("setting up manager")

	port, err := strconv.Atoi(webhookPort)

	if err != nil {
		setupLog.Error(err, "Failed in parsing webhook port")
		os.Exit(1)
	}

	// Load the external config

	err = lmconfig.LoadConfig(lmconfigFilePath)

	if err != nil {
		// As external config is optional
		if os.IsNotExist(err) {
			setupLog.Info("Config file is not provided")
		} else {
			setupLog.Error(err, "Error in reading the config file", "lmconfigFilePath", lmconfigFilePath)
			os.Exit(1)
		}
	}

	mgrOptions := ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricAddr,
		Port:                   port,
		HealthProbeBindAddress: probeAddr,
	}

	k8sRestConfig = config.GetConfigOrDie()
	mgr, err := manager.New(k8sRestConfig, mgrOptions)

	if err != nil {
		setupLog.Error(err, "unable to set up overall LM webhook")
		os.Exit(1)
	}

	// Setup webhooks
	setupLog.Info("setting up webhook server")
	lmWebhookServer := mgr.GetWebhookServer()
	lmWebhookServer.CertDir = webhookCertDir
	lmWebhookServer.CertName = webhookCertName
	lmWebhookServer.KeyName = webhookKeyName

	k8sClient, err := lmconfig.NewK8sClient(k8sRestConfig, lmconfig.NewK8sClientSet)
	if err != nil {
		setupLog.Error(err, "error in getting k8s client")
		os.Exit(1)
	}

	setupLog.Info("registering webhooks to the webhook server")
	lmWebhookServer.Register("/mutate", &webhook.Admission{Handler: &handler.LMPodMutationHandler{Client: k8sClient, Log: ctrl.Log.WithName("lm-podmutator-webhook")}})

	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	if lmconfig.GetConfig().MutationConfigProvided {
		setupLog.Info("setup config reloader")
		err := reloader.SetupConfigReloader(ctx, lmconfigFilePath)
		if err != nil {
			setupLog.Error(err, "failed to setup config-reloader")
			os.Exit(1)
		}
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "unable to run manager")
		cancel()
		os.Exit(1)
	}
}
