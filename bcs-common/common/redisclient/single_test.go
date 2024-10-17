package redisclient

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func setupSingleClient() Client {
	// Create Redis single instance configuration
	config := Config{
		Mode:  SingleMode,
		Addrs: []string{"127.0.0.1:6379"},
		DB:    0,
	}

	// Initialize Redis client
	client, _ := NewClient(config)
	return client
}

// Test for Ping command
func TestPing(t *testing.T) {
	client := setupSingleClient()
	result, err := client.GetCli().Ping(context.TODO()).Result()
	assert.NoError(t, err)
	assert.Equal(t, result, "PONG")
}

// Test basic functionalities of SingleClient
func TestSingleClient(t *testing.T) {
	client := setupSingleClient()
	assert.NotNil(t, client)

	ctx := context.Background()

	// Test Get operation
	_, err := client.Set(ctx, "key1", "value1", 10*time.Second)
	assert.NoError(t, err)

	// Test Get operation
	val, err := client.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Test key existence
	exists, err := client.Exists(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), exists)

	// Test Del operation
	_, err = client.Del(ctx, "key1")
	assert.NoError(t, err)

	// Test whether the key is deleted
	exists, err = client.Exists(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), exists)

}

// Test SetEX and SetNX operations
func TestSingleClientSetEXAndSetNX(t *testing.T) {
	client := setupSingleClient()
	assert.NotNil(t, client)

	ctx := context.Background()

	// Test SetEX operation by setting a key with expiration time
	_, err := client.SetEX(ctx, "key2", "value2", 5*time.Second)
	assert.NoError(t, err)

	// Get key2 and verify the value
	val, err := client.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.Equal(t, "value2", val)

	// Confirm that key2 exists in Redis
	exists, err := client.Exists(ctx, "key2")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), exists)

	// Wait for the expiration time and check if the key still exists
	time.Sleep(6 * time.Second)
	exists, err = client.Exists(ctx, "key2")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), exists)

	// Test SetNX operation, which sets the key only if it does not exist
	success, err := client.SetNX(ctx, "key3", "value3", 10*time.Second)
	assert.NoError(t, err)
	assert.True(t, success)

	// Try SetNX again, should return false as the key already exists
	success, err = client.SetNX(ctx, "key3", "value3", 10*time.Second)
	assert.NoError(t, err)
	assert.False(t, success)

	// Delete key3
	_, err = client.Del(ctx, "key3")
	assert.NoError(t, err)
}

// Test edge cases, such as retrieving a non-existent key
func TestSingleClientGetNonExistentKey(t *testing.T) {
	client := setupSingleClient()
	assert.NotNil(t, client)

	ctx := context.Background()

	// Test retrieving a non-existent key, should return an empty string and redis.Nil error
	val, err := client.Get(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, redis.Nil, err)
	assert.Equal(t, "", val)
}
