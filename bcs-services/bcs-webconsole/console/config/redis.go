package config

// RedisConf :
type RedisConf struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Password       string `yaml:"password"`
	DB             int    `yaml:"db"`
	MaxPoolSize    int    `yaml:"max_pool_size"`
	MaxConnTimeout int    `yaml:"max_conn_timeout"`
	IdleTimeout    int    `yaml:"idle_timeout"`
	ReadTimeout    int    `yaml:"read_timeout"`
	WriteTimeout   int    `yaml:"write_timeout"`
}

// Init : init default redis config
func (c *RedisConf) Init() error {
	// only for development
	c.Host = "127.0.0.1"
	c.Port = 6379
	c.Password = ""
	c.DB = 0

	c.MaxPoolSize = 100
	c.MaxConnTimeout = 6
	c.IdleTimeout = 600
	c.ReadTimeout = 10
	c.WriteTimeout = 10
	return nil
}
