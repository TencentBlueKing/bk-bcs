/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package store

import (
	"context"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
)

const (
	// CreateTimeKey key for create time
	CreateTimeKey = "create_time"
	// BucketTimeKey key for bucket time
	BucketTimeKey = "bucket_time"
	// ObjectTypeKey key for object type
	ObjectTypeKey = "object_type"
	// ProjectIDKey key for project id
	ProjectIDKey = "project_id"
	// ClusterIDKey key for cluster id
	ClusterIDKey = "cluster_id"
	// NamespaceKey key for namespace
	NamespaceKey = "namespace"
	// WorkloadTypeKey key for workload type
	WorkloadTypeKey = "workload_type"
	// WorkloadNameKey key for workload name
	WorkloadNameKey = "workload_name"
	// DimensionKey key for time dimension
	DimensionKey = "dimension"
	// MetricTimeKey key for metric time
	MetricTimeKey = "metrics.time"
)

const (
	// DefaultPage default list page
	DefaultPage = 0
	// DefaultSize default list size
	DefaultSize = 10
)

// Public public model set
type Public struct {
	TableName           string
	Indexes             []drivers.Index
	DB                  drivers.DB
	IsTableEnsured      bool
	IsTableEnsuredMutex sync.RWMutex
}

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

func getStartTime(dimension string) time.Time {
	switch dimension {
	case common.DimensionDay:
		return time.Now().AddDate(0, 0, -14)
	case common.DimensionHour:
		return time.Now().Add((-48) * time.Hour)
	case common.DimensionMinute:
		return time.Now().Add((-60) * time.Minute)
	default:
		return time.Now()

	}
}

func distinctSlice(key string, slice *[]map[string]string) []string {
	tempResult := make([]string, 0)
	result := make([]string, 0)
	for _, value := range *slice {
		tempResult = append(tempResult, value[key])
	}
	temp := make(map[string]struct{})
	for _, value := range tempResult {
		if _, ok := temp[value]; !ok {
			temp[value] = struct{}{}
			result = append(result, value)
		}
	}
	return result
}

func getPublicData(ctx context.Context, db drivers.DB, cond *operator.Condition) *common.PublicData {
	result := &common.PublicData{}
	err := db.Table(common.DataTableNamePrefix+common.PublicTableName).Find(cond).One(ctx, result)
	if err != nil {
		blog.Errorf("get public data error: %v", err)
	}
	return result
}
