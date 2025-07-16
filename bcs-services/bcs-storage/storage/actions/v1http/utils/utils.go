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

// Package utils xxx
package utils

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/utils"
	restful "github.com/emicklei/go-restful/v3"
	"github.com/opentracing/opentracing-go"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
)

// DBOperate DB operate options struct
type DBOperate struct {
	Data               operator.M
	Context            context.Context
	DBConfig           string
	IndexName          string
	ResourceType       string
	SoftDeletion       bool
	NeedTimeFormatList []string

	Index     drivers.Index
	GetOpt    *lib.StoreGetOption
	PutOpt    *lib.StorePutOption
	RemoveOpt *lib.StoreRemoveOption
}

// SetHTTPSpanContextInfo set restful.Request context
func SetHTTPSpanContextInfo(req *restful.Request, handler string) opentracing.Span {
	span, ctx := utils.StartSpanFromContext(req.Request.Context(), handler)
	utils.HTTPTagHandler.Set(span, handler)
	req.Request = req.Request.WithContext(ctx)

	return span
}

// CreateIndex 创建索引
func CreateIndex(o *DBOperate) error {
	// 创建连接
	db := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(o.DBConfig),
		apiserver.GetAPIResource().GetEventBus(o.DBConfig),
	)
	db.SetSoftDeletion(o.SoftDeletion)

	return db.CreateIndex(o.Context, o.ResourceType, o.Index)
}

// DeleteData 移除
// func DeleteData(ctx context.Context, dbConfig, resourceType string, opt *lib.StoreRemoveOption) error {
func DeleteData(o *DBOperate) error {
	// 创建连接
	db := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(o.DBConfig),
		apiserver.GetAPIResource().GetEventBus(o.DBConfig),
	)
	db.SetSoftDeletion(o.SoftDeletion)

	return db.Remove(o.Context, o.ResourceType, o.RemoveOpt)
}

// DeleteBatchData 批量删除
// func DeleteBatchData(ctx context.Context, dbConfig, resourceType string, getOption *lib.StoreGetOption,
//
//	rmOption *lib.StoreRemoveOption, needTimeFormatList []string
func DeleteBatchData(o *DBOperate) ([]operator.M, error) {
	mList, err := GetData(o)
	if err != nil {
		return nil, err
	}
	if len(o.NeedTimeFormatList) > 0 {
		lib.FormatTime(mList, o.NeedTimeFormatList)
	}

	if err = DeleteData(o); err != nil {
		return nil, err
	}
	return mList, nil
}

// DeleteIndex 删除索引
func DeleteIndex(o *DBOperate) error {
	// 创建连接
	db := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(o.DBConfig),
		apiserver.GetAPIResource().GetEventBus(o.DBConfig),
	)
	db.SetSoftDeletion(o.SoftDeletion)

	return db.DeleteIndex(o.Context, o.ResourceType, o.IndexName)
}

// PutData 新增
func PutData(o *DBOperate) error {
	// 创建连接
	db := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(o.DBConfig),
		apiserver.GetAPIResource().GetEventBus(o.DBConfig),
	)
	db.SetSoftDeletion(o.SoftDeletion)

	return db.Put(o.Context, o.ResourceType, o.Data, o.PutOpt)
}

// GetData 查询数据
func GetData(o *DBOperate) ([]operator.M, error) {
	// 创建连接
	db := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(o.DBConfig),
		apiserver.GetAPIResource().GetEventBus(o.DBConfig),
	)
	db.SetSoftDeletion(o.SoftDeletion)

	return db.Get(o.Context, o.ResourceType, o.GetOpt)
}

// Count 统计
func Count(o *DBOperate) (int64, error) {
	// 创建连接
	db := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(o.DBConfig),
		apiserver.GetAPIResource().GetEventBus(o.DBConfig),
	)
	db.SetSoftDeletion(o.SoftDeletion)

	return db.Count(o.Context, o.ResourceType, o.GetOpt)
}

// HasIndex 是否有index
func HasIndex(o *DBOperate) (bool, error) {
	// 创建连接
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(o.DBConfig),
		apiserver.GetAPIResource().GetEventBus(o.DBConfig),
	)
	store.SetSoftDeletion(o.SoftDeletion)

	return store.GetDB().Table(o.ResourceType).HasIndex(o.Context, o.IndexName)
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
func GetIndex(o *DBOperate) (*drivers.Index, error) {
	// Obtain table index
	db := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(o.DBConfig),
		apiserver.GetAPIResource().GetEventBus(o.DBConfig),
	)
	db.SetSoftDeletion(o.SoftDeletion)

	return db.GetIndex(o.Context, o.ResourceType)
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
