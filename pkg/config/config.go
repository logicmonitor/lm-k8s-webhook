package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	logr "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	configLock = new(sync.RWMutex)
	cfg        Config
	logger     = logr.Log.WithName(("config-loader"))
)

// Config holds the external configuration
type Config struct {
	MutationConfigProvided bool
	MutationConfig         MutationConfig
}

type MutationConfig struct {
	LMEnvVars LMEnvVars `yaml:"lmEnvVars"`
}

type LMEnvVars struct {
	/* Resource holds the resource environment variables,
	which will be the part of OTEL_RESOURCE_ATTRIBUTES
	*/
	Resource []ResourceEnv `yaml:"resource,omitempty"`

	/* Operation holds the operation environment variables,
	which will not be the part of OTEL_RESOURCE_ATTRIBUTES.
	*/
	Operation []OperationEnv `yaml:"operation,omitempty"`
}

type ResourceEnv struct {
	Env              corev1.EnvVar `yaml:"env"`
	ResAttrName      string        `yaml:"resAttrName,omitempty"`
	OverrideDisabled bool          `yaml:"overrideDisabled,omitempty"`
}

type OperationEnv struct {
	Env              corev1.EnvVar `yaml:"env"`
	ResAttrName      string        `yaml:"resAttrName,omitempty"`
	OverrideDisabled bool          `yaml:"overrideDisabled,omitempty"`
}

// LoadConfig loads the external config passed by the user
func LoadConfig(configFilePath string) error {
	logger = logr.Log.WithName(("load-config"))

	// Check if config file provided
	if _, err := os.Stat(filepath.Clean(configFilePath)); os.IsNotExist(err) {
		// As external config is optional
		logger.Info("Config file is not provided")
		return err
	}
	var tempCfg MutationConfig
	data, err := ioutil.ReadFile(filepath.Clean(configFilePath))
	if err != nil {
		logger.Error(err, "Error in reading the config file", "configFilePath", configFilePath)
		return err
	}
	if err := yaml.Unmarshal(data, &tempCfg); err != nil {
		logger.Error(err, "Error in reading the config file", "configFilePath", configFilePath)
		return err
	}

	configLock.Lock()
	cfg.MutationConfig = tempCfg
	cfg.MutationConfigProvided = true
	// logger.Info("Config:", "Config", cfg)
	configLock.Unlock()

	return nil
}

func GetConfig() Config {
	configLock.RLock()
	defer configLock.RUnlock()
	return cfg
}
