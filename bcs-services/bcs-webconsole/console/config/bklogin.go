package config

type BKLoginConf struct {
	Host string `yaml:"host"`
}

func (c *BKLoginConf) Init() error {
	c.Host = ""
	return nil
}
