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

package hostconfig

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	dbutils "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
)

// PutData 更新/新增
func PutData(ctx context.Context, resourceType string, data operator.M, opt *lib.StorePutOption) error {
	return dbutils.PutData(&dbutils.DBOperate{
		PutOpt:       opt,
		Context:      ctx,
		Data:         data,
		DBConfig:     dbConfig,
		ResourceType: resourceType,
	})
}

// RemoveDta 删除数据
func RemoveDta(ctx context.Context, resourceType string, opt *lib.StoreRemoveOption) error {
	return dbutils.DeleteData(&dbutils.DBOperate{
		Context:      ctx,
		RemoveOpt:    opt,
		DBConfig:     dbConfig,
		ResourceType: resourceType,
	})
}

// GetData 查询数据
func GetData(ctx context.Context, resourceType string, opt *lib.StoreGetOption) ([]operator.M, error) {
	return dbutils.GetData(&dbutils.DBOperate{
		GetOpt:       opt,
		Context:      ctx,
		DBConfig:     dbConfig,
		ResourceType: resourceType,
	})
}

// UpdateMany update many
func UpdateMany(ctx context.Context, resourceType string, condition *operator.Condition, data interface{}) error {
	// 创建连接
	db := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig),
	)
	// 执行
	_, err := db.GetDB().Table(resourceType).UpdateMany(ctx, condition, data)
	return err

}

// 业务方法

// UpdateIP 更新ip
func UpdateIP(ctx context.Context, resourceType string, mList []operator.M, now time.Time,
	relation *types.BcsStorageClusterRelationIf) (err error) {
	var ipList []string
	for _, doc := range mList {
		ip, ok := doc[ipTag]
		if !ok {
			return fmt.Errorf("failed to QueryHost ip from %+v", doc)
		}
		ipStr, aok := ip.(string)
		if !aok {
			return fmt.Errorf("failed to parse ip from %+v", doc)
		}
		ipList = append(ipList, ipStr)
	}
	currentIPList := deduplicateStringSlice(ipList)
	// expectIpList is the ipList which match the clusterId we expected
	expectIPList := relation.Ips

	// insert the ip with clusterId="" which is not in db yet, preparing for next ops
	insertList := make([]operator.M, 0, len(expectIPList))
	for _, ip := range expectIPList {
		if !inList(ip, currentIPList) {
			insertList = append(insertList, operator.M{
				ipTag:         ip,
				clusterIDTag:  "",
				createTimeTag: now,
				updateTimeTag: now,
			})
		}
	}
	if len(insertList) <= 0 {
		return nil
	}

	// 创建连接
	db := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig),
	)
	// 更新数据库操作
	_, err = db.GetDB().Table(resourceType).Insert(ctx, []interface{}{insertList})
	return
}

// DoRelation 处理关系
func DoRelation(ctx context.Context, opt *lib.StoreGetOption, data operator.M, isPut bool,
	relation *types.BcsStorageClusterRelationIf) error {
	resourceType := tableName

	// 获取数据
	mList, err := GetData(ctx, resourceType, opt)
	if err != nil {
		return fmt.Errorf("failed to query, err %s", err.Error())
	}

	// 更新IP
	now := time.Now()
	if err = UpdateIP(ctx, resourceType, mList, now, relation); err != nil {
		return err
	}

	// put will clean the all cluster first, if not then just update
	if isPut {
		clusterID, ok := data[clusterIDTag].(string)
		if !ok {
			return fmt.Errorf("cannot parse clusterID from %+v", data)
		}

		params := operator.M{
			clusterIDTag:  "",
			updateTimeTag: time.Now(),
		}
		condition := operator.NewLeafCondition(
			operator.Eq,
			operator.M{
				clusterIDTag: clusterID,
			},
		)
		// cleanCluster
		if err = UpdateMany(ctx, resourceType, condition, params); err != nil {
			return err
		}
	}

	data.Update(updateTimeTag, now)
	return UpdateMany(ctx, resourceType, opt.Cond, data)
}

// QueryHost query host
func QueryHost(ctx context.Context, condition *operator.Condition) ([]operator.M, error) {
	// option
	opt := &lib.StoreGetOption{
		Cond: condition,
	}

	// 查询数据
	return GetData(ctx, tableName, opt)
}

// PutHostToDB put host to db
func PutHostToDB(ctx context.Context, data operator.M, cond *operator.Condition) error {
	// option
	opt := &lib.StorePutOption{
		UniqueKey:     indexKeys,
		Cond:          cond,
		CreateTimeKey: createTimeTag,
		UpdateTimeKey: updateTimeTag,
	}
	return PutData(ctx, tableName, data, opt)
}

// RemoveHost remove host
func RemoveHost(ctx context.Context, condition *operator.Condition) error {
	// option
	opt := &lib.StoreRemoveOption{
		Cond: condition,
	}
	// 移除
	return RemoveDta(ctx, tableName, opt)
}
