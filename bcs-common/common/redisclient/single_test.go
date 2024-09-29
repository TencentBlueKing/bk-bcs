package redisclient

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func setupSingleClient() Client {
	// 创建 Redis 单机配置
	config := Config{
		Mode:  SingleMode,
		Addrs: []string{"127.0.0.1:6379"},
		DB:    0,
	}

	// 初始化 Redis 客户端，使用可选参数设置连接超时
	client, _ := NewClient(config)
	return client
}

// Ping 测试
func TestPing(t *testing.T) {
	client := setupSingleClient()
	result, err := client.GetCli().Ping(context.TODO()).Result()
	assert.NoError(t, err)
	assert.Equal(t, result, "PONG")
}

// 测试 SingleClient 基础功能
func TestSingleClient(t *testing.T) {
	client := setupSingleClient()
	assert.NotNil(t, client)

	ctx := context.Background()

	// 测试 Set 操作
	_, err := client.Set(ctx, "key1", "value1", 10*time.Second)
	assert.NoError(t, err)

	// 测试 Get 操作
	val, err := client.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	// 测试键存在性
	exists, err := client.Exists(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), exists)

	// 测试 Del 操作
	_, err = client.Del(ctx, "key1")
	assert.NoError(t, err)

	// 测试键是否已删除
	exists, err = client.Exists(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), exists)

}

// 测试 SetEX 和 SetNX 操作
func TestSingleClientSetEXAndSetNX(t *testing.T) {
	client := setupSingleClient()
	assert.NotNil(t, client)

	ctx := context.Background()

	// 测试 SetEX 操作，设置带有过期时间的键
	_, err := client.SetEX(ctx, "key2", "value2", 5*time.Second)
	assert.NoError(t, err)

	// 获取 key2，确保值正确
	val, err := client.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.Equal(t, "value2", val)

	// 确认 key2 在 Redis 中存在
	exists, err := client.Exists(ctx, "key2")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), exists)

	// 等待过期时间后检查键是否存在
	time.Sleep(6 * time.Second)
	exists, err = client.Exists(ctx, "key2")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), exists)

	// 测试 SetNX 操作，只有当键不存在时才能设置
	success, err := client.SetNX(ctx, "key3", "value3", 10*time.Second)
	assert.NoError(t, err)
	assert.True(t, success)

	// 再次尝试 SetNX 操作，这次应该返回 false，因为键已经存在
	success, err = client.SetNX(ctx, "key3", "value3", 10*time.Second)
	assert.NoError(t, err)
	assert.False(t, success)

	// 删除 key3
	_, err = client.Del(ctx, "key3")
	assert.NoError(t, err)
}

// 测试边界情况，例如不存在的键
func TestSingleClientGetNonExistentKey(t *testing.T) {
	client := setupSingleClient()
	assert.NotNil(t, client)

	ctx := context.Background()

	// 测试获取不存在的键，应该返回空字符串和 redis.Nil 错误
	val, err := client.Get(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, redis.Nil, err)
	assert.Equal(t, "", val)
}
