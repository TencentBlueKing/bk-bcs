package redisclient

import "time"

// Config 初始化 Redis 客户端需要使用的
type Config struct {
	Addrs      []string  // 节点列表
	MasterName string    // 哨兵模式下的主节点名
	Password   string    // 密码
	DB         int       // 单节点模式下的数据库
	Mode       RedisMode // single, sentinel, cluster

	// Options configs
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
	MinIdleConns int
	IdleTimeout  time.Duration
}
