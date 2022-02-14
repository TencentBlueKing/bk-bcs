package config

const (
	InternalMode = "internal" // inCluster 模式
	ExternalMode = "external" // 外部模式, 需要设置 AdminClusterId
)

// RedisConf :
type WebConsoleConf struct {
	Image          string `yaml:"image"`
	AdminClusterId string `yaml:"admin_cluster_id"`
	Mode           string `yaml:"mode"` // internal , external
}

// Init : init default WebConsoleConf config
func (c *WebConsoleConf) Init() error {
	// only for development
	c.Image = ""
	c.AdminClusterId = ""
	c.Mode = InternalMode

	return nil
}
