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
	"errors"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/utils"
	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
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
			Name: types.NamespaceTableName + "_idx",
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
			Name: types.NamespaceTableName + "_list_idx",
			Key: bson.D{
				bson.E{Key: ClusterIDKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
		},
		{
			Name: types.NamespaceTableName + "_get_idx",
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
			TableName: types.DataTableNamePrefix + types.NamespaceTableName,
			Indexes:   modelNamespaceIndexes,
			DB:        db,
		}}
}

// InsertNamespaceInfo insert namespace data
func (m *ModelNamespace) InsertNamespaceInfo(ctx context.Context, metrics *types.NamespaceMetrics,
	opts *types.JobCommonOpts) error {
	// Ensure that the table exists in the database.
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return err
	}
	// Get the bucket time for the current time and dimension.
	bucketTime, err := utils.GetBucketTime(opts.CurrentTime, opts.Dimension)
	if err != nil {
		return err
	}
	// Create a condition to find the existing namespace data in the database.
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		ProjectIDKey:  opts.ProjectID,
		ClusterIDKey:  opts.ClusterID,
		NamespaceKey:  opts.Namespace,
		DimensionKey:  opts.Dimension,
		BucketTimeKey: bucketTime,
	})

	// Find the existing namespace data in the database.
	retNamespace := &types.NamespaceData{}
	err = m.DB.Table(m.TableName).Find(cond).One(ctx, retNamespace)
	if err != nil {
		// If the namespace data is not found, create a new bucket and insert the metrics.
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Infof(" namespace info not found, create a new bucket")
			newMetrics := make([]*types.NamespaceMetrics, 0)
			newMetrics = append(newMetrics, metrics)
			newNamespaceBucket := &types.NamespaceData{
				CreateTime:  primitive.NewDateTimeFromTime(time.Now()),
				UpdateTime:  primitive.NewDateTimeFromTime(time.Now()),
				BucketTime:  bucketTime,
				Dimension:   opts.Dimension,
				ProjectID:   opts.ProjectID,
				BusinessID:  opts.BusinessID,
				ClusterID:   opts.ClusterID,
				ClusterType: opts.ClusterType,
				Namespace:   opts.Namespace,
				Metrics:     newMetrics,
				Label:       opts.Label,
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

	// If the namespace data is found, update it with the new metrics.
	m.preAggregate(retNamespace, metrics)
	if retNamespace.BusinessID == "" {
		retNamespace.BusinessID = opts.BusinessID
	}
	retNamespace.Label = opts.Label
	retNamespace.ProjectCode = opts.ProjectCode
	retNamespace.UpdateTime = primitive.NewDateTimeFromTime(time.Now())
	retNamespace.Metrics = append(retNamespace.Metrics, metrics)
	return m.DB.Table(m.TableName).
		Update(ctx, cond, operator.M{"$set": retNamespace})
}

// GetNamespaceInfoList get namespace list by cluster id
func (m *ModelNamespace) GetNamespaceInfoList(ctx context.Context,
	request *bcsdatamanager.GetNamespaceInfoListRequest) ([]*bcsdatamanager.Namespace, int64, error) {
	var total int64
	// Ensure that the table exists in the database.
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, total, err
	}
	dimension := request.Dimension
	if dimension == "" {
		dimension = types.DimensionMinute
	}
	cond := make([]*operator.Condition, 0)
	cond = append(cond,
		operator.NewLeafCondition(operator.Eq, operator.M{
			ClusterIDKey: request.ClusterID,
			DimensionKey: dimension,
		}),
	)
	startTime := getStartTime(dimension)
	if request.GetStartTime() != 0 {
		startTime = time.Unix(request.GetStartTime(), 0)
	}
	cond = append(cond, operator.NewLeafCondition(operator.Gte, operator.M{
		MetricTimeKey: primitive.NewDateTimeFromTime(startTime),
	}))
	if request.GetEndTime() != 0 {
		cond = append(cond, operator.NewLeafCondition(operator.Lte, operator.M{
			MetricTimeKey: primitive.NewDateTimeFromTime(time.Unix(request.GetEndTime(), 0)),
		}))
	}
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
			StartTime: request.GetStartTime(),
			EndTime:   request.GetEndTime(),
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

// GetNamespaceInfo get namespace data with default time range
// nolint funlen
func (m *ModelNamespace) GetNamespaceInfo(ctx context.Context,
	request *bcsdatamanager.GetNamespaceInfoRequest) (*bcsdatamanager.Namespace, error) {
	// Ensure that the table exists in the database.
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	namespaceMetricsMap := make([]*types.NamespaceData, 0)
	publicCond := operator.NewLeafCondition(operator.Eq, operator.M{
		ClusterIDKey:  request.ClusterID,
		ObjectTypeKey: types.NamespaceType,
		NamespaceKey:  request.Namespace,
	})
	namespacePublic := types.NamespacePublicMetrics{
		ResourceLimit: nil,
		SuggestCPU:    0,
		SuggestMemory: 0,
	}
	publicData := getPublicData(ctx, m.DB, publicCond)
	if publicData != nil && publicData.Metrics != nil {
		public, ok := publicData.Metrics.(types.NamespacePublicMetrics)
		if !ok {
			blog.Errorf("assert public data to namespace public failed")
		} else {
			namespacePublic = public
		}
	}

	dimension := request.Dimension
	if dimension == "" {
		dimension = types.DimensionMinute
	}
	metricStartTime := getStartTime(dimension)
	if request.GetStartTime() != 0 {
		metricStartTime = time.Unix(request.GetStartTime(), 0)
	}
	metricEndTime := time.Now()
	if request.GetEndTime() != 0 {
		metricEndTime = time.Unix(request.GetEndTime(), 0)
	}
	pipeline := make([]map[string]interface{}, 0)
	pipeline = append(pipeline, map[string]interface{}{"$match": map[string]interface{}{
		ClusterIDKey: request.ClusterID,
		DimensionKey: dimension,
		NamespaceKey: request.Namespace,
		MetricTimeKey: map[string]interface{}{
			"$gte": primitive.NewDateTimeFromTime(metricStartTime),
			"$lte": primitive.NewDateTimeFromTime(metricEndTime),
		},
	}}, map[string]interface{}{"$unwind": "$metrics"},
		map[string]interface{}{"$match": map[string]interface{}{
			MetricTimeKey: map[string]interface{}{
				"$gte": primitive.NewDateTimeFromTime(metricStartTime),
				"$lte": primitive.NewDateTimeFromTime(metricEndTime),
			},
		}},
		map[string]interface{}{"$project": map[string]interface{}{
			"_id":          0,
			"metrics":      1,
			"business_id":  1,
			"project_id":   1,
			"project_code": 1,
			"namespace":    1,
			"cluster_id":   1,
			"label":        1,
		}}, map[string]interface{}{"$group": map[string]interface{}{
			"_id":          nil,
			"cluster_id":   map[string]interface{}{"$first": "$cluster_id"},
			"namespace":    map[string]interface{}{"$first": "$namespace"},
			"project_id":   map[string]interface{}{"$first": "$project_id"},
			"business_id":  map[string]interface{}{"$max": "$business_id"},
			"metrics":      map[string]interface{}{"$push": "$metrics"},
			"label":        map[string]interface{}{"$max": "$label"},
			"project_code": map[string]interface{}{"$max": "$project_code"},
		}},
	)
	err = m.DB.Table(m.TableName).Aggregation(ctx, pipeline, &namespaceMetricsMap)
	if err != nil {
		blog.Errorf("find namespace data fail, err:%v", err)
		return nil, err
	}
	if len(namespaceMetricsMap) == 0 {
		return &bcsdatamanager.Namespace{}, nil
	}
	namespaceMetrics := make([]*types.NamespaceMetrics, 0)
	for _, metrics := range namespaceMetricsMap {
		namespaceMetrics = append(namespaceMetrics, metrics.Metrics...)
	}
	startTime := namespaceMetrics[0].Time.Time().String()
	endTime := namespaceMetrics[len(namespaceMetrics)-1].Time.Time().String()
	return m.generateNamespaceResponse(namespacePublic, namespaceMetrics, namespaceMetricsMap[0],
		startTime, endTime), nil
}

// GetRawNamespaceInfo is a function that retrieves raw namespace data without a time range.
func (m *ModelNamespace) GetRawNamespaceInfo(ctx context.Context, opts *types.JobCommonOpts,
	bucket string) ([]*types.NamespaceData, error) {
	// Ensure that the table exists in the database.
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	// Create a slice of conditions to filter the database query results.
	cond := make([]*operator.Condition, 0)
	// Add a condition for the project ID, dimension, and cluster ID.
	cond1 := operator.NewLeafCondition(operator.Eq, operator.M{
		ProjectIDKey: opts.ProjectID,
		DimensionKey: opts.Dimension,
		ClusterIDKey: opts.ClusterID,
	})
	cond = append(cond, cond1)
	// If a namespace is specified, add a condition for the namespace.
	if opts.Namespace != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			NamespaceKey: opts.Namespace,
		}))
	}
	// If a bucket is specified, add a condition for the bucket time.
	if bucket != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			BucketTimeKey: bucket,
		}))
	}
	// Combine the conditions into a single branch condition with an "and" operator.
	conds := operator.NewBranchCondition(operator.And, cond...)
	// Create an empty slice of NamespaceData to store the results of the database query.
	retNamespace := make([]*types.NamespaceData, 0)
	// Query the database with the conditions and store the results in retNamespace.
	err = m.DB.Table(m.TableName).Find(conds).All(ctx, &retNamespace)
	if err != nil {
		return nil, err
	}
	// Return the results.
	return retNamespace, nil
}

// generateNamespaceResponse generate response, transfer storage namespace to proto namespace
func (m *ModelNamespace) generateNamespaceResponse(public types.NamespacePublicMetrics,
	metricSlice []*types.NamespaceMetrics,
	data *types.NamespaceData, start, end string) *bcsdatamanager.Namespace {
	response := &bcsdatamanager.Namespace{
		ProjectID:     data.ProjectID,
		ProjectCode:   data.ProjectCode,
		BusinessID:    data.BusinessID,
		ClusterID:     data.ClusterID,
		StartTime:     start,
		EndTime:       end,
		Namespace:     data.Namespace,
		Metrics:       nil,
		SuggestCPU:    strconv.FormatFloat(public.SuggestCPU, 'f', 2, 64),
		SuggestMemory: strconv.FormatFloat(public.SuggestMemory, 'f', 2, 64),
		ResourceLimit: public.ResourceLimit,
		Label:         data.Label,
	}
	responseMetrics := make([]*bcsdatamanager.NamespaceMetrics, 0)
	for _, metric := range metricSlice {
		responseMetric := &bcsdatamanager.NamespaceMetrics{
			Time:              metric.Time.Time().String(),
			CPURequest:        strconv.FormatFloat(metric.CPURequest, 'f', 2, 64),
			CPULimit:          strconv.FormatFloat(metric.CPULimit, 'f', 2, 64),
			MemoryRequest:     strconv.FormatInt(metric.MemoryRequest, 10),
			MemoryLimit:       strconv.FormatInt(metric.MemoryLimit, 10),
			CPUUsageAmount:    strconv.FormatFloat(metric.CPUUsageAmount, 'f', 2, 64),
			MemoryUsageAmount: strconv.FormatInt(metric.MemoryUsageAmount, 10),
			CPUUsage:          strconv.FormatFloat(metric.CPUUsage, 'f', 4, 64),
			MemoryUsage:       strconv.FormatFloat(metric.MemoryUsage, 'f', 4, 64),
			MaxCPU:            metric.MaxCPUUsageTime,
			MinCPU:            metric.MinCPUUsageTime,
			MaxMemory:         metric.MaxMemoryUsageTime,
			MinMemory:         metric.MinMemoryUsageTime,
			WorkloadCount:     strconv.FormatInt(metric.WorkloadCount, 10),
			InstanceCount:     strconv.FormatInt(metric.InstanceCount, 10),
			MinInstance:       metric.MinInstanceTime,
			MaxInstance:       metric.MaxInstanceTime,
			MinWorkloadUsage:  metric.MinWorkloadUsage,
			MaxWorkloadUsage:  metric.MaxWorkloadUsage,
		}
		responseMetrics = append(responseMetrics, responseMetric)
	}
	response.Metrics = responseMetrics
	return response
}

// preAggregate is a function that performs pre-aggregation to get the minimum and maximum values of various metrics.
func (m *ModelNamespace) preAggregate(data *types.NamespaceData, newMetric *types.NamespaceMetrics) {
	// If data.MaxInstanceTime is nil, update it to newMetric.MaxInstanceTime.
	// Otherwise, if newMetric.MaxInstanceTime is greater than data.MaxInstanceTime,
	// update data.MaxInstanceTime to newMetric. MaxInstanceTime.
	if data.MaxInstanceTime == nil {
		data.MaxInstanceTime = newMetric.MaxInstanceTime
	} else if newMetric.MaxInstanceTime.Value > data.MaxInstanceTime.Value {
		data.MaxInstanceTime = newMetric.MaxInstanceTime
	}

	// Repeat the above process for MinInstanceTime, MaxCPUUsageTime, MinCPUUsageTime, MaxMemoryUsageTime,
	// and MinMemoryUsageTime.
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

	// Repeat the above process for MaxWorkloadUsage and MinWorkloadUsage.
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
