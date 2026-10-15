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

// Package events xxx
package events

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/dynamic"
	dbutils "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
)

// 业务方法

// CreateHashIndex 创建hash index
func CreateHashIndex(ctx context.Context, resourceType, indexName string) error {
	// hasIndex, err := dbutils.HasIndex(ctx, dbConfig, resourceType, indexName)
	hasIndex, err := dbutils.HasIndex(&dbutils.DBOperate{
		Context:      ctx,
		DBConfig:     dbConfig,
		IndexName:    indexName,
		ResourceType: resourceType,
	})
	if err != nil {
		return fmt.Errorf("failed to get index, err %s", err.Error())
	}

	// hash index 若已存在,则无需再创建
	if hasIndex {
		return nil
	}

	// 创建index
	index := drivers.Index{
		Name: indexName,
		Key: bson.D{
			bson.E{Key: indexName, Value: -1},
		},
	}

	// 创建index
	o := &dbutils.DBOperate{
		Context:      ctx,
		Index:        index,
		DBConfig:     dbConfig,
		ResourceType: resourceType,
	}
	if err = dbutils.CreateIndex(o); err != nil {
		return fmt.Errorf("failed to create index, err %s", err.Error())
	}

	return nil
}

// PushEvent push事件源
func PushEvent(data operator.M) error {
	// 克隆
	queueData := lib.CopyMap(data)
	queueData[resourceTypeTag] = eventResource

	env := typeofToString(queueData[envTag])

	extra, ok := queueData[extraInfoTag]
	if !ok {
		return nil
	}

	d, ok := extra.(types.EventExtraInfo)
	if !ok {
		return nil
	}

	queueData[nameSpaceTag] = d.Namespace
	queueData[resourceNameTag] = d.Name
	queueData[resourceKindTag] = func(env string) interface{} {
		switch env {
		case string(types.Event_Env_K8s):
			return data[kindTag]
		case string(types.Event_Env_Mesos):
			return d.Kind
		}
		return ""
	}(env)

	// queueFlag false
	if !apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		return nil
	}
	// queueFlag true
	return publishEventResourceToQueue(queueData, eventFeatTags, msgqueue.EventTypeUpdate)
}

// UpdateDynamicEvent 动态事件
func UpdateDynamicEvent(ctx context.Context, resourceType string, dynamicData operator.M) error {
	if dynamicData[dataTag] == nil {
		return nil
	}
	if _, ok := dynamicData[dataTag].(map[string]interface{}); !ok {
		return nil
	}
	// nolint
	dynamicData[namespaceTag] = dynamicData[dataTag].(map[string]interface{})["metadata"].(map[string]interface{})["namespace"].(string)
	// nolint
	dynamicData[resourceNameTag] = dynamicData[dataTag].(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
	dynamicData[resourceTypeTag] = eventResource

	features := make(operator.M)
	for _, key := range nsFeatTags {
		features[key] = dynamicData[key]
	}

	// dynamic event
	return dynamic.PutData(ctx, dynamicData, features, nsFeatTags, eventResource)
}

// AddEvent 新增事件
func AddEvent(ctx context.Context, resourceType string, data operator.M, opt *lib.StorePutOption) (err error) {
	// 参数
	dynamicData := lib.CopyMap(data)
	data[createTimeTag] = time.Now()
	// opt
	o := &dbutils.DBOperate{
		Context:      ctx,
		PutOpt:       opt,
		Data:         data,
		DBConfig:     dbConfig,
		ResourceType: resourceType,
	}
	// 添加数据
	if err = dbutils.PutData(o); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil
		}
		return fmt.Errorf("failed to insert, err %s", err.Error())
	}
	// update dynamic event
	if err = UpdateDynamicEvent(ctx, resourceType, dynamicData); err != nil {
		errMsg := fmt.Sprintf("dynamic event put data failed, err %s", err.Error())
		blog.Error(errMsg)
		return fmt.Errorf(errMsg)
	}
	// 创建哈希索引
	if err = CreateHashIndex(ctx, resourceType, eventTimeTag); err != nil {
		return err
	}
	// push事件源
	go func(data operator.M) {
		if err = PushEvent(data); err != nil {
			errMsg := fmt.Sprintf("publishEventResourceToQueue failed, err %s", err.Error())
			blog.Errorf(errMsg)
		}
	}(data)
	return nil
}

// GetEventList 获取 event list
func GetEventList(ctx context.Context, clusterIDs []string, opt *lib.StoreGetOption) (
	mList []operator.M, count int64, err error) {
	var o *dbutils.DBOperate
	for _, clusterID := range clusterIDs {
		// 查询或创建index
		if err = queryOrCreateIndex(ctx, clusterID); err != nil {
			return nil, 0, err
		}

		// 获取数据
		o = &dbutils.DBOperate{
			Context:      ctx,
			GetOpt:       opt,
			DBConfig:     dbConfig,
			ResourceType: TablePrefix + clusterID,
		}
		// nolint
		eList, err := dbutils.GetData(o)
		if err != nil {
			return nil, 0, errors.Wrapf(err, "get event list failed")
		}

		// 统计resourceType表
		o.SoftDeletion = true
		c, err := dbutils.Count(o)
		if err != nil {
			return nil, 0, errors.Wrapf(err, "count event list failed")
		}

		count += c
		mList = append(mList, eList...)
	}

	// 统计event表
	countEvent, err := dbutils.Count(&dbutils.DBOperate{
		Context:      ctx,
		GetOpt:       opt,
		SoftDeletion: true,
		DBConfig:     dbConfig,
		ResourceType: tableName,
	})
	if err != nil {
		return nil, 0, err
	}
	count += countEvent

	mListLength := int64(len(mList))
	if mListLength < opt.Limit {
		opt.Limit -= mListLength
		o := &dbutils.DBOperate{
			GetOpt:       opt,
			Context:      ctx,
			DBConfig:     dbConfig,
			ResourceType: tableName,
		}
		eventList, err := dbutils.GetData(o)
		if err != nil {
			return nil, 0, err
		}
		mList = append(mList, eventList...)
	} else if mListLength > opt.Limit {
		mList = mList[:opt.Limit]
	}
	lib.FormatTimeUTCRFC339(mList, []string{eventTimeTag, createTimeTag})

	return mList, count, nil
}

// GetStore 拿到当前接口，对应的数据库
func GetStore() *lib.Store {
	return lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig),
	)
}

// queryOrCreateIndex 查询或创建索引
func queryOrCreateIndex(ctx context.Context, clusterID string) error {
	for _, idxName := range eventQueryIndexKeys {
		hasIndex, err := dbutils.HasIndex(&dbutils.DBOperate{
			Context:      ctx,
			DBConfig:     dbConfig,
			ResourceType: TablePrefix + clusterID,
			IndexName:    idxName + "_idx",
		})
		if err != nil {
			return errors.Wrapf(err, "failed to get index for  clusterID(%s) with  %s", clusterID, idxName)
		}
		if hasIndex {
			continue
		}

		index := drivers.Index{
			Key:  bson.D{},
			Name: idxName + "_idx",
		}
		index.Key = append(index.Key, bson.E{Key: idxName, Value: 1})
		blog.Infof("create index for clusterID(%s) with key(%s)", clusterID, idxName)

		o := &dbutils.DBOperate{
			Context:      ctx,
			Index:        index,
			DBConfig:     dbConfig,
			ResourceType: TablePrefix + clusterID,
		}
		if err = dbutils.CreateIndex(o); err != nil {
			return errors.Wrapf(err, "failed to create index for clusterID(%s) with %s", clusterID, idxName)
		}
	}
	return nil
}
