package config

// BKLoginConf 配置
type BKLoginConf struct {
	Host string `yaml:"host"`
}

// Init
func (c *BKLoginConf) Init() error {
	c.Host = ""
	return nil
}
