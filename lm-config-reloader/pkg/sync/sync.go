package sync

import (
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/fetcher"
)

// ConfigSyncer needs to be implemented by the config syncer of the resource
type ConfigSyncer interface {
	Sync(response *fetcher.Response) error
}

// K8sResource needs to be implemented by the resource kind
type K8sResource interface {
	validate() error
}
