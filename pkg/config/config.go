package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	logr "sigs.k8s.io/controller-runtime/pkg/log"
)

// Config holds the external configuration
type Config struct {
	LMEnvVars LMEnvVars `yaml:"lmEnvVars"`
}

type LMEnvVars struct {
	/* Resource holds the resource environment variables,
	which will be the part of OTEL_RESOURCE_ATTRIBUTES
	*/
	Resource []corev1.EnvVar `yaml:"resource,omitempty"`

	/* Operation holds the operation environment variables,
	which will not be the part of OTEL_RESOURCE_ATTRIBUTES.
	*/
	Operation []corev1.EnvVar `yaml:"operation,omitempty"`
}

// LoadConfig loads the external config passed by the user
func LoadConfig(configFilePath string) (*Config, error) {
	logger := logr.Log.WithName(("load-config"))

	// Check if config file provided
	if _, err := os.Stat(filepath.Clean(configFilePath)); os.IsNotExist(err) {
		// As external config is optional
		logger.Info("Config file is not provided")
		return nil, err
	}

	data, err := ioutil.ReadFile(filepath.Clean(configFilePath))
	if err != nil {
		logger.Error(err, "Error in reading the config file", "configFilePath", configFilePath)
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		logger.Error(err, "Error in reading the config file", "configFilePath", configFilePath)
		return nil, err
	}
	return &cfg, nil
}
