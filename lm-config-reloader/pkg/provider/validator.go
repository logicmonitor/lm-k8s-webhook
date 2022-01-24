package provider

import (
	"fmt"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/config"

	"github.com/mitchellh/mapstructure"
)

// ValidateProviderConfig calls the provider's validate functions depending upon the providers configured
func ValidateProviderConfig(provider config.Provider) error {
	if provider.Git != nil {
		var gitProviderObj GitProvider
		err := mapstructure.Decode(provider.Git, &gitProviderObj)
		if err != nil {
			return err
		}
		if err = validateGitProviderConfig(gitProviderObj); err != nil {
			return err
		}
	}
	return nil
}

// validateGitProviderConfig validates git provider config
func validateGitProviderConfig(gitProvider GitProvider) error {
	if gitProvider.Git.Owner == "" {
		return fmt.Errorf("repo owner must not be empty")
	}
	if gitProvider.Git.Repo == "" {
		return fmt.Errorf("repo name must not be empty")
	}
	if gitProvider.Git.FilePath == "" {
		return fmt.Errorf("file path must not be empty")
	}
	if gitProvider.Git.AuthRequired && gitProvider.Git.AccessToken == "" {
		return fmt.Errorf("auth token must not be empty")
	}
	if _, err := gitProvider.GetParsedPullInterval(); err != nil {
		return err
	}
	return nil
}
