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
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
)

var (
	// performance 单例 Performance
	performance *Performance
	once        sync.Once
	// 存放用户设置的命令延时数据
	commandDelay = make(map[string]string, 0)
)

// GetGlobalPerformance : get global Performance
func GetGlobalPerformance() *Performance {
	if performance == nil {
		once.Do(func() {
			performance = &Performance{
				userDelayList: map[string][]string{},
				lock:          sync.Mutex{},
			}

		})
	}
	return performance
}

// Performance 性能统计列表临时存放map
type Performance struct {
	// 临时存放用户延时数据列表
	userDelayList map[string][]string
	lock          sync.Mutex
}

// IsOpenDelay 判断用户是否开启延时统计
func (p *Performance) IsOpenDelay(username, clusterId, msg string) bool {
	// 匹配子字符串，如果包含则表示开启了命令延时统计
	msgPart := "\"cluster_id\":\"" + clusterId + "\",\"enabled\":true,\"console_key\":\"" + msg
	return strings.Contains(commandDelay[username], msgPart)
}

// SetUserDelayList 设置用户列表数据
func (p *Performance) SetUserDelayList(key string, s string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.userDelayList[key] = append(p.userDelayList[key], s)
}

// Run Performance 启动
func (p *Performance) Run(ctx context.Context) error {

	timer := time.NewTicker(time.Second * 5)
	defer timer.Stop()

	userSync := time.NewTicker(time.Second * 1)
	defer userSync.Stop()

	for {
		select {
		case <-ctx.Done():
			// 获取退出信号, 全部上传
			err := p.batchUploadUserDelay()
			if err != nil {
				return err
			}
		case <-timer.C:
			err := p.batchUploadUserDelay()
			if err != nil {
				return err
			}
		case <-userSync.C:
			// 定时写入用户设置延时开关数据
			delayData, err := storage.GetDefaultRedisSession().Client.HGetAll(ctx, types.ConsoleKey).Result()
			if err != nil {
				blog.Errorf("failed to synchronize redis data, err: %s", err.Error())
				return err
			}
			commandDelay = delayData
		}
	}
}

// batchUploadUserDelay 批量上传
func (p *Performance) batchUploadUserDelay() error {
	p.lock.Lock()
	defer p.lock.Unlock()
	// 重置
	for i, value := range p.userDelayList {
		// 查看用户是否已经有统计数据在Redis中
		listLen := storage.GetDefaultRedisSession().Client.LLen(context.Background(), types.DelayUser+i).Val()
		// 往Redis追加数据
		err := storage.GetDefaultRedisSession().Client.RPush(context.Background(), types.DelayUser+i, value).Err()
		if err != nil {
			blog.Errorf("redis list push failed, err: %s", err.Error())
			return err
		}
		// 没有数据的情况下设置列表过期时间，暂定一天
		if listLen == 0 {
			// 列表设置过期时间
			storage.GetDefaultRedisSession().Client.Expire(
				context.Background(), types.DelayUser+i, types.DelayUserExpire)
		}
	}
	p.userDelayList = map[string][]string{}
	return nil
}
