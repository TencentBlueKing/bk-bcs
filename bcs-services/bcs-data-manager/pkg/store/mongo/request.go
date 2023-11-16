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

package mongo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	datamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

var (
	modelWorkloadRequestIndexes = []drivers.Index{
		{
			Key: bson.D{
				bson.E{Key: CreateTimeKey, Value: 1},
			},
			Name: CreateTimeKey + "_1",
		},
	}
)

// ModelWorkloadRequest defines predict result
type ModelWorkloadRequest struct {
	Public
}

// NewModelWorkloadRequest returns a new ModelCPURequest
func NewModelWorkloadRequest(db drivers.DB) *ModelWorkloadRequest {
	return &ModelWorkloadRequest{Public{
		TableName: types.PredictTableNamePrefix + types.WorkloadRequestTableName,
		Indexes:   modelWorkloadRequestIndexes,
		DB:        db,
	}}
}

func (m *ModelWorkloadRequest) GetLatestWorkloadRequest(ctx context.Context,
	req *datamanager.GetWorkloadRequestRecommendResultReq) (*datamanager.GetWorkloadRequestRecommendResultRsp, error) {
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	metricSlice := []string{"container_memory_working_set_bytes", "container_cpu_usage_seconds_total_rate_2m"}
	result := make([]*types.BKBaseRequestRecommendResult, 0)
	for _, metric := range metricSlice {
		metricResult := make([]*types.BKBaseRequestRecommendResult, 0)
		cond := m.generateCond(req)
		cond[MetricKey] = metric
		tableName := "bcs_predict_request_cpu"
		if strings.Contains(metric, "memory") {
			tableName = "bcs_predict_request_memory"
		}
		pipeline := make([]map[string]interface{}, 0)
		pipeline = append(pipeline, map[string]interface{}{"$match": cond},
			map[string]interface{}{"$sort": map[string]interface{}{
				DTEventTimeStampKey: -1,
			}},
			map[string]interface{}{"$group": map[string]interface{}{
				"_id": map[string]interface{}{BCSClusterIDKey: "$bcs_cluster_id",
					NamespaceKey:     "$namespace",
					WorkloadKindKey:  "$workload_kind",
					WorkloadNameKey:  "$workload_name",
					ContainerNameKey: "$container_name"},
				"bcs_cluster_id":      map[string]interface{}{"$first": "$bcs_cluster_id"},
				"namespace":           map[string]interface{}{"$first": "$namespace"},
				"workload_kind":       map[string]interface{}{"$first": "$workload_kind"},
				"workload_name":       map[string]interface{}{"$first": "$workload_name"},
				"container_name":      map[string]interface{}{"$first": "$container_name"},
				"p90":                 map[string]interface{}{"$first": "$p90"},
				"p99":                 map[string]interface{}{"$first": "$p99"},
				"max_val":             map[string]interface{}{"$first": "$max_val"},
				"metric":              map[string]interface{}{"$first": "$metric"},
				"dt_event_time":       map[string]interface{}{"$first": "$dt_event_time"},
				"dt_event_time_stamp": map[string]interface{}{"$first": "$dt_event_time_stamp"},
				"the_date":            map[string]interface{}{"$first": "$the_date"},
				"local_time":          map[string]interface{}{"$first": "$local_time"},
			}},
		)
		err = m.DB.Table(tableName).Aggregation(ctx, pipeline, &metricResult)
		if err != nil {
			blog.Errorf("find namespace data fail, err:%v", err)
			return nil, err
		}
		result = append(result, metricResult...)
	}
	rsp := generateWorkloadRequestRsp(result)
	return rsp, nil
}

func (m *ModelWorkloadRequest) generateCond(req *datamanager.GetWorkloadRequestRecommendResultReq) map[string]interface{} {
	cond := make(map[string]interface{})
	cond[BCSClusterIDKey] = req.ClusterID
	if req.Namespace != "" {
		cond[NamespaceKey] = req.Namespace
	}
	if req.WorkloadType != "" {
		cond[WorkloadKindKey] = req.WorkloadType
	}
	if req.WorkloadName != "" {
		cond[WorkloadNameKey] = req.WorkloadName
	}
	timestamp := time.Now().AddDate(0, 0, -8).UnixMilli()
	cond[DTEventTimeStampKey] = map[string]interface{}{
		"$gte": timestamp,
	}
	return cond
}

func generateWorkloadRequestRsp(originResult []*types.BKBaseRequestRecommendResult) *datamanager.GetWorkloadRequestRecommendResultRsp {
	rsp := &datamanager.GetWorkloadRequestRecommendResultRsp{
		Data: make([]*datamanager.WorkloadRequestRecommendResult, 0),
	}
	resultMap := make(map[string]*datamanager.WorkloadRequestRecommendResult)
	for key := range originResult {
		value := originResult[key]
		blog.Infof("dtEventTimeStamp:%d", value.DTEventTimeStamp)
		blog.Infof("dtEventTime:%s", value.DTEventTime)
		container := &datamanager.WorkloadRequestRecommendContainer{
			Container: value.ContainerName,
			MaxVal:    value.MaxVal,
			P90:       value.P90,
			P99:       value.P99,
		}
		workloadKey := fmt.Sprintf("%s-%s-%s-%s", value.BCSClusterID, value.Namespace, value.WorkloadKind, value.WorkloadName)
		if _, ok := resultMap[workloadKey]; !ok {
			workload := &datamanager.WorkloadRequestRecommendResult{
				ClusterID:        value.BCSClusterID,
				Namespace:        value.Namespace,
				WorkloadType:     value.WorkloadKind,
				WorkloadName:     value.WorkloadName,
				Cpu:              make([]*datamanager.WorkloadRequestRecommendContainer, 0),
				Memory:           make([]*datamanager.WorkloadRequestRecommendContainer, 0),
				DtEventTimeStamp: value.DTEventTimeStamp,
			}
			resultMap[workloadKey] = workload
		}
		if value.Metric == "container_memory_working_set_bytes" {
			resultMap[workloadKey].Memory = append(resultMap[workloadKey].Memory, container)
		} else {
			resultMap[workloadKey].Cpu = append(resultMap[workloadKey].Cpu, container)
		}
	}
	for workload := range resultMap {
		rsp.Data = append(rsp.Data, resultMap[workload])
	}
	return rsp
}
