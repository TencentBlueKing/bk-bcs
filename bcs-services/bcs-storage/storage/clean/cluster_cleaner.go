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

// Package clean xxx
package clean

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/types"
)

const (
	// MongoDB 数据库路径常量
	mongodbDynamicDB = "mongodb/dynamic"
	mongodbEventDB   = "mongodb/event"
)

// ClusterCleaner 集群数据清理器
type ClusterCleaner struct {
	dbClients map[string]drivers.DB // 数据库连接 map，key 为数据库名称
}

// NewClusterCleaner 创建集群清理器
func NewClusterCleaner() *ClusterCleaner {
	// 只清理 dynamic 和 event 两个数据库
	dbClients := map[string]drivers.DB{
		databaseDynamic: apiserver.GetAPIResource().GetDBClient(mongodbDynamicDB),
		databaseEvent:   apiserver.GetAPIResource().GetDBClient(mongodbEventDB),
	}

	return &ClusterCleaner{
		dbClients: dbClients,
	}
}

// CleanClusterData 清理集群数据（并发清理多个数据库）
func (cc *ClusterCleaner) CleanClusterData(ctx context.Context, clusterID string) (map[string]int64, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("clusterId is empty")
	}

	deletedCounts := make(map[string]int64)
	var mu sync.Mutex     // 保护 deletedCounts 和 errors
	var wg sync.WaitGroup // 等待所有 goroutine 完成
	var errors []error    // 收集所有错误

	// 并发清理各个数据库
	for dbName, dbClient := range cc.dbClients {
		wg.Add(1)
		go func(name string, client drivers.DB) {
			defer wg.Done()

			count, err := cc.deleteByCluster(ctx, client, clusterID)
			if err != nil {
				blog.Errorf("clean cluster %s database[%s] failed: %v", clusterID, name, err)
				mu.Lock()
				errors = append(errors, fmt.Errorf("database[%s]: %v", name, err))
				mu.Unlock()
				return
			}

			mu.Lock()
			deletedCounts[name] = count
			mu.Unlock()

			blog.Infof("clean cluster %s database[%s]: deleted %d records", clusterID, name, count)
		}(dbName, dbClient)
	}

	// 等待所有清理任务完成
	wg.Wait()

	// 检查是否有错误，如果有则返回聚合的错误信息
	if len(errors) > 0 {
		var errMsg string
		for i, err := range errors {
			if i > 0 {
				errMsg += "; "
			}
			errMsg += err.Error()
		}
		return deletedCounts, fmt.Errorf("clean cluster failed: %s", errMsg)
	}

	blog.Infof("clean cluster %s data completed, deleted: %v", clusterID, deletedCounts)
	return deletedCounts, nil
}

// deleteByCluster 删除集群数据
func (cc *ClusterCleaner) deleteByCluster(ctx context.Context, db drivers.DB, clusterID string) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("database client is nil")
	}

	// 获取该数据库下的所有表
	tables, err := cc.getTablesForDataType(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("get tables failed: %v", err)
	}

	var totalCount int64
	for _, tableName := range tables {
		// 构造条件：clusterId = clusterID
		cond := operator.NewLeafCondition(operator.Eq, operator.M{
			types.TagClusterID: clusterID,
		})

		count, err := db.Table(tableName).Delete(ctx, cond)
		if err != nil {
			return totalCount, fmt.Errorf("delete table %s failed: %v", tableName, err)
		}
		totalCount += count
	}

	return totalCount, nil
}

// getTablesForDataType 获取数据类型对应的表列表
func (cc *ClusterCleaner) getTablesForDataType(ctx context.Context, db drivers.DB) ([]string, error) {
	tables, err := db.ListTableNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tables failed: %v", err)
	}

	blog.Infof("found %d tables in database", len(tables))
	return tables, nil
}
