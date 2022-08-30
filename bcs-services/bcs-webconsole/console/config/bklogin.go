package config

// BKLoginConf 配置
type BKLoginConf struct {
	Host string `yaml:"host"`
}

// Init xxx
func (c *BKLoginConf) Init() error {
	c.Host = ""
	return nil
}
