package config

// IAMConf :
type IAMConf struct {
	// Endpoints is a seed list of host:port addresses of iam nodes.
	Host string `yaml:"host"`
}

// Init : init default redis config
func (c *IAMConf) Init() error {
	c.Host = ""
	return nil
}
