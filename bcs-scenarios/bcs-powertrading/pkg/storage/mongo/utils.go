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

package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
)

const (
	// defaultPage = 0
	defaultSize         = 10
	tableNamePrefix     = "bcs_powertrading"
	taskTableName       = "task"
	deviceDataTableName = "deviceData"
)

const (
	taskIDKey    = "taskID"
	taskTypeKey  = "type"
	isDeletedKey = "isDeleted"
	deviceIDKey  = "deviceID"
	checkTimeKey = "checkTime"
)

func ensureTable(ctx context.Context, public *Public) error {
	public.IsTableEnsuredMutex.RLock()
	if public.IsTableEnsured {
		public.IsTableEnsuredMutex.RUnlock()
		return nil
	}
	if err := ensure(ctx, public.DB, public.TableName, public.Indexes); err != nil {
		public.IsTableEnsuredMutex.RUnlock()
		return err
	}
	public.IsTableEnsuredMutex.RUnlock()

	public.IsTableEnsuredMutex.Lock()
	public.IsTableEnsured = true
	public.IsTableEnsuredMutex.Unlock()
	return nil
}

// ensure xxx
// EnsureTable ensure object database table and table indexes
func ensure(ctx context.Context, db drivers.DB, tableName string, indexes []drivers.Index) error {
	hasTable, err := db.HasTable(ctx, tableName)
	if err != nil {
		return err
	}
	if !hasTable {
		tErr := db.CreateTable(ctx, tableName)
		if tErr != nil {
			return tErr
		}
	}
	// only ensure index when index name is not empty
	for _, idx := range indexes {
		hasIndex, iErr := db.Table(tableName).HasIndex(ctx, idx.Name)
		if iErr != nil {
			return iErr
		}
		if !hasIndex {
			if iErr = db.Table(tableName).CreateIndex(ctx, idx); iErr != nil {
				return iErr
			}
		}
	}
	return nil
}

// MergePatch 合并更新
// 如果overwriteZeroOrEmptyStr为true，当modified的int类型字段为0时，更新为0；
// 当modified的string类型字段为""时，更新为""；否则跳过
// 如果为nil，更新为nil
func MergePatch(original, modified interface{}, overwriteZeroOrEmptyStr bool) ([]byte, error) {
	var originalByte, modifiedByte []byte
	originalByte, err := json.Marshal(original)
	if err != nil {
		return nil, fmt.Errorf("marshal interface{} to []byte error:%v", err)
	}
	modifiedByte, err = json.Marshal(modified)
	if err != nil {
		return nil, fmt.Errorf("marshal interface{} to []byte error:%v", err)
	}
	originalMap := map[string]interface{}{}
	if len(originalByte) > 0 {
		if err := json.Unmarshal(originalByte, &originalMap); err != nil {
			return nil, fmt.Errorf("unmarshal original to map error:%v", err)
		}
	}

	modifiedMap := map[string]interface{}{}
	if len(modifiedByte) > 0 {
		if err := json.Unmarshal(modifiedByte, &modifiedMap); err != nil {
			return nil, fmt.Errorf("unmarshal modified to map error:%v", err)
		}
	}
	for key, modifiedValue := range modifiedMap {
		if valueTime, ok := modifiedValue.(time.Time); ok {
			if valueTime.IsZero() && !overwriteZeroOrEmptyStr {
				continue
			}
			originalMap[key] = valueTime
			continue
		}
		if valueStr, ok := modifiedValue.(string); ok {
			if valueStr == "-" {
				originalMap[key] = ""
				continue
			} else if (valueStr == "" || valueStr == "0001-01-01T00:00:00Z") && !overwriteZeroOrEmptyStr {
				continue
			}
			originalMap[key] = valueStr
			continue
		}
		if valueInt, ok := modifiedValue.(int); ok {
			if valueInt == 0 && !overwriteZeroOrEmptyStr {
				continue
			}
			originalMap[key] = valueInt
			continue
		}
		if valueFloat, ok := modifiedValue.(float64); ok {
			if valueFloat == float64(0) && !overwriteZeroOrEmptyStr {
				continue
			}
			originalMap[key] = valueFloat
			continue
		}
		if modifiedValue == nil && !overwriteZeroOrEmptyStr {
			continue
		}
		originalMap[key] = modifiedValue
	}
	return json.Marshal(originalMap)
}
