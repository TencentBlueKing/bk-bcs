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
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	modelNamespaceIndexes = []drivers.Index{
		{
			Key: bson.D{
				bson.E{Key: CreateTimeKey, Value: 1},
			},
			Name: CreateTimeKey + "_1",
		},
		{
			Name: common.NamespaceTableName + "_idx",
			Key: bson.D{
				bson.E{Key: ProjectIDKey, Value: 1},
				bson.E{Key: ClusterIDKey, Value: 1},
				bson.E{Key: NamespaceKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: BucketTimeKey, Value: 1},
			},
			Unique: true,
		},
		{
			Name: common.NamespaceTableName + "_list_idx",
			Key: bson.D{
				bson.E{Key: ClusterIDKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
		},
		{
			Name: common.NamespaceTableName + "_get_idx",
			Key: bson.D{
				bson.E{Key: ClusterIDKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: NamespaceKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
		},
		{
			Key: bson.D{
				bson.E{Key: ClusterIDKey, Value: 1},
			},
			Name: ClusterIDKey + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: DimensionKey, Value: 1},
			},
			Name: DimensionKey + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: ProjectIDKey, Value: 1},
			},
			Name: ProjectIDKey + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: NamespaceKey, Value: 1},
			},
			Name: NamespaceKey + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: BucketTimeKey, Value: 1},
			},
			Name: BucketTimeKey + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: MetricTimeKey, Value: 1},
			},
			Name: MetricTimeKey + "_1",
		},
	}
)

// ModelNamespace namespace model
type ModelNamespace struct {
	Public
}

// NewModelNamespace new namespace model
func NewModelNamespace(db drivers.DB) *ModelNamespace {
	return &ModelNamespace{
		Public: Public{
			TableName: common.DataTableNamePrefix + common.NamespaceTableName,
			Indexes:   modelNamespaceIndexes,
			DB:        db,
		}}
}

// InsertNamespaceInfo insert namespace data
func (m *ModelNamespace) InsertNamespaceInfo(ctx context.Context, metrics *common.NamespaceMetrics,
	opts *common.JobCommonOpts) error {
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return err
	}
	bucketTime, err := common.GetBucketTime(opts.CurrentTime, opts.Dimension)
	if err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		ProjectIDKey:  opts.ProjectID,
		ClusterIDKey:  opts.ClusterID,
		NamespaceKey:  opts.Namespace,
		DimensionKey:  opts.Dimension,
		BucketTimeKey: bucketTime,
	})
	retNamespace := &common.NamespaceData{}
	err = m.DB.Table(m.TableName).Find(cond).One(ctx, retNamespace)
	if err != nil {
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Infof(" namespace info not found, create a new bucket")
			newMetrics := make([]*common.NamespaceMetrics, 0)
			newMetrics = append(newMetrics, metrics)
			newNamespaceBucket := &common.NamespaceData{
				CreateTime:  primitive.NewDateTimeFromTime(time.Now()),
				UpdateTime:  primitive.NewDateTimeFromTime(time.Now()),
				BucketTime:  bucketTime,
				Dimension:   opts.Dimension,
				ProjectID:   opts.ProjectID,
				ClusterID:   opts.ClusterID,
				ClusterType: opts.ClusterType,
				Namespace:   opts.Namespace,
				Metrics:     newMetrics,
			}
			m.preAggregate(newNamespaceBucket, metrics)
			_, err = m.DB.Table(m.TableName).Insert(ctx, []interface{}{newNamespaceBucket})
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	m.preAggregate(retNamespace, metrics)
	retNamespace.UpdateTime = primitive.NewDateTimeFromTime(time.Now())
	retNamespace.Metrics = append(retNamespace.Metrics, metrics)
	return m.DB.Table(m.TableName).
		Update(ctx, cond, operator.M{"$set": retNamespace})
}

// GetNamespaceInfoList get namespace list
func (m *ModelNamespace) GetNamespaceInfoList(ctx context.Context,
	request *bcsdatamanager.GetNamespaceInfoListRequest) ([]*bcsdatamanager.Namespace, int64, error) {
	var total int64
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, total, err
	}
	dimension := request.Dimension
	if dimension == "" {
		dimension = common.DimensionMinute
	}
	cond := make([]*operator.Condition, 0)
	cond = append(cond,
		operator.NewLeafCondition(operator.Eq, operator.M{
			ClusterIDKey: request.ClusterID,
			DimensionKey: dimension,
		}), operator.NewLeafCondition(operator.Gte, operator.M{
			MetricTimeKey: primitive.NewDateTimeFromTime(getStartTime(dimension)),
		}),
	)
	conds := operator.NewBranchCondition(operator.And, cond...)
	tempNamespaceList := make([]map[string]string, 0)
	err = m.DB.Table(m.TableName).Find(conds).WithProjection(map[string]int{NamespaceKey: 1, "_id": 0}).
		All(ctx, &tempNamespaceList)
	if err != nil {
		blog.Errorf("get namespace list error")
		return nil, total, err
	}
	namespaceList := distinctSlice(NamespaceKey, &tempNamespaceList)
	if len(namespaceList) == 0 {
		return nil, total, nil
	}
	total = int64(len(namespaceList))

	page := int(request.Page)
	size := int(request.Size)
	if size == 0 {
		size = DefaultSize
	}
	endIndex := (page + 1) * size
	startIndex := page * size
	if startIndex >= len(namespaceList) {
		return nil, total, nil
	}
	if endIndex >= len(namespaceList) {
		endIndex = len(namespaceList)
	}
	chooseNamespace := namespaceList[startIndex:endIndex]
	response := make([]*bcsdatamanager.Namespace, 0)
	for _, namespace := range chooseNamespace {
		namespaceRequest := &bcsdatamanager.GetNamespaceInfoRequest{
			ClusterID: request.ClusterID,
			Namespace: namespace,
			Dimension: dimension,
		}
		namespaceInfo, err := m.GetNamespaceInfo(ctx, namespaceRequest)
		if err != nil {
			blog.Errorf("get namespace[%s] info err:%v", namespace, err)
		} else {
			response = append(response, namespaceInfo)
		}
	}
	return response, total, nil
}

// GetNamespaceInfo get namespace data
func (m *ModelNamespace) GetNamespaceInfo(ctx context.Context,
	request *bcsdatamanager.GetNamespaceInfoRequest) (*bcsdatamanager.Namespace, error) {
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	namespaceMetricsMap := make([]map[string]*common.NamespaceMetrics, 0)
	publicCond := operator.NewLeafCondition(operator.Eq, operator.M{
		ClusterIDKey:  request.ClusterID,
		ObjectTypeKey: common.NamespaceType,
		NamespaceKey:  request.Namespace,
	})
	namespacePublic := common.NamespacePublicMetrics{
		ResourceLimit: nil,
		SuggestCPU:    0,
		SuggestMemory: 0,
	}
	publicData := getPublicData(ctx, m.DB, publicCond)
	if publicData != nil && publicData.Metrics != nil {
		public, ok := publicData.Metrics.(common.NamespacePublicMetrics)
		if !ok {
			blog.Errorf("assert public data to namespace public failed")
		} else {
			namespacePublic = public
		}
	}

	dimension := request.Dimension
	if dimension == "" {
		dimension = common.DimensionMinute
	}
	pipeline := make([]map[string]interface{}, 0)
	pipeline = append(pipeline, map[string]interface{}{"$unwind": "$metrics"})
	pipeline = append(pipeline, map[string]interface{}{"$match": map[string]interface{}{
		ClusterIDKey: request.ClusterID,
		DimensionKey: dimension,
		NamespaceKey: request.Namespace,
		MetricTimeKey: map[string]interface{}{
			"$gte": primitive.NewDateTimeFromTime(getStartTime(dimension)),
		},
	}})
	pipeline = append(pipeline, map[string]interface{}{"$project": map[string]interface{}{
		"_id":     0,
		"metrics": 1,
	}})
	err = m.DB.Table(m.TableName).Aggregation(ctx, pipeline, &namespaceMetricsMap)
	if err != nil {
		blog.Errorf("find namespace data fail, err:%v", err)
		return nil, err
	}
	if len(namespaceMetricsMap) == 0 {
		return &bcsdatamanager.Namespace{}, nil
	}
	namespaceMetrics := make([]*common.NamespaceMetrics, 0)
	for _, metrics := range namespaceMetricsMap {
		namespaceMetrics = append(namespaceMetrics, metrics["metrics"])
	}
	startTime := namespaceMetrics[0].Time.Time().String()
	endTime := namespaceMetrics[len(namespaceMetrics)-1].Time.Time().String()
	return m.generateNamespaceResponse(namespacePublic, namespaceMetrics, request.ClusterID, request.Namespace,
		dimension, startTime, endTime), nil
}

// GetRawNamespaceInfo get raw namespace data
func (m *ModelNamespace) GetRawNamespaceInfo(ctx context.Context, opts *common.JobCommonOpts,
	bucket string) ([]*common.NamespaceData, error) {
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	cond := make([]*operator.Condition, 0)
	cond1 := operator.NewLeafCondition(operator.Eq, operator.M{
		ProjectIDKey: opts.ProjectID,
		DimensionKey: opts.Dimension,
		ClusterIDKey: opts.ClusterID,
	})
	cond = append(cond, cond1)
	if opts.Namespace != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			NamespaceKey: opts.Namespace,
		}))
	}
	if bucket != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			BucketTimeKey: bucket,
		}))
	}
	conds := operator.NewBranchCondition(operator.And, cond...)
	retNamespace := make([]*common.NamespaceData, 0)
	err = m.DB.Table(m.TableName).Find(conds).All(ctx, &retNamespace)
	if err != nil {
		return nil, err
	}
	return retNamespace, nil
}

func (m *ModelNamespace) generateNamespaceResponse(public common.NamespacePublicMetrics,
	metricSlice []*common.NamespaceMetrics, clusterId, namespace, dimension, startTime,
	endTime string) *bcsdatamanager.Namespace {
	response := &bcsdatamanager.Namespace{
		ClusterID:     clusterId,
		Dimension:     dimension,
		StartTime:     startTime,
		EndTime:       endTime,
		Namespace:     namespace,
		Metrics:       nil,
		SuggestCPU:    strconv.FormatFloat(public.SuggestCPU, 'f', 2, 64),
		SuggestMemory: strconv.FormatFloat(public.SuggestMemory, 'f', 2, 64),
		ResourceLimit: public.ResourceLimit,
	}
	responseMetrics := make([]*bcsdatamanager.NamespaceMetrics, 0)
	for _, metric := range metricSlice {
		responseMetric := &bcsdatamanager.NamespaceMetrics{
			Time:               metric.Time.Time().String(),
			CPURequest:         strconv.FormatFloat(metric.CPURequest, 'f', 2, 64),
			MemoryRequest:      strconv.FormatInt(metric.MemoryRequest, 10),
			CPUUsageAmount:     strconv.FormatFloat(metric.CPUUsageAmount, 'f', 2, 64),
			MemoryUsageAmount:  strconv.FormatInt(metric.MemoryUsageAmount, 10),
			CPUUsage:           strconv.FormatFloat(metric.CPUUsage, 'f', 4, 64),
			MemoryUsage:        strconv.FormatFloat(metric.MemoryUsage, 'f', 4, 64),
			MaxCPUUsageTime:    metric.MaxCPUUsageTime,
			MinCPUUsageTime:    metric.MinCPUUsageTime,
			MaxMemoryUsageTime: metric.MaxMemoryUsageTime,
			MinMemoryUsageTime: metric.MinMemoryUsageTime,
			WorkloadCount:      strconv.FormatInt(metric.WorkloadCount, 10),
			InstanceCount:      strconv.FormatInt(metric.InstanceCount, 10),
			MinInstanceTime:    metric.MinInstanceTime,
			MaxInstanceTime:    metric.MaxInstanceTime,
			MinWorkloadUsage:   metric.MinWorkloadUsage,
			MaxWorkloadUsage:   metric.MaxWorkloadUsage,
		}
		responseMetrics = append(responseMetrics, responseMetric)
	}
	response.Metrics = responseMetrics
	return response
}

func (m *ModelNamespace) preAggregate(data *common.NamespaceData, newMetric *common.NamespaceMetrics) {
	if data.MaxInstanceTime == nil {
		data.MaxInstanceTime = newMetric.MaxInstanceTime
	} else if newMetric.MaxInstanceTime.Value > data.MaxInstanceTime.Value {
		data.MaxInstanceTime = newMetric.MaxInstanceTime
	}

	if data.MinInstanceTime == nil {
		data.MinInstanceTime = newMetric.MinInstanceTime
	} else if newMetric.MinInstanceTime.Value < data.MinInstanceTime.Value {
		data.MinInstanceTime = newMetric.MinInstanceTime
	}

	if data.MaxCPUUsageTime == nil {
		data.MaxCPUUsageTime = newMetric.MaxCPUUsageTime
	} else if newMetric.MaxCPUUsageTime.Value > data.MaxCPUUsageTime.Value {
		data.MaxCPUUsageTime = newMetric.MaxCPUUsageTime
	}

	if data.MinCPUUsageTime == nil {
		data.MinCPUUsageTime = newMetric.MinCPUUsageTime
	} else if newMetric.MinCPUUsageTime.Value < data.MinCPUUsageTime.Value {
		data.MinCPUUsageTime = newMetric.MinCPUUsageTime
	}

	if data.MaxMemoryUsageTime == nil {
		data.MaxMemoryUsageTime = newMetric.MaxMemoryUsageTime
	} else if newMetric.MaxMemoryUsageTime.Value > data.MaxMemoryUsageTime.Value {
		data.MaxMemoryUsageTime = newMetric.MaxMemoryUsageTime
	}
	if data.MinMemoryUsageTime == nil {
		data.MinMemoryUsageTime = newMetric.MinMemoryUsageTime
	} else if newMetric.MinMemoryUsageTime.Value < data.MinMemoryUsageTime.Value {
		data.MinMemoryUsageTime = newMetric.MinMemoryUsageTime
	}

	if data.MaxWorkloadUsage == nil {
		data.MaxWorkloadUsage = newMetric.MaxWorkloadUsage
	} else if newMetric.MaxWorkloadUsage.Value > data.MaxWorkloadUsage.Value {
		data.MaxWorkloadUsage = newMetric.MaxWorkloadUsage
	}
	if data.MinWorkloadUsage == nil {
		data.MinWorkloadUsage = newMetric.MinWorkloadUsage
	} else if newMetric.MinWorkloadUsage.Value < data.MinWorkloadUsage.Value {
		data.MinWorkloadUsage = newMetric.MinWorkloadUsage
	}
}
