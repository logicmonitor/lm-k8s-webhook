package sync

import (
	"context"
	"reflect"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/config"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/fetcher"
	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/logger"

	"github.com/ghodss/yaml"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MutatingWebhookConfigurationResource represents MutatingWebhookConfiguration Resource
type MutatingWebhookConfigurationResource struct {
	Name string
}

type mutatingWebhookConfigSyncer struct {
	Resource  MutatingWebhookConfigurationResource
	k8sClient *config.K8sClient
}

// BuildMutatingWebhookConfigurationResource builds MutatingWebhookConfigurationResource
func BuildMutatingWebhookConfigurationResource(resource config.Resource) (MutatingWebhookConfigurationResource, error) {
	var mutatingWebhookConfigResource MutatingWebhookConfigurationResource
	err := mapstructure.Decode(resource, &mutatingWebhookConfigResource)
	if err != nil {
		return mutatingWebhookConfigResource, err
	}
	return mutatingWebhookConfigResource, nil
}

// Sync compares the content from the MutatingWebhookConfiguration with the config from config providers
// and if not matched then updates the MutatingWebhookConfiguration
func (mutatingWebhookConfigSyncer mutatingWebhookConfigSyncer) Sync(response *fetcher.Response) error {
	mutatingWebhookConfig, err := mutatingWebhookConfigSyncer.k8sClient.Clientset.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(context.Background(), mutatingWebhookConfigSyncer.Resource.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// Updated config
	updatedConfig := admissionregistrationv1.MutatingWebhookConfiguration{}
	err = yaml.Unmarshal(response.FileData, &updatedConfig)
	if err != nil {
		return err
	}
	var updateRequired bool
	// Check for objectSelector
	if !reflect.DeepEqual(updatedConfig.Webhooks[0].ObjectSelector, mutatingWebhookConfig.Webhooks[0].ObjectSelector) {
		mutatingWebhookConfig.Webhooks[0].ObjectSelector = updatedConfig.Webhooks[0].ObjectSelector
		updateRequired = true
		logger.Logger().Debug("Change is detected in the objectSelector", zap.Any("updated objectSelector", updatedConfig.Webhooks[0].ObjectSelector))
	}
	if !reflect.DeepEqual(updatedConfig.Webhooks[0].NamespaceSelector, mutatingWebhookConfig.Webhooks[0].NamespaceSelector) {
		mutatingWebhookConfig.Webhooks[0].NamespaceSelector = updatedConfig.Webhooks[0].NamespaceSelector
		updateRequired = true
		logger.Logger().Debug("Change is detected in the namespaceSelector", zap.Any("updated namespaceSelector", updatedConfig.Webhooks[0].NamespaceSelector))
	}
	if updateRequired {
		_, err = mutatingWebhookConfigSyncer.k8sClient.Clientset.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(context.Background(), mutatingWebhookConfig, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		logger.Logger().Info("MutatingWebhookConfiguration is updated", zap.String("mutatingWebhookConfig name", mutatingWebhookConfig.Name))
	} else {
		logger.Logger().Info("MutatingWebhookConfiguration content is matched, no change is detected", zap.String("mutatingWebhookConfig name", mutatingWebhookConfig.Name))
	}
	return nil
}
