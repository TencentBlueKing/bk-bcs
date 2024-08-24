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

package etcd

import (
	"context"
	"path/filepath"
	"time"

	"github.com/RichardKnop/machinery/v2/log"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	revokePrefix = "/machinery/v2/revoker/tasks"
)

type revokeSign struct {
	taskID       string
	registerTime time.Time
	ctx          context.Context
	cancel       context.CancelFunc
}

// Revoke etcd revoker send sign
func (b *etcdBroker) Revoke(ctx context.Context, taskID string) error {
	key := revokePrefix + "/" + taskID

	// 2分钟自动过期
	lease, err := b.client.Grant(ctx, int64(time.Second*120))
	if err != nil {
		return err
	}

	_, err = b.client.Put(ctx, key, time.Now().Format(time.RFC3339), clientv3.WithLease(lease.ID))
	if err != nil {
		return err
	}

	return nil
}

// RevokeCtx etcd revoker ctx
func (b *etcdBroker) RevokeCtx(taskID string) context.Context {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	sign, ok := b.revokeSignMap[taskID]
	if ok {
		sign.registerTime = time.Now()
		return sign.ctx
	}

	ctx, cancel := context.WithCancel(context.Background())
	sign = &revokeSign{
		taskID:       taskID,
		registerTime: time.Now(),
		ctx:          ctx,
		cancel:       cancel,
	}
	b.revokeSignMap[taskID] = sign

	return sign.ctx
}

func (b *etcdBroker) tryRevoke(kv *mvccpb.KeyValue) {
	key := string(kv.Key)
	taskID := filepath.Base(key)

	b.mtx.Lock()
	sign, ok := b.revokeSignMap[taskID]
	if ok {
		delete(b.revokeSignMap, taskID)
	}
	b.mtx.Unlock()

	if !ok {
		return
	}

	sign.cancel()

	ctx, cancel := context.WithTimeout(b.ctx, time.Second*5)
	defer cancel()

	_, err := b.client.Delete(ctx, key)
	if err != nil {
		log.ERROR.Printf("revoke %s failed: %s", key, err)
	}
}

func (b *etcdBroker) listWatchRevoke(ctx context.Context) error {
	// List
	listCtx, listCancel := context.WithTimeout(ctx, time.Second*10)
	defer listCancel()

	resp, err := b.client.Get(listCtx, revokePrefix, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		return err
	}

	for _, kv := range resp.Kvs {
		b.tryRevoke(kv)
	}

	// Watch
	watchCtx, watchCancel := context.WithTimeout(ctx, time.Minute*60)
	defer watchCancel()

	watchOpts := []clientv3.OpOption{
		clientv3.WithPrefix(),
		clientv3.WithKeysOnly(),
		clientv3.WithRev(resp.Header.Revision),
	}
	wc := b.client.Watch(watchCtx, revokePrefix, watchOpts...)
	for wresp := range wc {
		if wresp.Err() != nil {
			return watchCtx.Err()
		}

		for _, ev := range wresp.Events {
			if ev.Type != clientv3.EventTypePut {
				continue
			}

			b.tryRevoke(ev.Kv)
		}
	}

	return nil
}

func (b *etcdBroker) expireRevokeSign() {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	for taskID, sign := range b.revokeSignMap {
		if time.Since(sign.registerTime) > time.Hour*24 {
			sign.cancel()
			delete(b.revokeSignMap, taskID)
		}
	}
}
