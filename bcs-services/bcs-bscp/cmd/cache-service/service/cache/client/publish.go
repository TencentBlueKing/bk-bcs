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

package client

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/cache-service/service/cache/keys"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// PublishInfo publish info
type PublishInfo struct {
	Key         string
	PublishTime int64
}

// GetPublishTime get publish time
func (c *client) GetPublishTime(kt *kit.Kit, publishTime int64) (map[uint32]PublishInfo, error) {
	result := make(map[uint32]PublishInfo)
	keys, err := c.bds.Keys(kt.Ctx, keys.Key.PublishPattern())
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		zValues, err := c.bds.ZRangeByScoreWithScores(kt.Ctx, key, &redis.ZRangeBy{
			Min: "1",
			Max: fmt.Sprintf("%d", publishTime),
		})
		if err != nil {
			return nil, err
		}
		for _, v := range zValues {
			fmt.Println(time.Unix(int64(v.Score), 0).Format(time.DateTime))
			result[getStrategyIdUint32(v.Member)] = PublishInfo{
				Key:         key,
				PublishTime: int64(v.Score),
			}
		}
	}

	return result, nil
}

// SetPublishTime set publish time
func (c *client) SetPublishTime(kt *kit.Kit, bizID, appID, strategyID uint32, publishTime int64) (int64, error) {
	return c.bds.ZAdd(kt.Ctx, keys.Key.PublishString(bizID, appID), float64(publishTime), strategyID)
}

func getStrategyIdUint32(v interface{}) uint32 {
	if value, ok := v.(string); ok {
		result, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return 0
		}
		return uint32(result)
	}
	return 0
}
