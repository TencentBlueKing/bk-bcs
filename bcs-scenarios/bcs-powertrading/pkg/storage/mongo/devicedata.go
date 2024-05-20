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

// Package mongo xxx
package mongo

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/storage"
)

var (
	modelDeviceDataIndexes = []drivers.Index{
		{
			Name: deviceDataTableName + "_idx",
			Key: bson.D{
				bson.E{Key: deviceIDKey, Value: 1},
				bson.E{Key: checkTimeKey, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelDeviceData defines deviceData
type ModelDeviceData struct {
	Public
}

// NewModelDeviceData returns a new ModelDeviceData
func NewModelDeviceData(db drivers.DB) *ModelDeviceData {
	return &ModelDeviceData{Public{
		TableName: tableNamePrefix + "_" + deviceDataTableName,
		Indexes:   modelDeviceDataIndexes,
		DB:        db,
	}}
}

// CreateDeviceData create device data
func (m *ModelDeviceData) CreateDeviceData(ctx context.Context, data *storage.DeviceOperationData,
	opt *storage.CreateOptions) error {
	if opt == nil {
		return fmt.Errorf("CreateOption is nil")
	}
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		deviceIDKey:  data.DeviceID,
		checkTimeKey: data.CheckTime,
	})
	retRecord := &storage.DeviceOperationData{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, retRecord); err != nil {
		// 如果查不到，创建
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			_, err = m.DB.Table(m.TableName).Insert(ctx, []interface{}{data})
			if err != nil {
				return fmt.Errorf("task does not exist, insert error: %v", err)
			}
			return nil
		}
		return fmt.Errorf("find data record error: %v", err)
	}
	// 如果查到，且opt.OverWriteIfExist为true，更新
	if !opt.OverWriteIfExist {
		return fmt.Errorf("data record exists")
	}
	if err := m.DB.Table(m.TableName).Update(ctx, cond, operator.M{"$set": data}); err != nil {
		return fmt.Errorf("update data record error: %v", err)
	}
	return nil
}
