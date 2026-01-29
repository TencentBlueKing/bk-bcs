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

// Package cluster xxx
package cluster

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	dbutils "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/utils"
)

// db方法

// PutData put data to db
func PutData(ctx context.Context, data, features operator.M, featTags []string) error {
	opt := &lib.StorePutOption{
		UniqueKey:     featTags,
		Cond:          operator.NewLeafCondition(operator.Eq, features),
		CreateTimeKey: createTimeTag,
		UpdateTimeKey: updateTimeTag,
	}
	return dbutils.PutData(&dbutils.DBOperate{
		PutOpt:       opt,
		Context:      ctx,
		Data:         data,
		SoftDeletion: false,
		ResourceType: tableCluster,
		DBConfig:     dbConfig,
	})
}

// DeleteBatchData 批量删除
func DeleteBatchData(ctx context.Context, getOption *lib.StoreGetOption,
	rmOpt *lib.StoreRemoveOption) ([]operator.M, error) {
	return dbutils.DeleteBatchData(&dbutils.DBOperate{
		Context:      ctx,
		SoftDeletion: false,
		RemoveOpt:    rmOpt,
		DBConfig:     dbConfig,
		GetOpt:       getOption,
		ResourceType: tableCluster,
	})
}
