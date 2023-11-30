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

// Package activity job
package activity

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
)

// IntervalDeleteActivity 定时清理操作记录
func IntervalDeleteActivity(ctx context.Context) (err error) {
	// 间隔时间，默认一天
	intervalTime := time.Hour * 24
	if config.GetGlobalConfig().Activity.Interval != "" {
		// 从配置文件里面解析间隔时间
		intervalTime, err = time.ParseDuration(config.GetGlobalConfig().Activity.Interval)
		if err != nil {
			return err
		}
	}
	timer := time.NewTicker(intervalTime)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			// 批量删除记录
			err = sqlstore.BatchDeleteActivity()
			if err != nil {
				return err
			}
		}
	}
}
