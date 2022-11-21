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

package metric

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	dbutils "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/utils"
)

// PutData 更新/新增
func PutData(ctx context.Context, resourceType string, data operator.M, opt *lib.StorePutOption) error {
	return dbutils.PutData(&dbutils.DBOperate{
		Context:      ctx,
		PutOpt:       opt,
		Data:         data,
		SoftDeletion: true,
		DBConfig:     dbConfig,
		ResourceType: resourceType,
	})
}

// RemoveData 删除数据
func RemoveData(ctx context.Context, resourceType string, opt *lib.StoreRemoveOption) error {
	return dbutils.DeleteData(&dbutils.DBOperate{
		RemoveOpt:    opt,
		Context:      ctx,
		SoftDeletion: true,
		DBConfig:     dbConfig,
		ResourceType: resourceType,
	})
}

// GetData 查询数据
func GetData(ctx context.Context, resourceType string, opt *lib.StoreGetOption) ([]operator.M, error) {
	mList, err := dbutils.GetData(&dbutils.DBOperate{
		Context:      ctx,
		GetOpt:       opt,
		SoftDeletion: true,
		DBConfig:     dbConfig,
		ResourceType: resourceType,
	})
	lib.FormatTime(mList, []string{createTimeTag, updateTimeTag})
	return mList, err
}

// GetList 获取list数据
func GetList(ctx context.Context) ([]string, error) {
	return dbutils.GetList(ctx, dbConfig)
}
