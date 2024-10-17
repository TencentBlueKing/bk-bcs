package redisclient

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// setupClusterClient function for initializing Redis SentinelClient
func setupSentinel(t *testing.T) *SentinelClient {
	config := Config{
		Mode:       SentinelMode,
		Addrs:      []string{"127.0.0.1:5001"}, // Sentinel addresses
		MasterName: "mymaster",                 // Master name
		DB:         0,
		Password:   "",
	}
	client, err := NewSentinelClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	return client
}

// TestSentinelClientPing tests SentinelClient connectivity
func TestSentinelClientPing(t *testing.T) {
	client := setupSentinel(t)
	result, err := client.GetCli().Ping(context.TODO()).Result()
	assert.NoError(t, err)
	assert.Equal(t, "PONG", result)
}

// TestSentinelClient tests SentinelClient basic functionality
func TestSentinelClient(t *testing.T) {
	client := setupSentinel(t)
	ctx := context.Background()

	// Test Set operation
	_, err := client.Set(ctx, "key1", "value1", 10*time.Second)
	assert.NoError(t, err)

	// Test Get operation
	val, err := client.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Test Exists operation
	exists, err := client.Exists(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), exists)

	// Test Del operation
	_, err = client.Del(ctx, "key1")
	assert.NoError(t, err)

	// Test if the key has been deleted
	exists, err = client.Exists(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), exists)
}
