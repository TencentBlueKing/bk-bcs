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

// Package bedis means bscp redis client package
package bedis

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/time/rate"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// ExpireMode defines the mode that how a key is to expire.
type ExpireMode string

const (
	// NX -- Set expiry only when the key has no expiry
	NX ExpireMode = "NX"
	// XX -- Set expiry only when the key has an existing expiry
	XX ExpireMode = "XX"
	// GT -- Set expiry only when the new expiry is greater than current one
	GT ExpireMode = "GT"
	// LT -- Set expiry only when the new expiry is less than current one
	LT ExpireMode = "LT"
)

// RedisClient redis cluster/standalone 方法
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	TxPipeline() redis.Pipeliner
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	GetSet(ctx context.Context, key string, value interface{}) *redis.StringCmd
	MGet(ctx context.Context, keys ...string) *redis.SliceCmd
	HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd
	HGet(ctx context.Context, key, field string) *redis.StringCmd
	HMGet(ctx context.Context, key string, fields ...string) *redis.SliceCmd
	HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Do(ctx context.Context, args ...interface{}) *redis.Cmd
	Ping(ctx context.Context) *redis.StatusCmd
	LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	LRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd
	RPop(ctx context.Context, key string) *redis.StringCmd
	Keys(ctx context.Context, pattern string) *redis.StringSliceCmd
	LLen(ctx context.Context, key string) *redis.IntCmd
	LTrim(ctx context.Context, key string, start, stop int64) *redis.StatusCmd
	LRem(ctx context.Context, key string, count int64, value interface{}) *redis.IntCmd
	RPopLPush(ctx context.Context, source, destination string) *redis.StringCmd
	BRPopLPush(ctx context.Context, source, destination string, timeout time.Duration) *redis.StringCmd
	ZAdd(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd
	ZRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.ZSliceCmd
	ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
}

// Client defines all the bscp used redis command
type Client interface {
	Set(ctx context.Context, key string, value interface{}, ttlSeconds int) error
	SetWithTxnPipe(ctx context.Context, kv map[string]string, ttlSeconds int) error
	HGetWithTxnPipe(ctx context.Context, hashKey string, field string) (string, error)
	SetNX(ctx context.Context, key string, value interface{}, ttlSeconds int) (bool, error)
	Get(ctx context.Context, key string) (string, error)
	GetSet(ctx context.Context, key string, value interface{}) (string, error)
	MGet(ctx context.Context, key ...string) ([]string, error)
	HSets(ctx context.Context, hashKey string, kv map[string]string, ttlSeconds int) error
	HDelete(ctx context.Context, hashKey string, subKey []string) error
	HDeleteWithTxPipe(ctx context.Context, multiHash map[string][]string) error
	DeleteWithTxPipe(ctx context.Context, keys []string) error
	HGet(ctx context.Context, hashKey string, field string) (string, error)
	HMGet(ctx context.Context, hashKey string, fields ...string) ([]string, error)
	HGetAll(ctx context.Context, hashKey string) (map[string]string, error)
	Delete(ctx context.Context, keys ...string) error
	Expire(ctx context.Context, key string, ttlSeconds int, mode ExpireMode) error
	Healthz() error
	LPush(ctx context.Context, key string, values ...interface{}) error
	RPush(ctx context.Context, key string, values ...interface{}) error
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	RPop(ctx context.Context, key string) (string, error)
	Keys(ctx context.Context, pattern string) ([]string, error)
	LLen(ctx context.Context, key string) (int64, error)
	LTrim(ctx context.Context, key string, start, stop int64) (string, error)
	LRem(ctx context.Context, key string, count int64, value interface{}) error
	RPopLPush(ctx context.Context, source, destination string) (string, error)
	BRPopLPush(ctx context.Context, source, destination string, ttlSeconds int) (string, error)
	ZAdd(ctx context.Context, key string, score float64, value interface{}) (int64, error)
	ZRangeByScoreWithScores(ctx context.Context, key string, zRangeBy *redis.ZRangeBy) ([]redis.Z, error)
	ZRem(ctx context.Context, key string, members ...interface{}) (int64, error)
}

// NewRedisCache create a redis cluster client.
func NewRedisCache(opt cc.RedisCluster) (Client, error) {
	var (
		client RedisClient
		err    error
	)

	// 支持多种 redis 模式
	switch opt.Mode {
	case cc.RedisClusterMode:
		client, err = newClusterClient(opt)
	default:
		client, err = newStandaloneClient(opt)
	}

	if err != nil {
		return nil, err
	}

	bs := &bedis{
		client:              client,
		mc:                  initMetric(),
		logLimiter:          rate.NewLimiter(50, 10),
		maxSlowLogLatencyMS: time.Duration(opt.MaxSlowLogLatencyMS) * time.Millisecond,
	}

	return bs, nil
}

var _ Client = (*bedis)(nil)

// bedis is an implement of the bscp redis client.
type bedis struct {
	client              RedisClient
	mc                  *metric
	maxSlowLogLatencyMS time.Duration
	logLimiter          *rate.Limiter
}

func (bs *bedis) logSlowCmd(ctx context.Context, key string, latency time.Duration) {

	if latency < bs.maxSlowLogLatencyMS {
		return
	}

	if !bs.logLimiter.Allow() {
		// if the log rate have already exceeded the limit, then skip the log.
		// we do this to avoid write lots of log to file and slow down the request.
		return
	}

	rid := ctx.Value(constant.RidKey)
	logs.InfoDepthf(2, "[bedis slow log], key: %s, latency: %d ms, rid: %v", key, latency.Milliseconds(), rid)
}

// ErrKeyNotExist describe the error that the key is not exist in redis.
var ErrKeyNotExist = errors.New("redis key not exist")

// IsNilError test if the error is returned with redis client and is a nil error
// which means get the key with a nil value.
func IsNilError(err error) bool {
	return strings.Contains(err.Error(), redis.Nil.Error())
}

// IsWrongTypeError test if an error is caused by used the wrong redis command
func IsWrongTypeError(err error) bool {
	if err == nil {
		return false
	}

	// redis server return with error:
	// 'WRONGTYPE Operation against a key holding the wrong kind of value'
	if strings.Contains(err.Error(), "WRONGTYPE Operation") {
		// this is because the hashmap is set with a null value without the hash field,
		// which convert the expected hash key to a common KV cache.
		return true
	}

	return false
}

// newClusterClient create a redis cluster client.
func newClusterClient(opt cc.RedisCluster) (RedisClient, error) {
	var tlsC *tls.Config
	if opt.TLS.Enable() {
		var err error
		tlsC, err = tools.ClientTLSConfVerify(opt.TLS.InsecureSkipVerify, opt.TLS.CAFile, opt.TLS.CertFile,
			opt.TLS.KeyFile, opt.TLS.Password)
		if err != nil {
			return nil, fmt.Errorf("init redis tls config failed, err: %v", err)
		}
	}

	clusterOpt := &redis.ClusterOptions{
		Addrs:              opt.Endpoints,
		Username:           opt.Username,
		Password:           opt.Password,
		DialTimeout:        time.Duration(opt.DialTimeoutMS) * time.Millisecond,
		ReadTimeout:        time.Duration(opt.ReadTimeoutMS) * time.Millisecond,
		WriteTimeout:       time.Duration(opt.WriteTimeoutMS) * time.Millisecond,
		PoolSize:           int(opt.PoolSize),
		MinIdleConns:       int(opt.MinIdleConn),
		MaxConnAge:         0,
		PoolTimeout:        0,
		IdleTimeout:        0,
		IdleCheckFrequency: 0,
		TLSConfig:          tlsC,
	}

	client := redis.NewClusterClient(clusterOpt)
	if err := client.Ping(context.TODO()).Err(); err != nil {
		return nil, fmt.Errorf("init redis cluster client, but ping failed, err: %v", err)
	}

	return client, nil
}

// newStandaloneClient create a redis standalone client.
func newStandaloneClient(opt cc.RedisCluster) (RedisClient, error) {
	var tlsC *tls.Config
	if opt.TLS.Enable() {
		var err error
		tlsC, err = tools.ClientTLSConfVerify(opt.TLS.InsecureSkipVerify, opt.TLS.CAFile, opt.TLS.CertFile,
			opt.TLS.KeyFile, opt.TLS.Password)
		if err != nil {
			return nil, fmt.Errorf("init redis tls config failed, err: %v", err)
		}
	}

	clusterOpt := &redis.Options{
		Addr:         opt.Endpoints[0],
		Password:     opt.Password,
		DB:           opt.DB,
		DialTimeout:  time.Duration(opt.DialTimeoutMS) * time.Millisecond,
		ReadTimeout:  time.Duration(opt.ReadTimeoutMS) * time.Millisecond,
		WriteTimeout: time.Duration(opt.WriteTimeoutMS) * time.Millisecond,
		PoolSize:     int(opt.PoolSize),
		MinIdleConns: int(opt.MinIdleConn),
		TLSConfig:    tlsC,
	}

	client := redis.NewClient(clusterOpt)
	if err := client.Ping(context.TODO()).Err(); err != nil {
		return nil, fmt.Errorf("init redis cluster client, but ping failed, err: %v", err)
	}

	return client, nil
}
