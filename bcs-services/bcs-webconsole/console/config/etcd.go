package config

// EtcdConf xxx
type EtcdConf struct {
	Endpoints string `yaml:"endpoints"`
	Ca        string `yaml:"ca"`
	Cert      string `yaml:"cert"`
	Key       string `yaml:"key"`
}

// Init xxx
func (c *EtcdConf) Init() error {
	c.Endpoints = ""
	c.Cert = ""
	c.Ca = ""
	c.Key = ""
	return nil
}
