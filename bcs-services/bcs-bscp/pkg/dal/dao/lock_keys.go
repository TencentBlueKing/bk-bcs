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

// Package dao NOTES
//
//nolint:unused
package dao

import (
	"strconv"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
)

// lockKey is an instance of the keyFactory
var lockKey = new(lockKeyGenerator)

type lockKeyGenerator struct{}

// ConfigItem generate config item's lock
func (k lockKeyGenerator) ConfigItem(bizID uint32, appID uint32) *table.ResourceLock {
	return &table.ResourceLock{
		BizID:   bizID,
		ResType: table.ConfigItemTable.String(),
		ResKey:  strconv.FormatInt(int64(appID), 10),
	}
}

// Group generate group's lock ResKey
func (k lockKeyGenerator) Group(bizID uint32, appID uint32) *table.ResourceLock {
	return &table.ResourceLock{
		BizID:   bizID,
		ResType: table.GroupTable.String(),
		ResKey:  strconv.FormatInt(int64(appID), 10),
	}
}
