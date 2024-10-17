package redisclient

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// ClusterClient Redis client for cluster mode
type ClusterClient struct {
	cli *redis.ClusterClient
}

// NewClusterClient init ClusterClient from config
func NewClusterClient(config Config) (*ClusterClient, error) {
	cli := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        config.Addrs,
		Password:     config.Password,
		DialTimeout:  config.DialTimeout * time.Second,
		ReadTimeout:  config.ReadTimeout * time.Second,
		WriteTimeout: config.WriteTimeout * time.Second,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		IdleTimeout:  config.IdleTimeout * time.Second,
	})
	return &ClusterClient{cli: cli}, nil
}

func (c *ClusterClient) GetCli() redis.UniversalClient {
	return c.cli
}

func (c *ClusterClient) Ping(ctx context.Context) (string, error) {
	return c.cli.Ping(ctx).Result()
}

func (c *ClusterClient) Get(ctx context.Context, key string) (string, error) {
	return c.cli.Get(ctx, key).Result()
}

func (c *ClusterClient) Exists(ctx context.Context, key ...string) (int64, error) {
	return c.cli.Exists(ctx, key...).Result()
}

func (c *ClusterClient) Set(ctx context.Context, key string, value interface{}, duration time.Duration) (string, error) {
	return c.cli.Set(ctx, key, value, duration).Result()
}

func (c *ClusterClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.cli.SetNX(ctx, key, value, expiration).Result()
}

func (c *ClusterClient) SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) (string, error) {
	return c.cli.SetEX(ctx, key, value, expiration).Result()
}

func (c *ClusterClient) Del(ctx context.Context, key string) (int64, error) {
	return c.cli.Del(ctx, key).Result()
}

func (c *ClusterClient) Expire(ctx context.Context, key string, duration time.Duration) (bool, error) {
	return c.cli.Expire(ctx, key, duration).Result()
}
