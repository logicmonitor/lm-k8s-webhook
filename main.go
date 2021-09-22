package main

import (
	"flag"
	lmconfig "lm-webhook/pkg/config"
	"lm-webhook/pkg/handler"
	"os"
	"strconv"

	"sigs.k8s.io/controller-runtime/pkg/healthz"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
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

	flag.StringVar(&metricAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&webhookPort, "webhook-bind-port", "9443", "The port webhook will listen on.")
	flag.StringVar(&webhookCertDir, "webhook-cert-dir", "/etc/lmwebhook/certs", "webhook certificate directory.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&lmconfigFilePath, "lmconfig-file-path", "/etc/lmwebhook/config/lmconfig.yaml", "File path of lmconfig")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog.Info("setting up manager")

	port, err := strconv.Atoi(webhookPort)

	if err != nil {
		setupLog.Error(err, "Failed in parsing webhook port")
		os.Exit(1)
	}

	// Load the external config

	cfg, err := lmconfig.LoadConfig(lmconfigFilePath)

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

	mgr, err := manager.New(config.GetConfigOrDie(), mgrOptions)

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

	setupLog.Info("registering webhooks to the webhook server")
	lmWebhookServer.Register("/mutate", &webhook.Admission{Handler: &handler.LMPodMutator{Client: mgr.GetClient(), Log: ctrl.Log.WithName("lm-podmutator-webhook"), LMConfig: cfg}})

	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "unable to run manager")
		os.Exit(1)
	}
}