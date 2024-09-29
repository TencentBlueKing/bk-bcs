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
	SingleMode   RedisMode = "single"   // 单机模式
	SentinelMode RedisMode = "sentinel" // 哨兵模式
	ClusterMode  RedisMode = "cluster"  // 集群模式
)

type Client interface {
	GetCli() redis.UniversalClient
	Ping(ctx context.Context) (string, error)
	Get(ctx context.Context, key string) (string, error)
	Exists(ctx context.Context, key ...string) (int64, error)
	Set(ctx context.Context, key string, value interface{}, duration time.Duration) (string, error)
	SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) (string, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Del(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, duration time.Duration) (bool, error)
}

// NewClient 根据配置文件创建不同部署模式的 redis 客户端
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

// NewTestClient 创建用于单元测试的 redis 客户端
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
