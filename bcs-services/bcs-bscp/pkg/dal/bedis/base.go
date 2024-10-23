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

package bedis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	prm "github.com/prometheus/client_golang/prometheus"
)

// Set a key with value
func (bs *bedis) Set(ctx context.Context, key string, value interface{}, ttlSeconds int) error {

	start := time.Now()
	_, err := bs.client.Set(ctx, key, value, time.Duration(ttlSeconds)*time.Second).Result()
	if err != nil {
		bs.mc.errCounter.With(prm.Labels{"cmd": "set"}).Inc()
		return err
	}

	bs.logSlowCmd(ctx, key, time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "set"}).Observe(float64(time.Since(start).Milliseconds()))

	return nil
}

// SetWithTxnPipe a list of key and value
func (bs *bedis) SetWithTxnPipe(ctx context.Context, kv map[string]string, ttlSeconds int) error {
	if len(kv) == 0 {
		return errors.New("not keys to set")
	}

	start := time.Now()
	expire := time.Duration(ttlSeconds) * time.Second

	// do with transaction pipe.
	pipe := bs.client.TxPipeline()
	for k, v := range kv {
		pipe.Set(ctx, k, v, expire)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		bs.mc.errCounter.With(prm.Labels{"cmd": "set_txn_pipe"}).Inc()
		return err
	}

	bs.logSlowCmd(ctx, "", time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "set_txn_pipe"}).Observe(float64(time.Since(start).Milliseconds()))

	return nil
}

// HGetWithTxnPipe a hashmap's key and its value
// If hashKey does not exist, return value is "" and err is nil.
// If hashKey exists but field does not exist, return ErrKeyNotExist.
func (bs *bedis) HGetWithTxnPipe(ctx context.Context, hashKey string, field string) (string, error) {
	start := time.Now()

	// do with transaction pipe.
	pipe := bs.client.TxPipeline()
	existResult := pipe.Exists(ctx, hashKey)
	valueResult := pipe.HGet(ctx, hashKey, field)

	_, err := pipe.Exec(ctx)
	if err != nil && !IsNilError(err) {
		bs.mc.errCounter.With(prm.Labels{"cmd": "hget_txn_pipe"}).Inc()
		return "", err
	}

	// judge hashKey if exist
	result, err := existResult.Result()
	if err != nil {
		return "", err
	}

	if result == 0 {
		return "", nil
	}

	// hashKey exist, judge field if exist.
	value, err := valueResult.Result()
	if err != nil {
		if IsNilError(err) {
			return "", ErrKeyNotExist
		}

		return "", err
	}

	bs.logSlowCmd(ctx, hashKey, time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "hget_txn_pipe"}).Observe(float64(time.Since(start).Milliseconds()))

	return value, nil
}

// SetNX set a key if the key is not exist.
func (bs *bedis) SetNX(ctx context.Context, key string, value interface{}, ttlSeconds int) (bool, error) {
	return bs.client.SetNX(ctx, key, value, time.Duration(ttlSeconds)*time.Second).Result()
}

// Get a key's value.
func (bs *bedis) Get(ctx context.Context, key string) (string, error) {

	start := time.Now()
	value, err := bs.client.Get(ctx, key).Result()
	if err != nil {
		if IsNilError(err) {
			return "", nil
		}

		bs.mc.errCounter.With(prm.Labels{"cmd": "get"}).Inc()
		return "", err
	}

	bs.logSlowCmd(ctx, key, time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "get"}).Observe(float64(time.Since(start).Milliseconds()))

	return value, nil
}

// GetSet get a key's value and replace it with a new value.
func (bs *bedis) GetSet(ctx context.Context, key string, value interface{}) (string, error) {

	start := time.Now()
	oldValue, err := bs.client.GetSet(ctx, key, value).Result()
	if err != nil {
		if IsNilError(err) {
			return "", nil
		}

		bs.mc.errCounter.With(prm.Labels{"cmd": "getset"}).Inc()
		return "", err
	}

	bs.logSlowCmd(ctx, key, time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "getset"}).Observe(float64(time.Since(start).Milliseconds()))

	return oldValue, nil
}

// MGet many key's value.
func (bs *bedis) MGet(ctx context.Context, key ...string) ([]string, error) {

	start := time.Now()
	list, err := bs.client.MGet(ctx, key...).Result()
	if err != nil {
		if IsNilError(err) {
			return nil, nil
		}

		bs.mc.errCounter.With(prm.Labels{"cmd": "mget"}).Inc()
		return nil, err
	}

	values := make([]string, 0)
	for _, val := range list {
		if val == nil {
			continue
		}

		one, yes := val.(string)
		if !yes {
			return nil, errors.New("invalid MGET cmd values, not string")
		}

		if len(one) == 0 {
			continue
		}

		values = append(values, one)
	}

	bs.logSlowCmd(ctx, "", time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "mget"}).Observe(float64(time.Since(start).Milliseconds()))

	return values, nil
}

// HSets set the hash key and kv list with a ttl.
func (bs *bedis) HSets(ctx context.Context, hashKey string, kv map[string]string, ttlSeconds int) error {
	if len(hashKey) == 0 || len(kv) == 0 {
		return errors.New("invalid redis HSET args, hash key or values is empty")
	}

	start := time.Now()

	// do with transaction pipe.
	pipe := bs.client.TxPipeline()
	for k, v := range kv {
		pipe.HSet(ctx, hashKey, k, v)
	}
	// set expire ttl.
	pipe.Expire(ctx, hashKey, time.Duration(ttlSeconds)*time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		bs.mc.errCounter.With(prm.Labels{"cmd": "hsets"}).Inc()
		return err
	}

	bs.logSlowCmd(ctx, hashKey, time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "hsets"}).Observe(float64(time.Since(start).Milliseconds()))

	return nil
}

// HDelete delete the hash key and kv list.
func (bs *bedis) HDelete(ctx context.Context, hashKey string, subKey []string) error {
	if len(hashKey) == 0 || len(subKey) == 0 {
		return errors.New("invalid redis HDEL args, hash key or sub-keys is empty")
	}

	start := time.Now()

	if err := bs.client.HDel(ctx, hashKey, subKey...).Err(); err != nil {
		bs.mc.errCounter.With(prm.Labels{"cmd": "hdel"}).Inc()
		return err
	}

	bs.logSlowCmd(ctx, hashKey, time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "hdel"}).Observe(float64(time.Since(start).Milliseconds()))

	return nil
}

// HDeleteWithTxPipe delete batch of the hash key and kv list.
func (bs *bedis) HDeleteWithTxPipe(ctx context.Context, multiHash map[string][]string) error {
	if len(multiHash) == 0 {
		return errors.New("invalid redis HDEL with pipe args, hash key is empty")
	}

	start := time.Now()

	// do with transaction pipe.
	pipe := bs.client.TxPipeline()
	for hash, subKeys := range multiHash {
		if len(hash) == 0 {
			return errors.New("invalid hdel hash key, can not be empty")
		}

		if len(subKeys) == 0 {
			return errors.New("invalid hdel hash sub keys, can not be empty")
		}

		pipe.HDel(ctx, hash, subKeys...)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		bs.mc.errCounter.With(prm.Labels{"cmd": "hdel-pipe"}).Inc()
		return err
	}

	bs.logSlowCmd(ctx, "", time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "hdel-pipe"}).Observe(float64(time.Since(start).Milliseconds()))

	return nil
}

// DeleteWithTxPipe delete batch of the key.
func (bs *bedis) DeleteWithTxPipe(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return errors.New("invalid redis keys with pipe args, keys is empty")
	}

	start := time.Now()

	// do with transaction pipe.
	pipe := bs.client.TxPipeline()
	pipe.Del(ctx, keys...)

	_, err := pipe.Exec(ctx)
	if err != nil {
		bs.mc.errCounter.With(prm.Labels{"cmd": "del-pipe"}).Inc()
		return err
	}

	bs.logSlowCmd(ctx, "", time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "del-pipe"}).Observe(float64(time.Since(start).Milliseconds()))

	return nil
}

// HGet get a hashmap's key and its value.
func (bs *bedis) HGet(ctx context.Context, hashKey string, field string) (string, error) {

	start := time.Now()

	value, err := bs.client.HGet(ctx, hashKey, field).Result()
	if err != nil {
		if IsNilError(err) {
			return "", ErrKeyNotExist
		}

		bs.mc.errCounter.With(prm.Labels{"cmd": "hget"}).Inc()
		return "", err
	}

	bs.logSlowCmd(ctx, hashKey, time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "hget"}).Observe(float64(time.Since(start).Milliseconds()))

	return value, nil
}

// HMGet get a hashmap's multiple key and its values.
func (bs *bedis) HMGet(ctx context.Context, hashKey string, fields ...string) ([]string, error) {
	start := time.Now()

	list, err := bs.client.HMGet(ctx, hashKey, fields...).Result()
	if err != nil {
		if IsNilError(err) {
			return make([]string, 0), nil
		}

		bs.mc.errCounter.With(prm.Labels{"cmd": "hmget"}).Inc()
		return nil, err
	}

	values := make([]string, 0)
	for _, val := range list {
		if val == nil {
			continue
		}

		one, yes := val.(string)
		if !yes {
			return nil, errors.New("invalid HMGET cmd values, not string")
		}

		if len(one) == 0 {
			continue
		}

		values = append(values, one)
	}

	bs.logSlowCmd(ctx, hashKey, time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "hmget"}).Observe(float64(time.Since(start).Milliseconds()))

	return values, nil
}

// HGetAll get a hashmap's all the key and its values.
func (bs *bedis) HGetAll(ctx context.Context, key string) (map[string]string, error) {

	start := time.Now()

	kv, err := bs.client.HGetAll(ctx, key).Result()
	if err != nil {
		if IsNilError(err) {
			return make(map[string]string), nil
		}

		bs.mc.errCounter.With(prm.Labels{"cmd": "hgetall"}).Inc()
		return nil, err
	}

	bs.logSlowCmd(ctx, key, time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "hgetall"}).Observe(float64(time.Since(start).Milliseconds()))

	return kv, nil
}

// Delete the keys.
func (bs *bedis) Delete(ctx context.Context, keys ...string) error {

	start := time.Now()

	if err := bs.client.Del(ctx, keys...).Err(); err != nil {
		bs.mc.errCounter.With(prm.Labels{"cmd": "delete"}).Inc()
		return err
	}

	bs.logSlowCmd(ctx, "", time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "delete"}).Observe(float64(time.Since(start).Milliseconds()))

	return nil
}

// Expire a key with mode if its mode is set.
func (bs *bedis) Expire(ctx context.Context, key string, ttlSeconds int, mode ExpireMode) error {
	start := time.Now()

	args := make([]interface{}, 3, 4)
	args[0] = "expire"
	args[1] = key
	args[2] = ttlSeconds

	switch mode {
	case NX, XX, GT, LT:
		args = append(args, mode)
	default:
		if len(mode) != 0 {
			return fmt.Errorf("unsupported expire mode: %s", mode)
		}
	}

	if err := bs.client.Do(ctx, args).Err(); err != nil {
		bs.mc.errCounter.With(prm.Labels{"cmd": "expire"}).Inc()
		return err
	}

	bs.logSlowCmd(ctx, key, time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "expire"}).Observe(float64(time.Since(start).Milliseconds()))

	return nil
}

// Healthz check redis-cluster health.
func (bs *bedis) Healthz() error {
	ctx, cancel := context.WithTimeout(context.TODO(), 15*time.Second)
	defer cancel()
	if err := bs.client.Ping(ctx).Err(); err != nil {
		return errors.New("redis cluster ping failed, err: " + err.Error())
	}

	return nil
}

// LPush 将一个或多个值插入到列表头部
func (bs *bedis) LPush(ctx context.Context, key string, values ...interface{}) error {
	start := time.Now()
	_, err := bs.client.LPush(ctx, key, values).Result()
	if err != nil {
		bs.mc.errCounter.With(prm.Labels{"cmd": "lpush"}).Inc()
		return err
	}
	bs.logSlowCmd(ctx, key, time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "lpush"}).Observe(float64(time.Since(start).Milliseconds()))
	return nil
}

// RPush 在列表中添加一个或多个值到列表尾部
func (bs *bedis) RPush(ctx context.Context, key string, values ...interface{}) error {
	start := time.Now()
	_, err := bs.client.RPush(ctx, key, values).Result()
	if err != nil {
		bs.mc.errCounter.With(prm.Labels{"cmd": "rpush"}).Inc()
		return err
	}
	bs.logSlowCmd(ctx, key, time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "rpush"}).Observe(float64(time.Since(start).Milliseconds()))
	return nil
}

// LRange 获取列表指定范围内的元素
func (bs *bedis) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	startTime := time.Now()
	list, err := bs.client.LRange(ctx, key, start, stop).Result()
	if err != nil {
		if IsNilError(err) {
			return nil, nil
		}
		bs.mc.errCounter.With(prm.Labels{"cmd": "lrange"}).Inc()
		return nil, err
	}
	values := make([]string, 0)
	for _, val := range list {
		if val == "" {
			continue
		}
		if len(val) == 0 {
			continue
		}

		values = append(values, val)
	}
	bs.logSlowCmd(ctx, "", time.Since(startTime))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "lrange"}).Observe(float64(time.Since(startTime).Milliseconds()))
	return values, nil
}

// RPop 移除列表的最后一个元素，返回值为移除的元素
func (bs *bedis) RPop(ctx context.Context, key string) (string, error) {
	start := time.Now()
	value, err := bs.client.RPop(ctx, key).Result()
	if err != nil {
		if IsNilError(err) {
			return "", nil
		}
		bs.mc.errCounter.With(prm.Labels{"cmd": "rpop"}).Inc()
		return "", err
	}
	bs.logSlowCmd(ctx, "", time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "rpop"}).Observe(float64(time.Since(start).Milliseconds()))

	return value, nil
}

// Keys finds all keys that match a given pattern
func (bs *bedis) Keys(ctx context.Context, pattern string) ([]string, error) {
	start := time.Now()
	list, err := bs.client.Keys(ctx, pattern).Result()
	if err != nil {
		if IsNilError(err) {
			return nil, nil
		}
		bs.mc.errCounter.With(prm.Labels{"cmd": "keys"}).Inc()
		return nil, err
	}
	values := make([]string, 0)
	for _, val := range list {
		if val == "" {
			continue
		}
		if len(val) == 0 {
			continue
		}

		values = append(values, val)
	}
	bs.logSlowCmd(ctx, "", time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "keys"}).Observe(float64(time.Since(start).Milliseconds()))
	return values, nil
}

// LLen get list length
func (bs *bedis) LLen(ctx context.Context, key string) (int64, error) {
	start := time.Now()
	value, err := bs.client.LLen(ctx, key).Result()
	if err != nil {
		if IsNilError(err) {
			return 0, nil
		}
		bs.mc.errCounter.With(prm.Labels{"cmd": "llen"}).Inc()
		return 0, err
	}
	bs.logSlowCmd(ctx, "", time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "llen"}).Observe(float64(time.Since(start).Milliseconds()))

	return value, nil
}

// LTrim trim a list, that is, make the list keep only the elements within the specified interval,
// and the elements that are not within the specified interval are deleted.
func (bs *bedis) LTrim(ctx context.Context, key string, start, stop int64) (string, error) {
	startTime := time.Now()
	value, err := bs.client.LTrim(ctx, key, start, stop).Result()
	if err != nil {
		if IsNilError(err) {
			return "", nil
		}
		bs.mc.errCounter.With(prm.Labels{"cmd": "ltrim"}).Inc()
		return "", err
	}
	bs.logSlowCmd(ctx, "", time.Since(startTime))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "ltrim"}).Observe(float64(time.Since(startTime).Milliseconds()))

	return value, nil
}

// LRem removes the first count occurrences of elements equal to element from the list stored at key.
// count argument influences the operation in the following ways:
// count > 0: Remove elements equal to element moving from head to tail.
// count < 0: Remove elements equal to element moving from tail to head.
// count = 0: Remove all elements equal to element.
func (bs *bedis) LRem(ctx context.Context, key string, count int64, value interface{}) error {
	start := time.Now()
	_, err := bs.client.LRem(ctx, key, count, value).Result()
	if err != nil {
		bs.mc.errCounter.With(prm.Labels{"cmd": "lrem"}).Inc()
		return err
	}
	bs.logSlowCmd(ctx, key, time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "lrem"}).Observe(float64(time.Since(start).Milliseconds()))
	return nil
}

// RPopLPush atomically returns and removes the last element (tail) of the list stored at source,
// and pushes the element at the first element (head) of the list stored at destination.
func (bs *bedis) RPopLPush(ctx context.Context, source, destination string) (string, error) {
	start := time.Now()
	value, err := bs.client.RPopLPush(ctx, source, destination).Result()
	if err != nil {
		if IsNilError(err) {
			return "", nil
		}
		bs.mc.errCounter.With(prm.Labels{"cmd": "rpoplpush"}).Inc()
		return "", err
	}
	bs.logSlowCmd(ctx, "", time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "rpoplpush"}).Observe(float64(time.Since(start).Milliseconds()))

	return value, nil
}

// BRPopLPush is the blocking variant of RPOPLPUSH.
// When source contains elements, this command behaves exactly like RPOPLPUSH.
// When source is empty, Redis will block the connection until another client pushes to it or until timeout is reached.
// A timeout of zero can be used to block indefinitely.
func (bs *bedis) BRPopLPush(ctx context.Context, source, destination string, ttlSeconds int) (string, error) {
	start := time.Now()
	value, err := bs.client.BRPopLPush(ctx, source, destination, time.Duration(ttlSeconds)*time.Second).Result()
	if err != nil {
		if IsNilError(err) {
			return "", nil
		}
		bs.mc.errCounter.With(prm.Labels{"cmd": "brpoplpush"}).Inc()
		return "", err
	}
	bs.logSlowCmd(ctx, "", time.Since(start))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "brpoplpush"}).Observe(float64(time.Since(start).Milliseconds()))

	return value, nil
}

// ZAdd Redis `ZADD key score member [score member ...]` command.
func (bs *bedis) ZAdd(ctx context.Context, key string, score float64, value interface{}) (int64, error) {
	startTime := time.Now()
	r, err := bs.client.ZAdd(ctx, key, &redis.Z{
		Score:  score,
		Member: value,
	}).Result()
	if err != nil {
		if IsNilError(err) {
			return 0, nil
		}
		bs.mc.errCounter.With(prm.Labels{"cmd": "zadd"}).Inc()
		return 0, err
	}
	bs.logSlowCmd(ctx, "", time.Since(startTime))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "zadd"}).Observe(float64(time.Since(startTime).Milliseconds()))

	return r, nil
}

// ZRangeByScoreWithScores zrangebyscore with scores
func (bs *bedis) ZRangeByScoreWithScores(ctx context.Context, key string, zRangeBy *redis.ZRangeBy) ([]redis.Z, error) {
	startTime := time.Now()
	r, err := bs.client.ZRangeByScoreWithScores(ctx, key, zRangeBy).Result()
	if err != nil {
		if IsNilError(err) {
			return []redis.Z{}, nil
		}
		bs.mc.errCounter.With(prm.Labels{"cmd": "zrange withscores"}).Inc()
		return []redis.Z{}, err
	}
	bs.logSlowCmd(ctx, "", time.Since(startTime))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "zrange withscores"}).Observe(float64(time.Since(startTime).Milliseconds()))

	return r, nil
}

// ZRem delete zset member
func (bs *bedis) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	startTime := time.Now()
	r, err := bs.client.ZRem(ctx, key, members...).Result()
	if err != nil {
		if IsNilError(err) {
			return 0, nil
		}
		bs.mc.errCounter.With(prm.Labels{"cmd": "zrem"}).Inc()
		return 0, err
	}
	bs.logSlowCmd(ctx, "", time.Since(startTime))
	bs.mc.cmdLagMS.With(prm.Labels{"cmd": "zrem"}).Observe(float64(time.Since(startTime).Milliseconds()))

	return r, nil
}
