/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package bedis

import (
	"context"
	"errors"
	"fmt"
	"time"

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
