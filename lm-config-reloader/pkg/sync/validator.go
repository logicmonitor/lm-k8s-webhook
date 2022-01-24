package sync

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/config"
)

// ValidateResourceConfig calls the respectve resource config validator
func ValidateResourceConfig(resource config.Resource) error {
	kind := resource["kind"].(string)
	switch kind {

	case "ConfigMap":
		return buildAndValidateConfigmapResource(resource)

	case "MutatingWebhookConfiguration":
		return buildAndValidateMutatingWebhookConfigResource(resource)

	default:
		return fmt.Errorf("invalid resource kind %s provided", kind)
	}
}

// ValidateReloadEndpoint validates the reload endpoint
func ValidateReloadEndpoint(reloadEndpoint string) error {
	// TODO: check for scheme
	_, err := url.ParseRequestURI(reloadEndpoint)
	if err != nil {
		return err
	}
	return nil
}

// buildAndValidateConfigmapResource builds and validates configmap resource
func buildAndValidateConfigmapResource(resource config.Resource) error {
	cmResource, err := BuildConfigMapResource(resource)
	if err != nil {
		return err
	}
	return cmResource.validate()
}

// buildAndValidateMutatingWebhookConfigResource builds and validates MutatingWebhookConfigResource
func buildAndValidateMutatingWebhookConfigResource(resource config.Resource) error {
	mutatingWebhookConfigResource, err := BuildMutatingWebhookConfigurationResource(resource)
	if err != nil {
		return err
	}
	return mutatingWebhookConfigResource.validate()
}

// validate validates the ConfigMap Resource
func (cmResource ConfigMapResource) validate() error {
	// Check if name property is available
	if len(strings.TrimSpace(cmResource.Name)) == 0 {
		return fmt.Errorf("property name not found or empty in configmap resource config")
	}

	// Check if fileName property is available
	if len(strings.TrimSpace(cmResource.FileName)) == 0 {
		return fmt.Errorf("property fileName not found or empty in configmap resource config")
	}
	return nil
}

// validate validates the MutatingWebhookConfiguration Resource
func (mutatingWebhookConfigResource MutatingWebhookConfigurationResource) validate() error {
	// Check if name property is available
	if len(strings.TrimSpace(mutatingWebhookConfigResource.Name)) == 0 {
		return fmt.Errorf("property name not found or empty in configmap resource config")
	}
	return nil
}
