package config

// AuthConf :
type AuthConf struct {
	Host       string `yaml:"host"`       // api 地址 可以内部地址或者网关地址
	IsGatewWay bool   `yaml:"is_gateway"` // 是否是网关地址
	SSMHost    string `yaml:"ssm_host"`   // 获取 access_token 的地址
}

// Init : init default redis config
func (c *AuthConf) Init() error {
	// only for development
	c.Host = ""
	c.SSMHost = ""
	return nil
}
