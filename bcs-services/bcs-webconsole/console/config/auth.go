package config

// AuthConf :
type AuthConf struct {
	Host       string `yaml:"host"`
	Env        string `yaml:"env"`
	IAMVersion string `yaml:"iam_version"` // 权限中心版本 v2 or v3
	SSMHost    string `yaml:"ssm_host"`    // v3时 ssmHOST地址
}

// Init : init default redis config
func (c *AuthConf) Init() error {
	// only for development
	c.Host = ""
	c.Env = ""
	c.IAMVersion = "v2"
	c.SSMHost = ""
	return nil
}
