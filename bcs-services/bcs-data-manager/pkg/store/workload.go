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
	modelWorkloadIndexes = []drivers.Index{
		{
			Key: bson.D{
				bson.E{Key: CreateTimeKey, Value: 1},
			},
			Name: CreateTimeKey + "_1",
		},
		{
			Name: common.WorkloadTableName + "_idx",
			Key: bson.D{
				bson.E{Key: ProjectIDKey, Value: 1},
				bson.E{Key: ClusterIDKey, Value: 1},
				bson.E{Key: NamespaceKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: WorkloadTypeKey, Value: 1},
				bson.E{Key: WorkloadNameKey, Value: 1},
				bson.E{Key: BucketTimeKey, Value: 1},
			},
			Unique: true,
		},
		{
			Name: common.WorkloadTableName + "_list_idx",
			Key: bson.D{
				bson.E{Key: ClusterIDKey, Value: 1},
				bson.E{Key: NamespaceKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: WorkloadTypeKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
		},
		{
			Name: common.WorkloadTableName + "_get_idx",
			Key: bson.D{
				bson.E{Key: ClusterIDKey, Value: 1},
				bson.E{Key: NamespaceKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: WorkloadTypeKey, Value: 1},
				bson.E{Key: WorkloadNameKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
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
				bson.E{Key: NamespaceKey, Value: 1},
			},
			Name: NamespaceKey + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: WorkloadNameKey, Value: 1},
			},
			Name: WorkloadNameKey + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: WorkloadTypeKey, Value: 1},
			},
			Name: WorkloadTypeKey + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: MetricTimeKey, Value: 1},
			},
			Name: MetricTimeKey + "_1",
		},
	}
)

// ModelWorkload workload model
type ModelWorkload struct {
	Public
}

// NewModelWorkload new workload model
func NewModelWorkload(db drivers.DB) *ModelWorkload {
	return &ModelWorkload{Public: Public{
		TableName: common.DataTableNamePrefix + common.WorkloadTableName,
		Indexes:   modelWorkloadIndexes,
		DB:        db,
	}}
}

// InsertWorkloadInfo insert workload info
func (m *ModelWorkload) InsertWorkloadInfo(ctx context.Context, metrics *common.WorkloadMetrics,
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
		ProjectIDKey:    opts.ProjectID,
		ClusterIDKey:    opts.ClusterID,
		NamespaceKey:    opts.Namespace,
		DimensionKey:    opts.Dimension,
		WorkloadTypeKey: opts.WorkloadType,
		WorkloadNameKey: opts.Name,
		BucketTimeKey:   bucketTime,
	})
	retWorkload := &common.WorkloadData{}
	err = m.DB.Table(m.TableName).Find(cond).One(ctx, retWorkload)
	if err != nil {
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Infof(" workload info not found, create a new bucket")
			newMetrics := make([]*common.WorkloadMetrics, 0)
			newMetrics = append(newMetrics, metrics)
			newWorkloadBucket := &common.WorkloadData{
				CreateTime:   primitive.NewDateTimeFromTime(time.Now()),
				UpdateTime:   primitive.NewDateTimeFromTime(time.Now()),
				BucketTime:   bucketTime,
				Dimension:    opts.Dimension,
				ProjectID:    opts.ProjectID,
				ClusterID:    opts.ClusterID,
				ClusterType:  opts.ClusterType,
				Namespace:    opts.Namespace,
				WorkloadType: opts.WorkloadType,
				Name:         opts.Name,
				Metrics:      newMetrics,
			}
			m.preAggregateMax(newWorkloadBucket, metrics)
			m.preAggregateMin(newWorkloadBucket, metrics)
			_, err = m.DB.Table(m.TableName).Insert(ctx, []interface{}{newWorkloadBucket})
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	m.preAggregateMax(retWorkload, metrics)
	m.preAggregateMin(retWorkload, metrics)
	retWorkload.UpdateTime = primitive.NewDateTimeFromTime(time.Now())
	retWorkload.Metrics = append(retWorkload.Metrics, metrics)
	return m.DB.Table(m.TableName).
		Update(ctx, cond, operator.M{"$set": retWorkload})
}

// GetWorkloadInfoList get workload list data by cluster id, namespace and workload type
func (m *ModelWorkload) GetWorkloadInfoList(ctx context.Context,
	request *bcsdatamanager.GetWorkloadInfoListRequest) ([]*bcsdatamanager.Workload, int64, error) {
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
			ClusterIDKey:    request.ClusterID,
			DimensionKey:    dimension,
			NamespaceKey:    request.Namespace,
			WorkloadTypeKey: request.WorkloadType,
		}), operator.NewLeafCondition(operator.Gte, operator.M{
			MetricTimeKey: primitive.NewDateTimeFromTime(getStartTime(dimension)),
		}))
	conds := operator.NewBranchCondition(operator.And, cond...)
	tempWorkloadList := make([]map[string]string, 0)
	err = m.DB.Table(m.TableName).Find(conds).WithProjection(map[string]int{WorkloadNameKey: 1}).
		All(ctx, &tempWorkloadList)
	if err != nil {
		blog.Errorf("get cluster id list error")
		return nil, total, err
	}
	workloadList := distinctSlice(WorkloadNameKey, &tempWorkloadList)
	if len(workloadList) == 0 {
		return nil, total, nil
	}
	total = int64(len(workloadList))

	page := int(request.Page)
	size := int(request.Size)
	if size == 0 {
		size = DefaultSize
	}
	endIndex := (page + 1) * size
	startIndex := page * size
	if startIndex >= len(workloadList) {
		return nil, total, nil
	}
	if endIndex >= len(workloadList) {
		endIndex = len(workloadList)
	}
	chooseWorkload := workloadList[startIndex:endIndex]
	response := make([]*bcsdatamanager.Workload, 0)
	for _, workload := range chooseWorkload {
		workloadRequest := &bcsdatamanager.GetWorkloadInfoRequest{
			ClusterID:    request.ClusterID,
			Namespace:    request.Namespace,
			Dimension:    dimension,
			WorkloadType: request.WorkloadType,
			WorkloadName: workload,
		}
		namespaceInfo, err := m.GetWorkloadInfo(ctx, workloadRequest)
		if err != nil {
			blog.Errorf("get workload[%s] info err:%v", workload, err)
		} else {
			response = append(response, namespaceInfo)
		}
	}
	return response, total, nil
}

// GetWorkloadInfo get workload data with default time range by cluster id, namespace, workload type and name
func (m *ModelWorkload) GetWorkloadInfo(ctx context.Context,
	request *bcsdatamanager.GetWorkloadInfoRequest) (*bcsdatamanager.Workload, error) {
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	workloadMetricsMap := make([]map[string]*common.WorkloadMetrics, 0)
	publicCond := operator.NewLeafCondition(operator.Eq, operator.M{
		ClusterIDKey:    request.ClusterID,
		ObjectTypeKey:   common.NamespaceType,
		NamespaceKey:    request.Namespace,
		WorkloadNameKey: request.WorkloadName,
		WorkloadTypeKey: request.WorkloadType,
	})
	workloadPublic := common.WorkloadPublicMetrics{
		SuggestCPU:    0,
		SuggestMemory: 0,
	}
	publicData := getPublicData(ctx, m.DB, publicCond)
	if publicData != nil && publicData.Metrics != nil {
		public, ok := publicData.Metrics.(common.WorkloadPublicMetrics)
		if !ok {
			blog.Errorf("assert public data to namespace public failed")
		} else {
			workloadPublic = public
		}
	}

	dimension := request.Dimension
	if dimension == "" {
		dimension = common.DimensionMinute
	}
	pipeline := make([]map[string]interface{}, 0)
	pipeline = append(pipeline, map[string]interface{}{"$unwind": "$metrics"})
	pipeline = append(pipeline, map[string]interface{}{"$match": map[string]interface{}{
		ClusterIDKey:    request.ClusterID,
		DimensionKey:    dimension,
		NamespaceKey:    request.Namespace,
		WorkloadTypeKey: request.WorkloadType,
		WorkloadNameKey: request.WorkloadName,
		MetricTimeKey: map[string]interface{}{
			"$gte": primitive.NewDateTimeFromTime(getStartTime(dimension)),
		},
	}})
	pipeline = append(pipeline, map[string]interface{}{"$project": map[string]interface{}{
		"_id":     0,
		"metrics": 1,
	}})
	err = m.DB.Table(m.TableName).Aggregation(ctx, pipeline, &workloadMetricsMap)
	if err != nil {
		blog.Errorf("find workload data fail, err:%v", err)
		return nil, err
	}
	if len(workloadMetricsMap) == 0 {
		return &bcsdatamanager.Workload{}, nil
	}
	workloadMetrics := make([]*common.WorkloadMetrics, 0)
	for _, metrics := range workloadMetricsMap {
		workloadMetrics = append(workloadMetrics, metrics["metrics"])
	}
	startTime := workloadMetrics[0].Time.Time().String()
	endTime := workloadMetrics[len(workloadMetrics)-1].Time.Time().String()
	return m.generateWorkloadResponse(workloadPublic, workloadMetrics, request.ClusterID, request.Namespace,
		dimension, request.WorkloadType, request.WorkloadName, startTime, endTime), nil
}

// GetRawWorkloadInfo get raw workload data
func (m *ModelWorkload) GetRawWorkloadInfo(ctx context.Context, opts *common.JobCommonOpts,
	bucket string) ([]*common.WorkloadData, error) {
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	cond := m.generateCond(opts, bucket)
	conds := operator.NewBranchCondition(operator.And, cond...)
	retWorkload := make([]*common.WorkloadData, 0)
	err = m.DB.Table(m.TableName).Find(conds).All(ctx, &retWorkload)
	if err != nil {
		return nil, err
	}
	return retWorkload, nil
}

// GetWorkloadCount get raw workload data
func (m *ModelWorkload) GetWorkloadCount(ctx context.Context, opts *common.JobCommonOpts,
	bucket string, after time.Time) (int64, error) {
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return 0, err
	}
	cond := m.generateCond(opts, bucket)
	cond = append(cond, operator.NewLeafCondition(operator.Gte, operator.M{
		MetricTimeKey: primitive.NewDateTimeFromTime(after),
	}))
	conds := operator.NewBranchCondition(operator.And, cond...)
	retWorkload := make([]*common.WorkloadData, 0)
	err = m.DB.Table(m.TableName).Find(conds).All(ctx, &retWorkload)
	if err != nil {
		return 0, err
	}
	return int64(len(retWorkload)), nil
}

func (m *ModelWorkload) generateWorkloadResponse(public common.WorkloadPublicMetrics,
	metricSlice []*common.WorkloadMetrics, clusterID, namespace, dimension, workloadType, workloadName, startTime,
	endTime string) *bcsdatamanager.Workload {
	response := &bcsdatamanager.Workload{
		ClusterID:     clusterID,
		Dimension:     dimension,
		StartTime:     startTime,
		EndTime:       endTime,
		Namespace:     namespace,
		WorkloadType:  workloadType,
		WorkloadName:  workloadName,
		Metrics:       nil,
		SuggestCPU:    strconv.FormatFloat(public.SuggestCPU, 'f', 2, 64),
		SuggestMemory: strconv.FormatFloat(public.SuggestMemory, 'f', 2, 64),
	}
	responseMetrics := make([]*bcsdatamanager.WorkloadMetrics, 0)
	for _, metric := range metricSlice {
		responseMetric := &bcsdatamanager.WorkloadMetrics{
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
			InstanceCount:      strconv.FormatInt(metric.InstanceCount, 10),
			MinInstanceTime:    metric.MinInstanceTime,
			MaxInstanceTime:    metric.MaxInstanceTime,
		}
		responseMetrics = append(responseMetrics, responseMetric)
	}
	response.Metrics = responseMetrics
	return response
}

// pre aggregate max value before update
func (m *ModelWorkload) preAggregateMax(data *common.WorkloadData, newMetric *common.WorkloadMetrics) {
	if data.MaxInstanceTime != nil && newMetric.MaxInstanceTime != nil {
		data.MaxInstanceTime = getMax(data.MaxInstanceTime, newMetric.MaxInstanceTime)
	} else if newMetric.MaxInstanceTime != nil {
		data.MaxInstanceTime = newMetric.MaxInstanceTime
	}

	if data.MaxCPUUsageTime != nil && newMetric.MaxCPUUsageTime != nil {
		data.MaxCPUUsageTime = getMax(data.MaxCPUUsageTime, newMetric.MaxCPUUsageTime)
	} else if newMetric.MaxCPUUsageTime != nil {
		data.MaxCPUUsageTime = newMetric.MaxCPUUsageTime
	}

	if data.MaxMemoryUsageTime != nil && newMetric.MaxMemoryUsageTime != nil {
		data.MaxMemoryUsageTime = getMax(data.MaxMemoryUsageTime, newMetric.MaxMemoryUsageTime)
	} else if newMetric.MaxMemoryUsageTime != nil {
		data.MaxMemoryUsageTime = newMetric.MaxMemoryUsageTime
	}

	if data.MaxCPUTime != nil && newMetric.MaxCPUTime != nil {
		data.MaxCPUTime = getMax(data.MaxCPUTime, newMetric.MaxCPUTime)
	} else if newMetric.MaxCPUTime != nil {
		data.MaxCPUTime = newMetric.MaxCPUTime
	}

	if data.MaxMemoryTime != nil && newMetric.MaxMemoryTime != nil {
		data.MaxMemoryTime = getMax(data.MaxMemoryTime, newMetric.MaxMemoryTime)
	} else if newMetric.MaxMemoryTime != nil {
		data.MaxMemoryTime = newMetric.MaxMemoryTime
	}
}

// pre aggragate min value before update
func (m *ModelWorkload) preAggregateMin(data *common.WorkloadData, newMetric *common.WorkloadMetrics) {
	if data.MinInstanceTime != nil && newMetric.MinInstanceTime != nil {
		data.MinInstanceTime = getMin(data.MinInstanceTime, newMetric.MinInstanceTime)
	} else if newMetric.MinInstanceTime != nil {
		data.MinInstanceTime = newMetric.MinInstanceTime
	}

	if data.MinCPUUsageTime != nil && newMetric.MinCPUUsageTime != nil {
		data.MinCPUUsageTime = getMin(data.MinCPUUsageTime, newMetric.MinCPUUsageTime)
	} else if newMetric.MinCPUUsageTime != nil {
		data.MinCPUUsageTime = newMetric.MinCPUUsageTime
	}

	if data.MinMemoryUsageTime != nil && newMetric.MinMemoryUsageTime != nil {
		data.MinMemoryUsageTime = getMin(data.MinMemoryUsageTime, newMetric.MinMemoryUsageTime)
	} else if newMetric.MinMemoryUsageTime != nil {
		data.MinMemoryUsageTime = newMetric.MinMemoryUsageTime
	}

	if data.MinCPUTime != nil && newMetric.MinCPUTime != nil {
		data.MinCPUTime = getMin(data.MinCPUTime, newMetric.MinCPUTime)
	} else if newMetric.MinCPUTime != nil {
		data.MinCPUTime = newMetric.MinCPUTime
	}

	if data.MinMemoryTime != nil && newMetric.MinMemoryTime != nil {
		data.MinMemoryTime = getMin(data.MinMemoryTime, newMetric.MinMemoryTime)
	} else if newMetric.MinMemoryTime != nil {
		data.MinMemoryTime = newMetric.MinMemoryTime
	}
}

func getMax(old *bcsdatamanager.ExtremumRecord, new *bcsdatamanager.ExtremumRecord) *bcsdatamanager.ExtremumRecord {
	if old.Value > new.Value {
		return old
	}
	return new
}

func getMin(old *bcsdatamanager.ExtremumRecord, new *bcsdatamanager.ExtremumRecord) *bcsdatamanager.ExtremumRecord {
	if old.Value < new.Value {
		return old
	}
	return new
}

// generate cond according to job options
func (m *ModelWorkload) generateCond(opts *common.JobCommonOpts, bucket string) []*operator.Condition {
	cond := make([]*operator.Condition, 0)
	if opts.ProjectID != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			ProjectIDKey: opts.ProjectID,
		}))
	}
	if opts.Dimension != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			DimensionKey: opts.Dimension,
		}))
	}
	if opts.ClusterID != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			ClusterIDKey: opts.ClusterID,
		}))
	}
	if opts.Namespace != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			NamespaceKey: opts.Namespace,
		}))
	}
	if opts.WorkloadType != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			WorkloadTypeKey: opts.WorkloadType,
		}))
	}
	if opts.Name != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			WorkloadNameKey: opts.Name,
		}))
	}
	if bucket != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			BucketTimeKey: bucket,
		}))
	}
	return cond
}
