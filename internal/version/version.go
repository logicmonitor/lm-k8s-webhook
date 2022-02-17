package version

import (
	"fmt"
	"runtime"
)

var (
	lmK8sWebhook string
	buildDate    string
)

// Version holds the version of LMK8sWebhook
type Version struct {
	LMK8sWebhook string
	BuildDate    string
	Go           string
}

// Get returns the Version information
func Get() Version {
	return Version{
		LMK8sWebhook: LMK8sWebhook(),
		BuildDate:    buildDate,
		Go:           runtime.Version(),
	}
}

// LMK8sWebhook returns lm-k8s-webhook's version
func LMK8sWebhook() string {
	if len(lmK8sWebhook) != 0 {
		return lmK8sWebhook
	}
	return "0.0.0"
}

func (v Version) String() string {
	return fmt.Sprintf("Version (LMK8sWebhook=%s, BuildDate=%s)", v.LMK8sWebhook, v.BuildDate)
}
