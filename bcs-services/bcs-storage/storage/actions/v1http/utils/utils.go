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

// Package utils xxx
package utils

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
	"github.com/emicklei/go-restful"
	"github.com/opentracing/opentracing-go"
)

// SetHTTPSpanContextInfo set restful.Request context
func SetHTTPSpanContextInfo(req *restful.Request, handler string) opentracing.Span {
	span, ctx := utils.StartSpanFromContext(req.Request.Context(), handler)
	utils.HTTPTagHandler.Set(span, handler)
	req.Request = req.Request.WithContext(ctx)

	return span
}

// CreateIndex 创建索引
func CreateIndex(ctx context.Context, dbConfig, resourceType string, index drivers.Index) error {
	// 创建连接
	db := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig),
	)

	return db.CreateIndex(ctx, resourceType, index)
}

// DeleteData 移除
func DeleteData(ctx context.Context, dbConfig, resourceType string, opt *lib.StoreRemoveOption) error {
	// 创建连接
	db := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig),
	)
	db.SetSoftDeletion(true)

	return db.Remove(ctx, resourceType, opt)
}

// DeleteBatchData 批量删除
func DeleteBatchData(ctx context.Context, dbConfig, resourceType string, getOption *lib.StoreGetOption,
	rmOption *lib.StoreRemoveOption, needTimeFormatList []string) ([]operator.M, error) {
	mList, err := GetData(ctx, dbConfig, resourceType, getOption)
	if err != nil {
		return nil, err
	}

	if len(needTimeFormatList) > 0 {
		lib.FormatTime(mList, needTimeFormatList)
	}

	err = DeleteData(ctx, dbConfig, resourceType, rmOption)
	if err != nil {
		return nil, err
	}

	return mList, nil
}

// DeleteIndex 删除索引
func DeleteIndex(ctx context.Context, dbConfig, resourceType string, indexName string) error {
	// 创建连接
	db := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig),
	)

	return db.DeleteIndex(ctx, resourceType, indexName)
}

// PutData 新增
func PutData(ctx context.Context, dbConfig, resourceType string, data operator.M, opt *lib.StorePutOption) error {
	// 创建连接
	db := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig),
	)

	return db.Put(ctx, resourceType, data, opt)
}

// GetData 查询数据
func GetData(ctx context.Context, dbConfig, resourceType string, opt *lib.StoreGetOption) ([]operator.M, error) {
	// 创建连接
	db := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig),
	)

	return db.Get(ctx, resourceType, opt)
}

// Count 统计
func Count(ctx context.Context, dbConfig, resourceType string, opt *lib.StoreGetOption) (int64, error) {
	// 创建连接
	db := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig),
	)
	db.SetSoftDeletion(true)

	return db.Count(ctx, resourceType, opt)
}

// HasIndex 是否有index
func HasIndex(ctx context.Context, dbConfig, resourceType, indexName string) (bool, error) {
	// 创建连接
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig),
	)

	return store.GetDB().Table(resourceType).HasIndex(ctx, indexName)
}

// HasTable 是否有table
func HasTable(ctx context.Context, dbConfig, resourceType string) (bool, error) {
	// 创建连接
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig),
	)

	return store.GetDB().HasTable(ctx, resourceType)
}

// GetIndex index
func GetIndex(ctx context.Context, dbConfig, resourceType string) (*drivers.Index, error) {
	// Obtain table index
	db := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig),
	)
	db.SetSoftDeletion(true)

	return db.GetIndex(ctx, resourceType)
}

// GetList 获取list
func GetList(ctx context.Context, dbConfig string) ([]string, error) {
	// 创建连接
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig),
	)
	store.SetSoftDeletion(true)

	return store.GetDB().ListTableNames(ctx)
}
