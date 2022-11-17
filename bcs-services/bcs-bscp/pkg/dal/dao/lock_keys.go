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

package dao

import (
	"fmt"
	"strconv"

	"bscp.io/pkg/dal/table"
)

// lockKey is an instance of the keyFactory
var lockKey = new(lockKeyGenerator)

type lockKeyGenerator struct{}

// ConfigItem generate config item's lock
func (k lockKeyGenerator) ConfigItem(bizID uint32, appID uint32) *table.ResourceLock {
	return &table.ResourceLock{
		BizID:   bizID,
		ResType: string(table.ConfigItemTable),
		ResKey:  strconv.FormatInt(int64(appID), 10),
	}
}

// CurReleasedInst generate current released instance's lock ResKey
func (k lockKeyGenerator) CurReleasedInst(bizID uint32, appID uint32) *table.ResourceLock {
	return &table.ResourceLock{
		BizID:   bizID,
		ResType: string(table.CurrentReleasedInstanceTable),
		ResKey:  strconv.FormatInt(int64(appID), 10),
	}
}

// Strategy generate strategy's lock ResKey
func (k lockKeyGenerator) Strategy(bizID uint32, appID uint32) *table.ResourceLock {
	return &table.ResourceLock{
		BizID:   bizID,
		ResType: string(table.StrategyTable),
		ResKey:  strconv.FormatInt(int64(appID), 10),
	}
}

// DefaultStrategy generate default strategy's lock ResKey
func (k lockKeyGenerator) DefaultStrategy(bizID uint32, strategySetID uint32) *table.ResourceLock {
	return &table.ResourceLock{
		BizID:   bizID,
		ResType: string(table.StrategyTable),
		ResKey:  fmt.Sprintf("default-%d", strategySetID),
	}
}

// NamespaceStrategy generate namespace strategy's lock ResKey
func (k lockKeyGenerator) NamespaceStrategy(bizID uint32, strategySetID uint32, ns string) *table.ResourceLock {
	return &table.ResourceLock{
		BizID:   bizID,
		ResType: string(table.StrategyTable),
		ResKey:  fmt.Sprintf("namespace-%d-%s", strategySetID, ns),
	}
}

// StrategySet generate strategy set's lock ResKey
func (k lockKeyGenerator) StrategySet(bizID uint32, appID uint32) *table.ResourceLock {
	return &table.ResourceLock{
		BizID:   bizID,
		ResType: string(table.StrategySetTable),
		ResKey:  strconv.FormatInt(int64(appID), 10),
	}
}
