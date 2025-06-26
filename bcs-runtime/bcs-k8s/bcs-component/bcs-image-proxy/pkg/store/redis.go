/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package store

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"k8s.io/utils/strings/slices"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/internal/logctx"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/options"
)

// LayerType defines the layer type
type LayerType string

const (
	DOCKERD    LayerType = "DOCKERD"
	CONTAINERD LayerType = "CONTAINERD"
	TORRENT    LayerType = "TORRENT"
	StaticFile LayerType = "STATIC"
)

// LayerLocatedInfo defines the layer located info
type LayerLocatedInfo struct {
	Layer   string
	Type    LayerType
	Located string
	Data    string
}

// CacheStore defines the interface of cache store
type CacheStore interface {
	SaveOCILayer(ctx context.Context, ociType LayerType, layer, filePath string) error
	SaveStaticLayer(ctx context.Context, layer, filePath string, printLog bool) error
	SaveTorrent(ctx context.Context, layer, torrentBase64 string) error
	DeleteTorrent(ctx context.Context, layer string) error
	DeleteStaticLayer(ctx context.Context, layer string) error
	QueryTorrent(ctx context.Context, layer string) ([]*LayerLocatedInfo, error)
	QueryStaticLayer(ctx context.Context, layer string) ([]*LayerLocatedInfo, error)
	QueryOCILayer(ctx context.Context, layer string) ([]*LayerLocatedInfo, error)

	CleanHostCache(ctx context.Context) error

	AcquireLayerLock(ctx context.Context, layer string, afterTime time.Duration) (*RedisLock, error)
}

// RedisStore defines the redis store object
type RedisStore struct {
	op          *options.ImageProxyOption
	redisClient *redis.ClusterClient
}

var (
	globalRS *RedisStore
	syncOnce sync.Once
)

// NewRedisStore create the redis store instance
func NewRedisStore(op *options.ImageProxyOption) CacheStore {
	globalRS = &RedisStore{
		op: op,
		redisClient: redis.NewFailoverClusterClient(&redis.FailoverOptions{
			MasterName:    "mymaster",
			SentinelAddrs: strings.Split(op.RedisAddress, ","),
			Password:      op.RedisPassword,
		}),
	}
	return globalRS
}

// GlobalRedisStore returns the global redis store instance
func GlobalRedisStore() CacheStore {
	syncOnce.Do(func() {
		op := options.GlobalOptions()
		globalRS = &RedisStore{
			op: op,
			redisClient: redis.NewFailoverClusterClient(&redis.FailoverOptions{
				MasterName:    "mymaster",
				SentinelAddrs: strings.Split(op.RedisAddress, ","),
				Password:      op.RedisPassword,
			}),
		}
	})
	return globalRS
}

// AcquireLayerLock acquire the lock of layer
func (r *RedisStore) AcquireLayerLock(ctx context.Context, layer string, afterTime time.Duration) (*RedisLock, error) {
	<-time.After(afterTime)
	op := options.GlobalOptions()
	l := NewRedisLock(r.redisClient, "lock:"+layer, op.Address, 20*time.Second)
	ok, err := l.Acquire(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "lock digest '%s' failed", layer)
	}
	if ok {
		return l, nil
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			ok, err = l.Acquire(ctx)
			if err != nil {
				return nil, errors.Wrapf(err, "lock digest '%s' failed", layer)
			}
			if ok {
				return l, nil
			}
			logctx.Infof(ctx, "waiting for get the distribute layer lock...")
		case <-ctx.Done():
			return nil, errors.Errorf("context canceled")
		}
	}
}

func (r *RedisStore) buildLayerKey(layer string, ociType LayerType) string {
	return fmt.Sprintf("%s/%s/%s", layer, r.op.Address, string(ociType))
}

// SaveOCILayer save the dockerd/containerd layers with filepath
func (r *RedisStore) SaveOCILayer(ctx context.Context, ociType LayerType, layer, filePath string) error {
	rdk := r.buildLayerKey(layer, ociType)
	if err := r.redisClient.Set(ctx, rdk, filePath, 180*time.Second).Err(); err != nil {
		return errors.Wrapf(err, "redis set key '%s' with vaule '%s' failed", rdk, filePath)
	}
	logctx.Infof(ctx, "cache save oci layer '%s = %s' success", rdk, filePath)
	return nil
}

// SaveStaticLayer save static layer
func (r *RedisStore) SaveStaticLayer(ctx context.Context, layer string, filePath string, printLog bool) error {
	rdk := r.buildLayerKey(layer, StaticFile)
	if err := r.redisClient.Set(ctx, rdk, filePath, 180*time.Second).Err(); err != nil {
		return errors.Wrapf(err, "redis set key '%s' with vaule '%s' failed", rdk, filePath)
	}
	if printLog {
		logctx.Infof(ctx, "cache save static layer '%s = %s' success", rdk, filePath)
	}
	return nil
}

// SaveTorrent save torrent layer
func (r *RedisStore) SaveTorrent(ctx context.Context, layer string, torrentBase64 string) error {
	rdk := r.buildLayerKey(layer, TORRENT)
	if err := r.redisClient.Set(ctx, rdk, torrentBase64, 180*time.Second).Err(); err != nil {
		return errors.Wrapf(err, "redis set key '%s' with vaule '%s' failed", rdk, torrentBase64)
	}
	logctx.Infof(ctx, "cache save torrent layer '%s = (too long)' success", rdk)
	return nil
}

// DeleteTorrent delete torrent layer
func (r *RedisStore) DeleteTorrent(ctx context.Context, layer string) error {
	rdk := r.buildLayerKey(layer, TORRENT)
	if err := r.redisClient.Del(ctx, rdk).Err(); err != nil {
		return errors.Wrapf(err, "redis del key '%s' failed", rdk)
	}
	return nil
}

// DeleteStaticLayer delete static layer
func (r *RedisStore) DeleteStaticLayer(ctx context.Context, layer string) error {
	rdk := r.buildLayerKey(layer, StaticFile)
	if err := r.redisClient.Del(ctx, rdk).Err(); err != nil {
		return errors.Wrapf(err, "redis del key '%s' failed", rdk)
	}
	return nil
}

// CleanHostCache clean host cache
func (r *RedisStore) CleanHostCache(ctx context.Context) error {
	var keys []string
	var cursor uint64
	var resultKeys = make([]string, 0)
	for {
		var err error
		keys, cursor, err = r.redisClient.Scan(ctx, cursor, fmt.Sprintf("*/%s/*", r.op.Address), 50).Result()
		if err != nil {
			return errors.Wrapf(err, "redis clean host layers failed")
		}
		resultKeys = append(resultKeys, keys...)
		if cursor == 0 {
			break
		}
	}
	if len(resultKeys) == 0 {
		return nil
	}
	blog.Infof("clean host layers keys(%d): %s", len(resultKeys), strings.Join(resultKeys, ", "))
	if v, err := r.redisClient.Del(ctx, resultKeys...).Result(); err != nil {
		return errors.Wrapf(err, "redis clean host layers failed")
	} else {
		blog.Infof("clean host layers: %d", v)
	}
	return nil
}

// QueryTorrent query torrent layer
func (r *RedisStore) QueryTorrent(ctx context.Context, layer string) ([]*LayerLocatedInfo, error) {
	return r.commonQuery(ctx, layer, []string{string(TORRENT)})
}

// QueryStaticLayer query static layer
func (r *RedisStore) QueryStaticLayer(ctx context.Context, layer string) ([]*LayerLocatedInfo, error) {
	return r.commonQuery(ctx, layer, []string{string(StaticFile)})
}

// QueryOCILayer query oci layer
func (r *RedisStore) QueryOCILayer(ctx context.Context, layer string) ([]*LayerLocatedInfo, error) {
	return r.commonQuery(ctx, layer, []string{string(DOCKERD), string(CONTAINERD)})
}

func (r *RedisStore) commonQuery(ctx context.Context, layer string, keyTypes []string) ([]*LayerLocatedInfo, error) {
	var cursor uint64
	var keys []string
	var resultKeys []string
	for {
		var err error
		keys, cursor, err = r.redisClient.Scan(ctx, cursor, fmt.Sprintf("%s/*", layer), 5000).Result()
		if err != nil {
			return nil, errors.Wrapf(err, "redis scan layer '%s' with cursor '%d' failed", layer, cursor)
		}
		resultKeys = append(resultKeys, keys...)
		if cursor == 0 {
			break
		}
	}

	result := make([]*LayerLocatedInfo, 0, len(resultKeys))
	for _, key := range resultKeys {
		t := strings.Split(key, "/")
		if len(t) != 3 {
			logctx.Warnf(ctx, "redis key '%s' not normal", key)
			continue
		}
		located := t[1]
		ociType := t[2]
		if !slices.Contains(keyTypes, ociType) {
			continue
		}
		v, err := r.redisClient.Get(ctx, key).Result()
		if err != nil {
			logctx.Warnf(ctx, "redis key '%s' get failed: %s", key, err)
			continue
		}
		result = append(result, &LayerLocatedInfo{
			Layer:   layer,
			Type:    LayerType(ociType),
			Located: located,
			Data:    v,
		})
	}
	return result, nil
}
