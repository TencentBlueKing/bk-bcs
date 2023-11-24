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

// Package perf performance statistics
package perf

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
)

var (
	// performance 单例 Performance
	performance *Performance
	once        sync.Once
)

// GetGlobalPerformance : get global Performance
func GetGlobalPerformance() *Performance {
	if performance == nil {
		once.Do(func() {
			performance = &Performance{
				lock:            sync.Mutex{},
				meterDataChan:   make(chan *types.DelayData),
				meterKeyChanMap: map[string]*meterKeyChan{},
			}
		})
	}
	return performance
}

// Performance 性能统计列表临时存放map
type Performance struct {
	lock            sync.Mutex
	meterDataChan   chan *types.DelayData
	meterKeyChanMap map[string]*meterKeyChan
}

// 存放用户设置的命令延时数据, 格式为 key={username}, value={cluster_id}:{console_key}, 其中 cluster_id 为空代表任意
type meterKeyChan struct {
	username string
	c        chan string
}

// Run Performance 启动
func (p *Performance) Run(ctx context.Context) error {
	timer := time.NewTicker(time.Second * 2)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case m := <-p.meterDataChan:
			value, err := json.Marshal(m)
			if err != nil {
				logger.Errorf("meterData json Marshal failed, err: %s", err.Error())
				continue
			}
			key := types.GetMeterDataKey(m.Username)

			func() {
				pCtx, pCancel := context.WithTimeout(context.Background(), time.Second*2)
				defer pCancel()

				if err := storage.GetDefaultRedisSession().Client.RPush(pCtx, key, value).Err(); err != nil {
					logger.Errorf("redis list push failed, err: %s, key: %s", err.Error(), key)
					return
				}

				// 只保留最后 100 条数据
				if err := storage.GetDefaultRedisSession().Client.LTrim(pCtx, key, -100, -1).Err(); err != nil {
					logger.Errorf("redis ltrim failed, err: %s, key: %s", err.Error(), key)
					return
				}

				// 24小时过期
				if err := storage.GetDefaultRedisSession().Client.Expire(ctx, key, time.Hour*24).Err(); err != nil {
					logger.Errorf("redis expire failed, err: %s, key: %s", err.Error(), key)
					return
				}
			}()

		case <-timer.C:
			// 定时写入用户设置延时开关数据
			delayData, err := storage.GetDefaultRedisSession().Client.HGetAll(ctx, types.GetMeterKey()).Result()
			if errors.Is(err, context.Canceled) {
				return nil
			}

			if err != nil {
				logger.Errorf("failed to synchronize redis data, err: %s", err.Error())
				continue
			}

			for k, v := range delayData {
				p.Broadcast(k, v)
			}
		}
	}
}

// PushMeter 写入数据, 设置2秒超时保护机制
func (p *Performance) PushMeter(meters []*types.DelayData) int {
	for idx, m := range meters {
		select {
		case p.meterDataChan <- m:
		case <-time.After(2 * time.Second):
			logger.Warnf("timeout to push meter data, raw data: %v", m)
			return idx
		}
	}

	return len(meters)
}

// Subscribe 订阅 key 的变化
func (p *Performance) Subscribe(sessionID string, username string) chan string {
	st := time.Now()

	p.lock.Lock()
	defer p.lock.Unlock()

	outputC := &meterKeyChan{
		username: username,
		c:        make(chan string),
	}
	p.meterKeyChanMap[sessionID] = outputC

	logger.Infof("perf subscribe success, sessionID=%s, username=%s, duration=%s", sessionID, username, time.Since(st))
	return outputC.c
}

// UnSubscribe 取消订阅 key 的变化
func (p *Performance) UnSubscribe(sessionID string) {
	st := time.Now()

	p.lock.Lock()
	defer p.lock.Unlock()

	close(p.meterKeyChanMap[sessionID].c)
	delete(p.meterKeyChanMap, sessionID)

	logger.Infof("perf unsubscribe success, sessionID=%s, duration=%s", sessionID, time.Since(st))
}

// Broadcast 广播 key 配置
func (p *Performance) Broadcast(username, key string) {
	st := time.Now()

	p.lock.Lock()
	defer p.lock.Unlock()

	for _, v := range p.meterKeyChanMap {
		if v.username != username {
			continue
		}

		v.c <- key
	}

	duration := time.Since(st)
	if duration > time.Millisecond*100 {
		logger.Warnf("perf broadcast slow, username=%s, key=%s, duration=%s", username, key, duration)
	}
}
