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
 *
 */

package dynamic

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	dbutils "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/utils"
)

// db方法

func GetData(ctx context.Context, resourceType string, opt *lib.StoreGetOption) ([]operator.M, error) {
	if resourceType == eventResourceType {
		resourceType = eventDBConfig
	}

	return dbutils.GetData(ctx, dbConfig, resourceType, opt)
}

// PutData put data to db
func PutData(ctx context.Context, data, features operator.M, resourceFeatList []string, table string) error {
	if table == eventResourceType {
		table = eventDBConfig
	}
	opt := &lib.StorePutOption{
		UniqueKey:     resourceFeatList,
		Cond:          operator.NewLeafCondition(operator.Eq, features),
		CreateTimeKey: createTimeTag,
		UpdateTimeKey: updateTimeTag,
	}
	return dbutils.PutData(ctx, dbConfig, table, data, opt)
}

// DeleteBatchData 批量删除
func DeleteBatchData(ctx context.Context, resourceType string, getOption *lib.StoreGetOption,
	rmOption *lib.StoreRemoveOption) ([]operator.M, error) {
	return dbutils.DeleteBatchData(ctx, dbConfig, resourceType, getOption, rmOption, needTimeFormatList)
}

// Count 统计
func Count(ctx context.Context, resourceType string, opt *lib.StoreGetOption) (int64, error) {
	return dbutils.Count(ctx, dbConfig, resourceType, opt)
}

// GetIndex index
func GetIndex(ctx context.Context, resourceType string) (*drivers.Index, error) {
	return dbutils.GetIndex(ctx, dbConfig, resourceType)
}

// CreateIndex 创建索引
func CreateIndex(ctx context.Context, resourceType string, index drivers.Index) error {
	// 创建连接
	return dbutils.CreateIndex(ctx, dbConfig, resourceType, index)
}

// DeleteIndex 删除索引
func DeleteIndex(ctx context.Context, resourceType string, indexName string) error {
	return dbutils.DeleteIndex(ctx, dbConfig, resourceType, indexName)
}

// 业务方法

// GetDataWithPageInfo 分页查询
func GetDataWithPageInfo(ctx context.Context, resourceType string, opt *lib.StoreGetOption) (data []operator.M, extra operator.M, err error) {
	if resourceType == eventResourceType {
		resourceType = eventDBConfig
	}

	count, err := Count(ctx, resourceType, opt)
	if err != nil {
		return nil, nil, err
	}

	mList, err := GetData(ctx, resourceType, opt)
	if err != nil {
		return nil, nil, err
	}
	lib.FormatTime(mList, needTimeFormatList)

	extra = operator.M{
		"total":    count,
		"pageSize": opt.Limit,
		"offset":   opt.Offset,
	}
	return mList, extra, err
}

// PutCustomResourceToDB 保存 custom resources 到数据库
func PutCustomResourceToDB(ctx context.Context, resourceType string, data operator.M, opt *lib.StorePutOption) error {
	index, err := GetIndex(ctx, resourceType)
	if err != nil {
		return err
	}

	var uniIdx drivers.Index
	if index != nil {
		uniIdx = *index
	}

	condition := make([]*operator.Condition, 0)
	if len(uniIdx.Key) != 0 {
		for _, bsonElem := range uniIdx.Key {
			key := bsonElem.Key
			condition = append(condition, operator.NewLeafCondition(operator.Eq, operator.M{key: data[key]}))
		}
	}
	if len(condition) != 0 {
		opt.Cond = operator.NewBranchCondition(operator.And, condition...)
	}

	return dbutils.PutData(ctx, dbConfig, resourceType, data, opt)
}

// CreateCustomResourceIndex 创建 Custom Resources 索引
func CreateCustomResourceIndex(ctx context.Context, resourceType string, index drivers.Index) error {
	return CreateIndex(ctx, resourceType, index)
}

//DeleteCustomResourceIndex 删除 Custom Resources 索引
func DeleteCustomResourceIndex(ctx context.Context, resourceType string, indexName string) error {
	return DeleteIndex(ctx, resourceType, indexName)
}
