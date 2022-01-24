package sync

import (
	"fmt"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/config"
)

const (
	// ConfigMapResourceKind represents the ConfigMap resource kind
	ConfigMapResourceKind = "ConfigMap"
	// MutatingWebhookConfigurationKind represents the MutatingWebhookConfiguration kind
	MutatingWebhookConfigurationKind = "MutatingWebhookConfiguration"
)

// CreateConfigSyncer creates ConfigSyncer for the specified resource kind
func CreateConfigSyncer(reloader config.Reloader, k8sClient *config.K8sClient) (ConfigSyncer, error) {
	kind := reloader.Resource["kind"].(string)
	switch kind {
	case ConfigMapResourceKind:
		configMapResource, err := BuildConfigMapResource(reloader.Resource)
		if err != nil {
			return nil, err
		}
		return configMapConfigSyncer{Resource: configMapResource, ReloaderEndpoint: reloader.ReloadEndpoint, k8sClient: k8sClient}, nil

	case MutatingWebhookConfigurationKind:
		mutatingWebhookConfigRes, err := BuildMutatingWebhookConfigurationResource(reloader.Resource)
		if err != nil {
			return nil, err
		}
		return mutatingWebhookConfigSyncer{Resource: mutatingWebhookConfigRes, k8sClient: k8sClient}, nil

	default:
		return nil, fmt.Errorf("invalid resource kind %s provided", kind)
	}
}
