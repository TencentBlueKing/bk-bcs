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

// Package etcd is revoker use etcd
package etcd

import (
	"context"
	"sync"
	"time"

	"github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/log"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/revokers/iface"
)

type etcdRevoker struct {
	ctx           context.Context
	client        *clientv3.Client
	mtx           sync.Mutex
	wg            sync.WaitGroup
	revokeSignMap map[string]*revokeSign
}

// New ..
func New(ctx context.Context, conf *config.Config) (iface.Revoker, error) {
	etcdConf := clientv3.Config{
		Endpoints:   []string{conf.Broker}, // 复用broker的etcd配置
		Context:     ctx,
		DialTimeout: time.Second * 5,
		TLS:         conf.TLSConfig,
	}

	client, err := clientv3.New(etcdConf)
	if err != nil {
		return nil, err
	}

	revoker := etcdRevoker{
		ctx:           ctx,
		client:        client,
		revokeSignMap: map[string]*revokeSign{},
	}

	go revoker.Run()

	return &revoker, nil
}

func (r *etcdRevoker) Run() {
	// list and watch revoke sign
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		for {
			select {
			case <-r.ctx.Done():
				return
			default:
				err := r.listWatchRevoke(r.ctx)
				if err != nil {
					log.ERROR.Printf("list and watch revoke failed, err: %s", err)
					time.Sleep(time.Second)
				}
			}
		}
	}()

	// cleanup revoke sign
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		ticker := time.NewTicker(time.Minute * 10)
		defer ticker.Stop()

		for {
			select {
			case <-r.ctx.Done():
				return
			case <-ticker.C:
				r.cleanupRevokeSign()
			}
		}
	}()

	r.wg.Wait()
}
