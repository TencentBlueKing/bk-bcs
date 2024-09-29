package redisclient

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

type SentinelClient struct {
	cli *redis.Client
}

func NewSentinelClient(config Config) (*SentinelClient, error) {
	if config.Mode != SentinelMode {
		return nil, errors.New("redis mode not supported")
	}
	cli := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    config.MasterName,
		SentinelAddrs: config.Addrs,
		Password:      config.Password,
		DB:            config.DB,
		DialTimeout:   config.DialTimeout,
		ReadTimeout:   config.ReadTimeout,
		WriteTimeout:  config.WriteTimeout,
		PoolSize:      config.PoolSize,
		MinIdleConns:  config.MinIdleConns,
		IdleTimeout:   config.IdleTimeout,
	})
	return &SentinelClient{cli: cli}, nil
}

func (c *SentinelClient) GetCli() redis.UniversalClient {
	return c.cli
}

func (c *SentinelClient) Ping(ctx context.Context) (string, error) {
	return c.cli.Ping(ctx).Result()
}

func (c *SentinelClient) Set(ctx context.Context, key string, value interface{}, duration time.Duration) (string, error) {
	return c.cli.Set(ctx, key, value, duration).Result()
}

func (c *SentinelClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.cli.SetNX(ctx, key, value, expiration).Result()
}

func (c *SentinelClient) Get(ctx context.Context, key string) (string, error) {
	return c.cli.Get(ctx, key).Result()
}

func (c *SentinelClient) Exists(ctx context.Context, key ...string) (int64, error) {
	return c.cli.Exists(ctx, key...).Result()
}

func (c *SentinelClient) SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) (string, error) {
	return c.cli.SetEX(ctx, key, value, expiration).Result()
}

func (c *SentinelClient) Del(ctx context.Context, key string) (int64, error) {
	return c.cli.Del(ctx, key).Result()
}

func (c *SentinelClient) Expire(ctx context.Context, key string, duration time.Duration) (bool, error) {
	return c.cli.Expire(ctx, key, duration).Result()
}
