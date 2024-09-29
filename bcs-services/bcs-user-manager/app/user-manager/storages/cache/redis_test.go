package cache

import (
	"context"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
	"github.com/stretchr/testify/assert"
)

var testRDB Cache

func setup() error {
	conf := &config.UserMgrConfig{
		RedisConfig: config.RedisConfig{
			Addr:      "127.0.0.1:7021,127.0.0.1:7022,127.0.0.1:7023,127.0.0.1:7024,127.0.0.1:7025,127.0.0.1:7026",
			RedisMode: "cluster",
		},
	}
	err := InitRedis(conf)
	testRDB = RDB
	return err
}

func TestInitRedis(t *testing.T) {
	err := setup()
	assert.Nil(t, err)
	assert.NotNil(t, testRDB)
}

func TestRedisCache_Set(t *testing.T) {
	err := setup()
	assert.Nil(t, err)

	key := "test:key"
	value := "testValue"
	expiration := time.Second * 10

	result, err := testRDB.Set(context.Background(), key, value, expiration)
	assert.NoError(t, err)
	assert.Equal(t, "OK", result)

	// 验证值是否正确设置
	val, err := testRDB.Get(context.Background(), key)
	assert.NoError(t, err)
	assert.Equal(t, value, val)
}

func TestRedisCache_Get(t *testing.T) {
	err := setup()
	assert.Nil(t, err)
	key := "test:key"
	value := "testValue"
	_, err = testRDB.Set(context.Background(), key, value, 0)
	assert.NoError(t, err)

	result, err := testRDB.Get(context.Background(), key)
	assert.NoError(t, err)
	assert.Equal(t, value, result)
}

func TestRedisCache_Del(t *testing.T) {
	err := setup()
	assert.Nil(t, err)
	key := "test:key"
	_, err = testRDB.Set(context.Background(), key, "value", 0)
	assert.NoError(t, err)

	count, err := testRDB.Del(context.Background(), key)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), count)

	// 验证值是否已删除
	val, err := testRDB.Get(context.Background(), key)
	assert.Error(t, err)
	assert.Equal(t, "", val)
}

func TestRedisCache_Expire(t *testing.T) {
	err := setup()
	assert.Nil(t, err)

	key := "test:key"
	value := "testValue"
	_, err = testRDB.Set(context.Background(), key, value, 0)
	assert.NoError(t, err)

	// 设置过期时间
	expired, err := testRDB.Expire(context.Background(), key, time.Second*1)
	assert.NoError(t, err)
	assert.True(t, expired)

	// 等待过期
	time.Sleep(2 * time.Second)

	val, err := testRDB.Get(context.Background(), key)
	assert.Error(t, err)
	assert.Equal(t, "", val)
}
