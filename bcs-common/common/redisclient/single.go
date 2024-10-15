package redisclient

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

// SingleClient Redis client for single mode
type SingleClient struct {
	cli *redis.Client
}

// NewSingleClient init SingleClient from config
func NewSingleClient(config Config) (*SingleClient, error) {
	if config.Mode != SingleMode {
		return nil, errors.New("redis mode not supported")
	}
	if len(config.Addrs) == 0 {
		return nil, errors.New("address is empty")
	}
	cli := redis.NewClient(&redis.Options{
		Addr:         config.Addrs[0],
		Password:     config.Password,
		DB:           config.DB,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		IdleTimeout:  config.IdleTimeout,
	})
	return &SingleClient{cli: cli}, nil
}

// NewSingleClientFromDSN init SingleClient by dsn
func NewSingleClientFromDSN(dsn string) (*SingleClient, error) {
	options, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, err
	}
	cli := redis.NewClient(options)
	return &SingleClient{cli: cli}, nil
}

func (c *SingleClient) GetCli() redis.UniversalClient {
	return c.cli
}

func (c *SingleClient) Ping(ctx context.Context) (string, error) {
	return c.cli.Ping(ctx).Result()
}

func (c *SingleClient) Get(ctx context.Context, key string) (string, error) {
	return c.cli.Get(ctx, key).Result()
}

func (c *SingleClient) Set(ctx context.Context, key string, value interface{}, duration time.Duration) (string, error) {
	return c.cli.Set(ctx, key, value, duration).Result()
}

func (c *SingleClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.cli.SetNX(ctx, key, value, expiration).Result()
}

func (c *SingleClient) SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) (string, error) {
	return c.cli.SetEX(ctx, key, value, expiration).Result()
}

func (c *SingleClient) Exists(ctx context.Context, key ...string) (int64, error) {
	return c.cli.Exists(ctx, key...).Result()
}

func (c *SingleClient) Del(ctx context.Context, key string) (int64, error) {
	return c.cli.Del(ctx, key).Result()
}

func (c *SingleClient) Expire(ctx context.Context, key string, duration time.Duration) (bool, error) {
	return c.cli.Expire(ctx, key, duration).Result()
}
