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
	"errors"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/utils"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	modelClusterIndexes = []drivers.Index{
		{
			Key: bson.D{
				bson.E{Key: CreateTimeKey, Value: 1},
			},
			Name: CreateTimeKey + "_1",
		},
		{
			Name: types.ClusterTableName + "_idx",
			Key: bson.D{
				bson.E{Key: ProjectIDKey, Value: 1},
				bson.E{Key: ClusterIDKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: BucketTimeKey, Value: 1},
			},
			Unique: true,
		},
		{
			Name: types.ClusterTableName + "_list_idx",
			Key: bson.D{
				bson.E{Key: ProjectIDKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
		},
		{
			Name: types.ClusterTableName + "_list_idx2",
			Key: bson.D{
				bson.E{Key: BusinessIDKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
			Background: true,
		},
		{
			Name: types.ClusterTableName + "_list_idx3",
			Key: bson.D{
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
			Background: true,
		},
		{
			Name: types.ClusterTableName + "_get_idx",
			Key: bson.D{
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: ClusterIDKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
			Background: true,
		},
		{
			Key: bson.D{
				bson.E{Key: BucketTimeKey, Value: 1},
			},
			Name: BucketTimeKey + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: ClusterIDKey, Value: 1},
			},
			Name: ClusterIDKey + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: ProjectIDKey, Value: 1},
			},
			Name: ProjectIDKey + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: DimensionKey, Value: 1},
			},
			Name: DimensionKey + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: MetricTimeKey, Value: 1},
			},
			Name: MetricTimeKey + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: BusinessIDKey, Value: 1},
			},
			Name: BusinessIDKey + "_1",
		},
	}
)

// ModelCluster cluster model
type ModelCluster struct {
	Public
}

// NewModelCluster new cluster model
func NewModelCluster(db drivers.DB) *ModelCluster {
	return &ModelCluster{Public: Public{
		TableName: types.DataTableNamePrefix + types.ClusterTableName,
		Indexes:   modelClusterIndexes,
		DB:        db,
	}}
}

// InsertClusterInfo insert cluster data
func (m *ModelCluster) InsertClusterInfo(ctx context.Context, metrics *types.ClusterMetrics,
	opts *types.JobCommonOpts) error {
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return err
	}
	bucketTime, err := utils.GetBucketTime(opts.CurrentTime, opts.Dimension)
	if err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		ProjectIDKey:  opts.ProjectID,
		ClusterIDKey:  opts.ClusterID,
		DimensionKey:  opts.Dimension,
		BucketTimeKey: bucketTime,
	})
	retCluster := &types.ClusterData{}
	err = m.DB.Table(m.TableName).Find(cond).One(ctx, retCluster)
	if err != nil {
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Infof("cluster info not found, create a new bucket")
			newMetrics := make([]*types.ClusterMetrics, 0)
			newMetrics = append(newMetrics, metrics)
			newClusterBucket := &types.ClusterData{
				CreateTime:   primitive.NewDateTimeFromTime(time.Now()),
				UpdateTime:   primitive.NewDateTimeFromTime(time.Now()),
				BucketTime:   bucketTime,
				Dimension:    opts.Dimension,
				ProjectID:    opts.ProjectID,
				ProjectCode:  opts.ProjectCode,
				BusinessID:   opts.BusinessID,
				ClusterID:    opts.ClusterID,
				ClusterType:  opts.ClusterType,
				Metrics:      newMetrics,
				TotalCACount: metrics.CACount,
				Label:        opts.Label,
			}
			m.preAggregate(newClusterBucket, metrics)
			_, err = m.DB.Table(m.TableName).Insert(ctx, []interface{}{newClusterBucket})
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	m.preAggregate(retCluster, metrics)
	retCluster.UpdateTime = primitive.NewDateTimeFromTime(time.Now())
	if retCluster.BusinessID == "" {
		retCluster.BusinessID = opts.BusinessID
	}
	retCluster.Label = opts.Label
	retCluster.ProjectCode = opts.ProjectCode
	retCluster.Metrics = append(retCluster.Metrics, metrics)
	retCluster.TotalCACount += metrics.CACount
	return m.DB.Table(m.TableName).
		Update(ctx, cond, operator.M{"$set": retCluster})
}

// GetClusterInfoList get cluster list by project id
func (m *ModelCluster) GetClusterInfoList(ctx context.Context,
	request *bcsdatamanager.GetClusterListRequest) ([]*bcsdatamanager.Cluster, int64, error) {
	err := ensureTable(ctx, &m.Public)
	var total int64
	if err != nil {
		return nil, total, err
	}
	dimension := request.Dimension
	if dimension == "" {
		dimension = types.DimensionMinute
	}
	cond := make([]*operator.Condition, 0)
	if request.GetProject() != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			ProjectIDKey: request.Project,
		}))
	} else if request.GetBusiness() != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			BusinessIDKey: request.GetBusiness(),
		}))
	}
	cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
		DimensionKey: dimension,
	}))
	cond = append(cond, operator.NewLeafCondition(operator.Gte, operator.M{
		MetricTimeKey: primitive.NewDateTimeFromTime(getStartTime(dimension)),
	}))
	conds := operator.NewBranchCondition(operator.And, cond...)
	tempClusterList := make([]map[string]string, 0)
	err = m.DB.Table(m.TableName).Find(conds).WithProjection(map[string]int{ClusterIDKey: 1, "_id": 0}).
		WithSort(map[string]interface{}{ClusterIDKey: 1}).All(ctx, &tempClusterList)
	if err != nil {
		blog.Errorf("get cluster id list error")
		return nil, total, err
	}

	clusterList := distinctSlice("cluster_id", &tempClusterList)
	if len(clusterList) == 0 {
		return nil, total, nil
	}
	total = int64(len(clusterList))
	page := int(request.Page)
	size := int(request.Size)
	if size == 0 {
		size = DefaultSize
	}
	endIndex := (page + 1) * size
	startIndex := page * size
	if startIndex >= len(clusterList) {
		return nil, total, nil
	}
	if endIndex >= len(clusterList) {
		endIndex = len(clusterList)
	}
	chooseCluster := clusterList[startIndex:endIndex]
	response := make([]*bcsdatamanager.Cluster, 0)
	for _, cluster := range chooseCluster {
		clusterRequest := &bcsdatamanager.GetClusterInfoRequest{
			ClusterID: cluster,
			Dimension: dimension,
		}
		clusterInfo, err := m.GetClusterInfo(ctx, clusterRequest)
		if err != nil {
			blog.Errorf("get cluster[%s] info err:%v", cluster, err)
		} else {
			response = append(response, clusterInfo)
		}
	}
	return response, total, nil
}

// GetClusterInfo get cluster data for api, return metrics with default time range
func (m *ModelCluster) GetClusterInfo(ctx context.Context,
	request *bcsdatamanager.GetClusterInfoRequest) (*bcsdatamanager.Cluster, error) {
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	dimension := request.Dimension
	if dimension == "" {
		dimension = types.DimensionMinute
	}
	clusterMetricsMap := make([]*types.ClusterData, 0)
	metricStartTime := getStartTime(dimension)
	pipeline := make([]map[string]interface{}, 0)
	pipeline = append(pipeline,
		map[string]interface{}{"$match": map[string]interface{}{
			ClusterIDKey: request.ClusterID,
			DimensionKey: dimension,
			MetricTimeKey: map[string]interface{}{
				"$gte": primitive.NewDateTimeFromTime(metricStartTime),
			},
		}},
		map[string]interface{}{"$unwind": "$metrics"},
		map[string]interface{}{"$match": map[string]interface{}{
			MetricTimeKey: map[string]interface{}{
				"$gte": primitive.NewDateTimeFromTime(metricStartTime),
			},
		}},
		map[string]interface{}{"$project": map[string]interface{}{
			"_id":          0,
			"metrics":      1,
			"cluster_id":   1,
			"business_id":  1,
			"project_id":   1,
			"project_code": 1,
			"label":        1,
		}}, map[string]interface{}{"$group": map[string]interface{}{
			"_id":          nil,
			"cluster_id":   map[string]interface{}{"$first": "$cluster_id"},
			"project_id":   map[string]interface{}{"$first": "$project_id"},
			"business_id":  map[string]interface{}{"$max": "$business_id"},
			"metrics":      map[string]interface{}{"$push": "$metrics"},
			"label":        map[string]interface{}{"$max": "$label"},
			"project_code": map[string]interface{}{"$max": "$project_code"},
		}})
	err = m.DB.Table(m.TableName).Aggregation(ctx, pipeline, &clusterMetricsMap)
	if err != nil {
		blog.Errorf("find cluster data fail, err:%v", err)
		return nil, err
	}
	if len(clusterMetricsMap) == 0 {
		return &bcsdatamanager.Cluster{}, nil
	}
	clusterMetrics := make([]*types.ClusterMetrics, 0)
	for _, metrics := range clusterMetricsMap {
		clusterMetrics = append(clusterMetrics, metrics.Metrics...)
	}
	startTime := clusterMetrics[0].Time.Time().String()
	endTime := clusterMetrics[len(clusterMetrics)-1].Time.Time().String()
	return m.generateClusterResponse(clusterMetrics, clusterMetricsMap[0], dimension, startTime, endTime), nil
}

// GetRawClusterInfo get raw cluster data without time range
func (m *ModelCluster) GetRawClusterInfo(ctx context.Context, opts *types.JobCommonOpts,
	bucket string) ([]*types.ClusterData, error) {
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	cond := make([]*operator.Condition, 0)
	cond1 := operator.NewLeafCondition(operator.Eq, operator.M{
		ProjectIDKey: opts.ProjectID,
		DimensionKey: opts.Dimension,
	})
	cond = append(cond, cond1)
	if opts.ClusterID != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			ClusterIDKey: opts.ClusterID,
		}))
	}
	if bucket != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			BucketTimeKey: bucket,
		}))
	}
	conds := operator.NewBranchCondition(operator.And, cond...)
	retCluster := make([]*types.ClusterData, 0)
	err = m.DB.Table(m.TableName).Find(conds).All(ctx, &retCluster)
	if err != nil {
		return nil, err
	}
	return retCluster, nil
}

func (m *ModelCluster) generateClusterResponse(metricSlice []*types.ClusterMetrics, data *types.ClusterData,
	dimension, startTime, endTime string) *bcsdatamanager.Cluster {
	response := &bcsdatamanager.Cluster{
		ProjectID:   data.ProjectID,
		ProjectCode: data.ProjectCode,
		BusinessID:  data.BusinessID,
		ClusterID:   data.ClusterID,
		Dimension:   dimension,
		StartTime:   startTime,
		EndTime:     endTime,
		Metrics:     nil,
		Label:       data.Label,
	}
	responseMetrics := make([]*bcsdatamanager.ClusterMetrics, 0)
	for _, metric := range metricSlice {
		responseMetric := &bcsdatamanager.ClusterMetrics{
			Time:               metric.Time.Time().String(),
			NodeCount:          strconv.FormatInt(metric.NodeCount, 10),
			AvailableNodeCount: strconv.FormatInt(metric.AvailableNodeCount, 10),
			MinUsageNode:       metric.MinUsageNode,
			TotalCPU:           strconv.FormatFloat(metric.TotalCPU, 'f', 2, 64),
			TotalMemory:        strconv.FormatInt(metric.TotalMemory, 10),
			TotalLoadCPU:       strconv.FormatFloat(metric.TotalLoadCPU, 'f', 2, 64),
			TotalLoadMemory:    strconv.FormatInt(metric.TotalLoadMemory, 10),
			AvgLoadCPU:         strconv.FormatFloat(metric.AvgLoadCPU, 'f', 2, 64),
			AvgLoadMemory:      strconv.FormatInt(metric.AvgLoadMemory, 10),
			CPUUsage:           strconv.FormatFloat(metric.CPUUsage, 'f', 4, 64),
			MemoryUsage:        strconv.FormatFloat(metric.MemoryUsage, 'f', 4, 64),
			WorkloadCount:      strconv.FormatInt(metric.WorkloadCount, 10),
			InstanceCount:      strconv.FormatInt(metric.InstanceCount, 10),
			CpuRequest:         strconv.FormatFloat(metric.CpuRequest, 'f', 2, 64),
			MemoryRequest:      strconv.FormatInt(metric.MemoryRequest, 10),
			MinNode:            metric.MinNode,
			MaxNode:            metric.MaxNode,
			MaxInstanceTime:    metric.MaxInstance,
			MinInstance:        metric.MinInstance,
			NodeQuantile:       metric.NodeQuantile,
			CACount:            strconv.FormatInt(metric.CACount, 10),
		}
		responseMetrics = append(responseMetrics, responseMetric)
	}
	response.Metrics = responseMetrics
	return response
}

func (m *ModelCluster) preAggregate(data *types.ClusterData, newMetric *types.ClusterMetrics) {
	if data.MaxInstance == nil {
		data.MaxInstance = newMetric.MaxInstance
	} else if newMetric.MaxInstance.Value > data.MaxInstance.Value {
		data.MaxInstance = newMetric.MaxInstance
	}

	if data.MinInstance == nil {
		data.MinInstance = newMetric.MinInstance
	} else if newMetric.MinInstance.Value < data.MinInstance.Value {
		data.MinInstance = newMetric.MinInstance
	}

	if data.MaxNode == nil {
		data.MaxNode = newMetric.MaxNode
	} else if newMetric.MaxNode.Value > data.MaxNode.Value {
		data.MaxNode = newMetric.MaxNode
	}

	if data.MinNode == nil {
		data.MinNode = newMetric.MinNode
	} else if newMetric.MinNode.Value < data.MinNode.Value {
		data.MinNode = newMetric.MinNode
	}
}
