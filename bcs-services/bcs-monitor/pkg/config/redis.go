package config

const (
	RedisStandAloneType = "standalone" // 单节点redis
	RedisSentinelType   = "sentinel"   // 哨兵模式redis，哨兵实例
)

// RedisConf :
type RedisConf struct {
	Type             string   `yaml:"type" mapstructure:"type"`
	Host             string   `yaml:"host" mapstructure:"host"`
	Port             int      `yaml:"port" mapstructure:"port"`
	Password         string   `yaml:"password" mapstructure:"password"`
	DB               int      `yaml:"db" mapstructure:"db"`
	MasterName       string   `yaml:"master_name" mapstructure:"master_name"`
	SentinelAddrs    []string `yaml:"sentinel_addrs" mapstructure:"sentinel_addrs"`
	SentinelPassword string   `yaml:"sentinel_password" mapstructure:"sentinel_password"`
	MaxPoolSize      int      `yaml:"max_pool_size" mapstructure:"max_pool_size"`
	MaxConnTimeout   int      `yaml:"max_conn_timeout" mapstructure:"max_conn_timeout"`
	IdleTimeout      int      `yaml:"idle_timeout" mapstructure:"idle_timeout"`
	ReadTimeout      int      `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout     int      `yaml:"write_timeout" mapstructure:"write_timeout"`
}

func DefaultRedisConf() *RedisConf {
	// only for development
	return &RedisConf{
		Type:             RedisStandAloneType,
		Host:             "127.0.0.1",
		Port:             6379,
		Password:         "",
		DB:               0,
		MasterName:       "",
		SentinelAddrs:    []string{},
		SentinelPassword: "",
		MaxPoolSize:      100,
		MaxConnTimeout:   5,
		IdleTimeout:      600,
		ReadTimeout:      10,
		WriteTimeout:     10,
	}
}

// IsSentinelType 是否是哨兵模式
func (r *RedisConf) IsSentinelType() bool {
	return r.Type == RedisSentinelType
}
