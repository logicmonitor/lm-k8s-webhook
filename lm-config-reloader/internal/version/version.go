package version

import (
	"fmt"
	"runtime"
)

var (
	lmConfigReloader string
	buildDate        string
)

// Version holds the version of LMConfigReloader
type Version struct {
	LMConfigReloader string
	BuildDate        string
	Go               string
}

// Get returns the Version information
func Get() Version {
	return Version{
		LMConfigReloader: LMConfigReloader(),
		BuildDate:        buildDate,
		Go:               runtime.Version(),
	}
}

// LMConfigReloader returns lm-config-reloader's version
func LMConfigReloader() string {
	if len(lmConfigReloader) != 0 {
		return lmConfigReloader
	}
	return "0.0.0"
}

func (v Version) String() string {
	return fmt.Sprintf("Version (LMConfigReloader=%s, BuildDate=%s)", v.LMConfigReloader, v.BuildDate)
}
