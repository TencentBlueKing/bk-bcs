package config

// BCSConf :
type BCSConf struct {
	Host   string `yaml:"host"`
	Token  string `yaml:"token"`
	Verify bool   `yaml:"verify"`
}

// Init : init default redis config
func (c *BCSConf) Init() error {
	// only for development
	c.Host = ""
	c.Token = ""
	c.Verify = false
	return nil
}
