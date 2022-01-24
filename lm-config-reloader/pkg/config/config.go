package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Resource holds resource config
type Resource map[string]interface{}

// ReloaderConfig holds slice of reloaders
type ReloaderConfig struct {
	Reloaders []Reloader `yaml:"reloaders"`
}

// Reloader holds reloader config
type Reloader struct {
	Provider       Provider `yaml:"configProvider"`
	Resource       Resource `yaml:"resource"`
	ReloadEndpoint string   `yaml:"reloadEndpoint"`
}

// Provider holds the config of all the supported config providers
type Provider struct {
	Git *Git `yaml:"git"`
}

// LoadConfig loads the reloaders config from the config file
func LoadConfig(configFilePath string) (*ReloaderConfig, error) {
	if _, err := os.Stat(filepath.Clean(configFilePath)); os.IsNotExist(err) {
		return nil, err
	}
	data, err := ioutil.ReadFile(filepath.Clean(configFilePath))
	if err != nil {
		return nil, err
	}
	var reloaderCfg ReloaderConfig
	if err := yaml.Unmarshal(data, &reloaderCfg); err != nil {
		return nil, err
	}
	return &reloaderCfg, nil
}
