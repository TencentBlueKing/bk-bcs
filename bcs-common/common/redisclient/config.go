package redisclient

import "time"

// Config contains the configuration required to initialize a Redis client
type Config struct {
	Addrs      []string  // List of nodes (addresses)
	MasterName string    // Master node name in Sentinel mode
	Password   string    // Password
	DB         int       // Database index in single node mode
	Mode       RedisMode // Redis mode: single, sentinel, or cluster

	// Options configs
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
	MinIdleConns int
	IdleTimeout  time.Duration
}
