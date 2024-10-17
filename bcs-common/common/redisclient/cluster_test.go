package redisclient

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// setupClusterClient function for initializing Redis ClusterClient
func setupClusterClient(t *testing.T) *ClusterClient {
	config := Config{
		Mode:  ClusterMode,
		Addrs: []string{"127.0.0.1:7021", "127.0.0.1:7022", "127.0.0.1:7023"},
	}
	client, err := NewClusterClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	return client
}

// TestClusterPing tests ClusterClient connectivity
func TestClusterPing(t *testing.T) {
	client := setupClusterClient(t)
	result, err := client.GetCli().Ping(context.TODO()).Result()
	assert.NoError(t, err)
	assert.Equal(t, "PONG", result)
}

// TestClusterClient tests ClusterClient basic functionality
func TestClusterClient(t *testing.T) {
	client := setupClusterClient(t)
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
