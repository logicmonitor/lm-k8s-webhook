package config

// Git holds config that can be used to connect to Git repo
type Git struct {
	Owner        string `yaml:"owner"`
	Repo         string `yaml:"repo"`
	Ref          string `yaml:"ref"`
	FilePath     string `yaml:"filePath"`
	AccessToken  string `yaml:"accessToken"`
	PullInterval string `yaml:"pullInterval"`
	AuthRequired bool   `yaml:"authRequired"`
	Disabled     bool   `yaml:"disabled"`
}
