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

package mongo

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	datamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

var (
	modelWorkloadOriginRequestIndexes = []drivers.Index{
		{
			Key: bson.D{
				bson.E{Key: CreateTimeKey, Value: 1},
			},
			Name: CreateTimeKey + "_1",
		},
		{
			Name: types.WorkloadTableName + "_idx",
			Key: bson.D{
				bson.E{Key: ProjectIDKey, Value: 1},
				bson.E{Key: ClusterIDKey, Value: 1},
				bson.E{Key: NamespaceKey, Value: 1},
				bson.E{Key: WorkloadTypeKey, Value: 1},
				bson.E{Key: WorkloadNameKey, Value: 1},
				bson.E{Key: BucketTimeKey, Value: 1},
			},
			Background: true,
		},
		{
			Key: bson.D{
				bson.E{Key: ClusterIDKey, Value: 1},
			},
			Name:       ClusterIDKey + "_1",
			Background: true,
		},
		{
			Key: bson.D{
				bson.E{Key: ProjectIDKey, Value: 1},
			},
			Name:       ProjectIDKey + "_1",
			Background: true,
		},
		{
			Key: bson.D{
				bson.E{Key: NamespaceKey, Value: 1},
			},
			Name:       NamespaceKey + "_1",
			Background: true,
		},
		{
			Key: bson.D{
				bson.E{Key: WorkloadNameKey, Value: 1},
			},
			Name:       WorkloadNameKey + "_1",
			Background: true,
		},
		{
			Key: bson.D{
				bson.E{Key: WorkloadTypeKey, Value: 1},
			},
			Name:       WorkloadTypeKey + "_1",
			Background: true,
		},
	}
)

// ModelWorkloadOriginRequest workload request model
type ModelWorkloadOriginRequest struct {
	Public
}

// NewModelWorkloadOriginRequest new workload model
func NewModelWorkloadOriginRequest(db drivers.DB) *ModelWorkloadOriginRequest {
	return &ModelWorkloadOriginRequest{Public: Public{
		TableName: types.DataTableNamePrefix + types.WorkloadInfoTableName,
		Indexes:   modelWorkloadOriginRequestIndexes,
		DB:        db,
	}}
}

// CreateWorkloadOriginRequest create workload origin request
func (m *ModelWorkloadOriginRequest) CreateWorkloadOriginRequest(ctx context.Context,
	result *types.WorkloadOriginRequestResult) error {
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		ProjectIDKey:    result.ProjectID,
		ClusterIDKey:    result.ClusterID,
		NamespaceKey:    result.Namespace,
		WorkloadTypeKey: result.WorkloadType,
		WorkloadNameKey: result.WorkloadName,
	})
	return m.DB.Table(m.TableName).Upsert(ctx, cond, operator.M{"$set": result})
}

// ListWorkloadOriginRequest list workload origin request
func (m *ModelWorkloadOriginRequest) ListWorkloadOriginRequest(ctx context.Context,
	req *datamanager.GetWorkloadOriginRequestResultReq) ([]*datamanager.WorkloadOriginRequestResult, error) {
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	result := make([]*types.WorkloadOriginRequestResult, 0)
	cond := m.generateCond(req)
	pipeline := make([]map[string]interface{}, 0)
	pipeline = append(pipeline, map[string]interface{}{"$match": cond},
		map[string]interface{}{"$sort": map[string]interface{}{
			CreateTimeKey: -1,
		}},
		map[string]interface{}{"$group": map[string]interface{}{
			"_id": map[string]interface{}{ClusterIDKey: "$cluster_id",
				NamespaceKey:    "$namespace",
				WorkloadTypeKey: "$workload_type",
				WorkloadNameKey: "$workload_name"},
			"project_id":    map[string]interface{}{"$first": "$project_id"},
			"cluster_id":    map[string]interface{}{"$first": "$cluster_id"},
			"namespace":     map[string]interface{}{"$first": "$namespace"},
			"workload_type": map[string]interface{}{"$first": "$workload_type"},
			"workload_name": map[string]interface{}{"$first": "$workload_name"},
			"cpu":           map[string]interface{}{"$first": "$cpu"},
			"memory":        map[string]interface{}{"$first": "$memory"},
			"create_time":   map[string]interface{}{"$first": "$create_time"},
		}},
	)
	err = m.DB.Table(m.TableName).Aggregation(ctx, pipeline, &result)
	if err != nil {
		blog.Errorf("find workload origin request data fail, err:%v", err)
		return nil, err
	}
	blog.Infof("%d", len(result))
	return m.generateWorkloadRequestRsp(result), nil
}

func (m *ModelWorkloadOriginRequest) generateCond(
	req *datamanager.GetWorkloadOriginRequestResultReq) map[string]interface{} {
	cond := make(map[string]interface{})
	if req.ProjectID != "" {
		cond[ProjectIDKey] = req.ProjectID
	}
	if req.ClusterID != "" {
		cond[ClusterIDKey] = req.ClusterID
	}
	if req.Namespace != "" {
		cond[NamespaceKey] = req.Namespace
	}
	if req.WorkloadType != "" {
		cond[WorkloadTypeKey] = req.WorkloadType
	}
	if req.WorkloadName != "" {
		cond[WorkloadNameKey] = req.WorkloadName
	}
	return cond
}

func (m *ModelWorkloadOriginRequest) generateWorkloadRequestRsp(
	origin []*types.WorkloadOriginRequestResult) []*datamanager.WorkloadOriginRequestResult {
	resultList := make([]*datamanager.WorkloadOriginRequestResult, 0)
	for _, originResult := range origin {
		result := &datamanager.WorkloadOriginRequestResult{
			ClusterID:    originResult.ClusterID,
			Namespace:    originResult.Namespace,
			WorkloadType: originResult.WorkloadType,
			WorkloadName: originResult.WorkloadName,
			Cpu:          originResult.Cpu,
			Memory:       originResult.Memory,
			ProjectID:    originResult.ProjectID,
		}
		resultList = append(resultList, result)
	}
	return resultList
}
