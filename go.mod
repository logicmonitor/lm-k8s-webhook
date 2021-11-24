module github.com/logicmonitor/lm-k8s-webhook

go 1.16

require (
	github.com/fsnotify/fsnotify v1.5.1
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.4.0
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.3.0 // indirect
	github.com/spf13/viper v1.9.0
	go.opentelemetry.io/otel v0.20.0
	golang.org/x/sys v0.0.0-20210921065528-437939a70204 // indirect
	golang.org/x/tools v0.1.6 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	sigs.k8s.io/controller-runtime v0.9.2
)
