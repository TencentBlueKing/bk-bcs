package redisclient

import (
	"context"
	"fmt"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"
)

type RedisMode string

const (
	SingleMode   RedisMode = "single"   // Single mode
	SentinelMode RedisMode = "sentinel" // Sentinel mode
	ClusterMode  RedisMode = "cluster"  // Cluster mode
)

type Client interface {
	// GetCli return the underlying Redis client
	GetCli() redis.UniversalClient
	// Ping checks the Redis server connection
	Ping(ctx context.Context) (string, error)
	Get(ctx context.Context, key string) (string, error)
	Exists(ctx context.Context, key ...string) (int64, error)
	Set(ctx context.Context, key string, value interface{}, duration time.Duration) (string, error)
	SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) (string, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Del(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, duration time.Duration) (bool, error)
}

// NewClient creates a Redis client based on the configuration for different deployment modes
func NewClient(config Config) (Client, error) {
	switch config.Mode {
	case SingleMode:
		return NewSingleClient(config)
	case SentinelMode:
		return NewSentinelClient(config)
	case ClusterMode:
		return NewClusterClient(config)
	}
	return nil, fmt.Errorf("invalid config mode: %s", config.Mode)
}

// NewTestClient creates a Redis client for unit testing
func NewTestClient() (Client, error) {
	mr, err := miniredis.Run()
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return &SingleClient{cli: client}, nil
}
