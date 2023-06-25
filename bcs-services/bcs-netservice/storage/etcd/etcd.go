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
 *
 */

// Package etcd xxx
package etcd

import (
	"context"
	"crypto/tls"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/storage"
)

const (
	etcdDialTimeout           = 5
	defaultEtcdOperateTimeout = 10
)

// etcdStorage storage data in etcd
type etcdStorage struct {
	sync.Mutex

	client  *clientv3.Client
	session *concurrency.Session
	lockMap sync.Map
}

// NewStorage create etcd storage
func NewStorage(endpoints []string, tlsCfg *tls.Config) (storage.Storage, error) {
	cli, err := clientv3.New(clientv3.Config{
		DialTimeout: etcdDialTimeout * time.Second,
		Endpoints:   endpoints,
		TLS:         tlsCfg,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "create etcd client failed")
	}
	blog.Infof("create etcd client success.")
	// create concurrency session
	session, err := concurrency.NewSession(cli)
	if err != nil {
		return nil, errors.Wrapf(err, "create etcd concurrency session failed")
	}
	blog.Infof("create etcd session success.")
	s := &etcdStorage{
		client:  cli,
		session: session,
	}
	return s, nil
}

// Add create the key into etcd
func (e *etcdStorage) Add(key string, value []byte) error {
	if err := e.Update(key, value); err != nil {
		return errors.Wrapf(err, "etcd add key '%s' failed", key)
	}
	return nil
}

// Delete xxx
func (e *etcdStorage) Delete(key string) ([]byte, error) {
	result, err := e.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "get key '%s' failed", key)
	}

	kv := clientv3.NewKV(e.client)
	ctx, cancel := context.WithTimeout(context.Background(), defaultEtcdOperateTimeout*time.Second)
	defer cancel()
	if _, err = kv.Delete(ctx, key); err != nil {
		return nil, errors.Wrapf(err, "etcd delete key '%s' failed", key)
	}
	blog.Infof("delete key '%s' success: %s", key, string(result))
	return result, nil
}

// Update xxx
func (e *etcdStorage) Update(key string, value []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultEtcdOperateTimeout*time.Second)
	defer cancel()

	kv := clientv3.NewKV(e.client)
	v := string(value)
	if _, err := kv.Put(ctx, key, v); err != nil {
		return errors.Wrapf(err, "etcd update '%s' with value '%s' failed", key, v)
	}
	return nil
}

// Get xxx
func (e *etcdStorage) Get(key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultEtcdOperateTimeout*time.Second)
	defer cancel()
	kv := clientv3.NewKV(e.client)
	getResp, err := kv.Get(ctx, key)
	if err != nil {
		return nil, errors.Wrapf(err, "etcd get key '%s' failed", key)
	}
	if len(getResp.Kvs) != 1 {
		return nil, errors.Errorf("etcd get key '%s' resp len not 1 but %d", key, len(getResp.Kvs))
	}
	gr := getResp.Kvs[0]
	return gr.Value, nil
}

// List xxx
func (e *etcdStorage) List(key string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultEtcdOperateTimeout*time.Second)
	defer cancel()
	kv := clientv3.NewKV(e.client)
	getResp, err := kv.Get(ctx, key+"/", clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrapf(err, "etcd list key '%s' failed", key)
	}
	result := make([]string, 0, len(getResp.Kvs))
	for _, kv := range getResp.Kvs {
		keySlice := strings.Split(string(kv.Key), "/")
		result = append(result, keySlice[len(keySlice)-1])
	}
	return result, nil
}

// Register xxx
func (e *etcdStorage) Register(path string, data []byte) error {
	lease := clientv3.NewLease(e.client)
	leaseResp, err := lease.Grant(context.TODO(), 10)
	if err != nil {
		return errors.Wrapf(err, "lease grant failed")
	}
	kv := clientv3.NewKV(e.client)
	if _, err := kv.Put(context.TODO(), path, string(data), clientv3.WithLease(leaseResp.ID)); err != nil {
		return errors.Wrapf(err, "update register key '%s' failed", path)
	}
	return nil
}

// RegisterAndWatch xxx
func (e *etcdStorage) RegisterAndWatch(path string, data []byte) error {
	if err := e.Register(path, data); err != nil {
		return errors.Wrapf(err, "register and watch '%s' failed", path)
	}
	go func() {
		tick := time.NewTicker(5 * time.Second)
		defer tick.Stop()
		for range tick.C {
			if err := e.Register(path, data); err != nil {
				blog.Errorf("register failed: %s", err.Error())
			}
		}
	}()
	return nil
}

// Exist check key exist
func (e *etcdStorage) Exist(key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultEtcdOperateTimeout*time.Second)
	defer cancel()
	kv := clientv3.NewKV(e.client)
	getResp, err := kv.Get(ctx, key)
	if err != nil {
		return false, errors.Wrapf(err, "etcd get key '%s' failed", key)
	}
	if len(getResp.Kvs) == 0 {
		return false, nil
	}
	return true, nil
}

// GetLocker will return the locker with key
func (e *etcdStorage) GetLocker(key string) (storage.Locker, error) {
	e.Lock()
	defer e.Unlock()

	etcdLock, _ := e.lockMap.LoadOrStore(key, concurrency.NewMutex(e.session, key))
	locker := &etcdLocker{
		path:  key,
		mutex: etcdLock.(*concurrency.Mutex),
	}
	return locker, nil
}

// Stop xxx
func (e *etcdStorage) Stop() {
	e.client.Close()
}

type etcdLocker struct {
	path  string
	mutex *concurrency.Mutex
}

// Lock the key
func (l *etcdLocker) Lock() error {
	blog.Infof("etcd locking '%s'", l.path)
	if err := l.mutex.Lock(context.Background()); err != nil {
		return errors.Wrapf(err, "etcd lock '%s' failed", l.path)
	}
	blog.Infof("etcd locked '%s'", l.path)
	return nil
}

// Unlock the key
func (l *etcdLocker) Unlock() error {
	if err := l.mutex.Unlock(context.Background()); err != nil {
		return errors.Wrapf(err, "etcd unlock '%s' failed", l.path)
	}
	blog.Infof("etcd unlocked '%s'", l.path)
	return nil
}
