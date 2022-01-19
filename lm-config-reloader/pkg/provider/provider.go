package provider

import (
	"time"

	"github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/pkg/fetcher"
)

// RemoteProvider is the interface to be implmented by the remote config providers
type RemoteProvider interface {
	fetcher.Fetcher
	GetParsedPullInterval() (time.Duration, error)
	GetPullInterval() string
	SetPullInterval(string)
}

// DefaultRemoteProvider holds basic providers data
type DefaultRemoteProvider struct {
	Provider     string
	PullInterval string
}

// GetParsedPullInterval returns the parsed duration
func (rp *DefaultRemoteProvider) GetParsedPullInterval() (time.Duration, error) {
	// If pull interval is not provided
	if rp.PullInterval == "" {
		return 0, nil
	}
	return time.ParseDuration(rp.PullInterval)
}

// SetPullInterval sets the pullInterval
func (rp *DefaultRemoteProvider) SetPullInterval(pullInterval string) {
	rp.PullInterval = pullInterval
}

// GetPullInterval returns the pullInterval assigned to remote provider
func (rp *DefaultRemoteProvider) GetPullInterval() string {
	return rp.PullInterval
}
