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
	"errors"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/utils"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
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
			Name: types.WorkloadTableName + "_idx",
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
			Name: types.WorkloadTableName + "_list_idx",
			Key: bson.D{
				bson.E{Key: ClusterIDKey, Value: 1},
				bson.E{Key: NamespaceKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: WorkloadTypeKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
		},
		{
			Name: types.WorkloadTableName + "_list_idx2",
			Key: bson.D{
				bson.E{Key: ClusterIDKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
			Background: true,
		},
		{
			Name: types.WorkloadTableName + "_list_idx3",
			Key: bson.D{
				bson.E{Key: ClusterIDKey, Value: 1},
				bson.E{Key: NamespaceKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
			Background: true,
		},
		{
			Name: types.WorkloadTableName + "_get_idx",
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
	newModelWorkloadIndexes = []drivers.Index{
		{
			Name: types.WorkloadTableName + "_idx",
			Key: bson.D{
				bson.E{Key: NamespaceKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: WorkloadTypeKey, Value: 1},
				bson.E{Key: WorkloadNameKey, Value: 1},
				bson.E{Key: BucketTimeKey, Value: 1},
			},
			Unique:     true,
			Background: true,
		},
		{
			Name: types.WorkloadTableName + "_get_idx",
			Key: bson.D{
				bson.E{Key: NamespaceKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: WorkloadTypeKey, Value: 1},
				bson.E{Key: WorkloadNameKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
			Background: true,
		}, {
			Name: types.WorkloadTableName + "_list_idx",
			Key: bson.D{
				bson.E{Key: NamespaceKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: WorkloadTypeKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
			Background: true,
		},
		{
			Name: types.WorkloadTableName + "_list_idx2",
			Key: bson.D{
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
			Background: true,
		},
		{
			Name: types.WorkloadTableName + "_list_idx3",
			Key: bson.D{
				bson.E{Key: NamespaceKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
			Background: true,
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
		TableName: types.DataTableNamePrefix + types.WorkloadTableName,
		Indexes:   modelWorkloadIndexes,
		DB:        db,
	}}
}

// InsertWorkloadInfo insert workload info
// It takes in the ctx, metrics, and opts parameters.
// It inserts the workload information into the database.
func (m *ModelWorkload) InsertWorkloadInfo(ctx context.Context, metrics *types.WorkloadMetrics,
	opts *types.JobCommonOpts) error {
	newTableInfo := &Public{
		TableName: types.WorkloadTableName + "_" + opts.ClusterID,
		Indexes:   newModelWorkloadIndexes,
		DB:        m.DB,
	}
	err := ensureTable(ctx, newTableInfo)
	if err != nil {
		return err
	}
	err = ensureTable(ctx, &m.Public)
	if err != nil {
		return err
	}
	bucketTime, err := utils.GetBucketTime(opts.CurrentTime, opts.Dimension)
	if err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		ProjectIDKey:    opts.ProjectID,
		ClusterIDKey:    opts.ClusterID,
		NamespaceKey:    opts.Namespace,
		DimensionKey:    opts.Dimension,
		WorkloadTypeKey: map[string]string{"$regex": opts.WorkloadType, "$options": "$i"},
		WorkloadNameKey: opts.WorkloadName,
		BucketTimeKey:   bucketTime,
	})
	newCond := operator.NewLeafCondition(operator.Eq, operator.M{
		NamespaceKey:    opts.Namespace,
		DimensionKey:    opts.Dimension,
		WorkloadTypeKey: map[string]string{"$regex": opts.WorkloadType, "$options": "$i"},
		WorkloadNameKey: opts.WorkloadName,
		BucketTimeKey:   bucketTime,
	})
	retWorkload := &types.WorkloadData{}
	var isGetFromOldCollection bool
	err = m.DB.Table(newTableInfo.TableName).Find(newCond).One(ctx, retWorkload)
	if err != nil {
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Infof("find workload info from new collection failed:%s", err.Error())
			err = m.DB.Table(m.TableName).Find(cond).One(ctx, retWorkload)
			if err != nil {
				if errors.Is(err, drivers.ErrTableRecordNotFound) {
					blog.Infof(" workload info not found, create a new bucket")
					newMetrics := make([]*types.WorkloadMetrics, 0)
					newMetrics = append(newMetrics, metrics)
					newWorkloadBucket := &types.WorkloadData{
						CreateTime:   primitive.NewDateTimeFromTime(time.Now()),
						UpdateTime:   primitive.NewDateTimeFromTime(time.Now()),
						BucketTime:   bucketTime,
						Dimension:    opts.Dimension,
						ProjectID:    opts.ProjectID,
						BusinessID:   opts.BusinessID,
						ClusterID:    opts.ClusterID,
						ClusterType:  opts.ClusterType,
						Namespace:    opts.Namespace,
						WorkloadType: opts.WorkloadType,
						Name:         opts.WorkloadName,
						Metrics:      newMetrics,
					}
					m.preAggregateMax(newWorkloadBucket, metrics)
					m.preAggregateMin(newWorkloadBucket, metrics)
					_, err = m.DB.Table(newTableInfo.TableName).Insert(ctx, []interface{}{newWorkloadBucket})
					if err != nil {
						return err
					}
					return nil
				}
				return err
			}
			isGetFromOldCollection = true
		} else {
			return err
		}
	}
	m.preAggregateMax(retWorkload, metrics)
	m.preAggregateMin(retWorkload, metrics)
	if retWorkload.BusinessID == "" {
		retWorkload.BusinessID = opts.BusinessID
	}
	retWorkload.Label = opts.Label
	retWorkload.ProjectCode = opts.ProjectCode
	retWorkload.UpdateTime = primitive.NewDateTimeFromTime(time.Now())
	retWorkload.Metrics = append(retWorkload.Metrics, metrics)
	if !isGetFromOldCollection {
		return m.DB.Table(newTableInfo.TableName).
			Update(ctx, newCond, operator.M{"$set": retWorkload})
	}
	_, err = m.DB.Table(newTableInfo.TableName).
		Insert(ctx, []interface{}{retWorkload})
	if err == nil {
		_, err = m.DB.Table(m.TableName).Delete(ctx, cond)
		return err
	}
	return err
}

// GetWorkloadInfoList get workload list data by cluster id, namespace and workload type
// if startTime or endTime is empty, return metrics with default time range
// It takes in the ctx and request parameters.
// If the startTime or endTime parameters are empty, it returns metrics with the default time range.
// It retrieves the workload data from the database and returns it along with the count of workloads.
func (m *ModelWorkload) GetWorkloadInfoList(ctx context.Context,
	request *bcsdatamanager.GetWorkloadInfoListRequest) ([]*bcsdatamanager.Workload, int64, error) {
	var total int64
	newTableInfo := &Public{
		TableName: types.WorkloadTableName + "_" + request.ClusterID,
		Indexes:   newModelWorkloadIndexes,
		DB:        m.DB,
	}
	err := ensureTable(ctx, newTableInfo)
	if err != nil {
		return nil, total, err
	}
	err = ensureTable(ctx, &m.Public)
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
			DimensionKey: dimension,
		}))
	if request.Namespace != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			NamespaceKey: request.Namespace,
		}))
	}
	if request.WorkloadType != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			WorkloadTypeKey: map[string]string{"$regex": request.WorkloadType, "$options": "$i"},
		}))
	}
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
	tempWorkloadList := make([]map[string]string, 0)
	err = m.DB.Table(newTableInfo.TableName).Find(conds).WithProjection(map[string]int{ProjectIDKey: 1, ClusterIDKey: 1,
		NamespaceKey: 1, WorkloadTypeKey: 1, WorkloadNameKey: 1}).
		All(ctx, &tempWorkloadList)
	if err != nil {
		blog.Errorf("get workload list error")
		return nil, total, err
	}
	workloadList := distinctWorkloadSlice(&tempWorkloadList)
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
			ClusterID:    workload[ClusterIDKey],
			Namespace:    workload[NamespaceKey],
			Dimension:    dimension,
			WorkloadType: workload[WorkloadTypeKey],
			WorkloadName: workload[WorkloadNameKey],
			StartTime:    request.GetStartTime(),
			EndTime:      request.GetEndTime(),
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
// It takes in the ctx and request parameters.
// If the startTime or endTime parameters are empty, it returns metrics with the default time range.
// It retrieves the workload data from the database and returns it.
func (m *ModelWorkload) GetWorkloadInfo(ctx context.Context,
	request *bcsdatamanager.GetWorkloadInfoRequest) (*bcsdatamanager.Workload, error) {
	newTableInfo := &Public{
		TableName: types.WorkloadTableName + "_" + request.ClusterID,
		Indexes:   newModelWorkloadIndexes,
		DB:        m.DB,
	}
	err := ensureTable(ctx, newTableInfo)
	if err != nil {
		return nil, err
	}
	err = ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	workloadMetricsMap := make([]*types.WorkloadData, 0)
	publicCond := operator.NewLeafCondition(operator.Eq, operator.M{
		ClusterIDKey:    request.ClusterID,
		ObjectTypeKey:   types.NamespaceType,
		NamespaceKey:    request.Namespace,
		WorkloadNameKey: request.WorkloadName,
		WorkloadTypeKey: map[string]string{"$regex": request.WorkloadType, "$options": "$i"},
	})
	workloadPublic := types.WorkloadPublicMetrics{
		SuggestCPU:    0,
		SuggestMemory: 0,
	}
	publicData := getPublicData(ctx, m.DB, publicCond)
	if publicData != nil && publicData.Metrics != nil {
		public, ok := publicData.Metrics.(types.WorkloadPublicMetrics)
		if !ok {
			blog.Errorf("assert public data to namespace public failed")
		} else {
			workloadPublic = public
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
		ClusterIDKey:    request.ClusterID,
		DimensionKey:    dimension,
		NamespaceKey:    request.Namespace,
		WorkloadTypeKey: map[string]string{"$regex": request.WorkloadType, "$options": "$i"},
		WorkloadNameKey: request.WorkloadName,
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
			"_id":           0,
			"metrics":       1,
			"business_id":   1,
			"project_id":    1,
			"project_code":  1,
			"namespace":     1,
			"cluster_id":    1,
			"workload_name": 1,
			"workload_type": 1,
			"label":         1,
		}}, map[string]interface{}{"$group": map[string]interface{}{
			"_id":           nil,
			"cluster_id":    map[string]interface{}{"$first": "$cluster_id"},
			"namespace":     map[string]interface{}{"$first": "$namespace"},
			"project_id":    map[string]interface{}{"$first": "$project_id"},
			"workload_type": map[string]interface{}{"$first": "$workload_type"},
			"workload_name": map[string]interface{}{"$first": "$workload_name"},
			"business_id":   map[string]interface{}{"$max": "$business_id"},
			"metrics":       map[string]interface{}{"$push": "$metrics"},
			"label":         map[string]interface{}{"$max": "$label"},
			"project_code":  map[string]interface{}{"$max": "$project_code"},
		}},
	)
	err = m.DB.Table(m.TableName).Aggregation(ctx, pipeline, &workloadMetricsMap)
	if err != nil {
		blog.Errorf("find workload data fail, err:%v", err)
		return nil, err
	}
	newWorkloadMetricsMap := make([]*types.WorkloadData, 0)
	err = m.DB.Table(newTableInfo.TableName).Aggregation(ctx, pipeline, &newWorkloadMetricsMap)
	if err != nil {
		blog.Errorf("find workload data fail, err:%v", err)
		return nil, err
	}
	workloadMetrics := make([]*types.WorkloadMetrics, 0)
	workloadInfo := &types.WorkloadData{}
	for _, metrics := range workloadMetricsMap {
		workloadMetrics = append(workloadMetrics, metrics.Metrics...)
		workloadInfo = metrics
	}
	for _, metrics := range newWorkloadMetricsMap {
		workloadMetrics = append(workloadMetrics, metrics.Metrics...)
		workloadInfo = metrics
	}
	if len(workloadMetrics) == 0 {
		return &bcsdatamanager.Workload{}, nil
	}
	startTime := workloadMetrics[0].Time.Time().String()
	endTime := workloadMetrics[len(workloadMetrics)-1].Time.Time().String()
	return m.generateWorkloadResponse(workloadPublic, workloadMetrics, workloadInfo, startTime, endTime), nil
}

// GetRawWorkloadInfo get raw workload data
// It takes in the ctx, opts, and bucket parameters.
// It generates a condition based on the opts and bucket parameters.
// It retrieves the workload data from the database and returns it.
func (m *ModelWorkload) GetRawWorkloadInfo(ctx context.Context, opts *types.JobCommonOpts,
	bucket string) ([]*types.WorkloadData, error) {
	newTableInfo := &Public{
		TableName: types.WorkloadTableName + "_" + opts.ClusterID,
		Indexes:   newModelWorkloadIndexes,
		DB:        m.DB,
	}
	// Ensure that the table exists.
	err := ensureTable(ctx, newTableInfo)
	if err != nil {
		return nil, err
	}
	err = ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	cond := m.generateCond(opts, bucket)
	conds := operator.NewBranchCondition(operator.And, cond...)
	retWorkload := make([]*types.WorkloadData, 0)
	err = m.DB.Table(newTableInfo.TableName).Find(conds).All(ctx, &retWorkload)
	if err != nil {
		return nil, err
	}
	oldRetWorkload := make([]*types.WorkloadData, 0)
	err = m.DB.Table(m.TableName).Find(conds).All(ctx, &oldRetWorkload)
	if err != nil {
		return nil, err
	}
	retWorkload = append(retWorkload, oldRetWorkload...)
	return retWorkload, nil
}

// GetWorkloadCount get workload count
// It takes in the ctx, opts, bucket, and after parameters.
// It generates a condition based on the opts and bucket parameters.
// If the after parameter is not zero, it adds a condition to only retrieve data after that time.
// It retrieves the workload data from the database and returns the count.
func (m *ModelWorkload) GetWorkloadCount(ctx context.Context, opts *types.JobCommonOpts,
	bucket string, after time.Time) (int64, error) {
	newTableInfo := &Public{
		TableName: types.WorkloadTableName + "_" + opts.ClusterID,
		Indexes:   newModelWorkloadIndexes,
		DB:        m.DB,
	}
	// Ensure that the table exists.
	err := ensureTable(ctx, newTableInfo)
	if err != nil {
		return 0, err
	}
	err = ensureTable(ctx, &m.Public)
	if err != nil {
		return 0, err
	}
	cond := m.generateCond(opts, bucket)
	if !after.IsZero() {
		cond = append(cond, operator.NewLeafCondition(operator.Gte, operator.M{
			MetricTimeKey: primitive.NewDateTimeFromTime(after),
		}))
	}
	conds := operator.NewBranchCondition(operator.And, cond...)
	retWorkload := make([]*types.WorkloadData, 0)
	err = m.DB.Table(newTableInfo.TableName).Find(conds).All(ctx, &retWorkload)
	if err != nil {
		return 0, err
	}
	return int64(len(retWorkload)), nil
}

// generateWorkloadResponse 构造response，将storage结构转化为proto结构
func (m *ModelWorkload) generateWorkloadResponse(public types.WorkloadPublicMetrics,
	metricSlice []*types.WorkloadMetrics, data *types.WorkloadData, startTime,
	endTime string) *bcsdatamanager.Workload {
	response := &bcsdatamanager.Workload{
		ProjectID:     data.ProjectID,
		ProjectCode:   data.ProjectCode,
		BusinessID:    data.BusinessID,
		ClusterID:     data.ClusterID,
		Dimension:     data.Dimension,
		StartTime:     startTime,
		EndTime:       endTime,
		Namespace:     data.Namespace,
		WorkloadType:  data.WorkloadType,
		WorkloadName:  data.Name,
		Label:         data.Label,
		Metrics:       nil,
		SuggestCPU:    strconv.FormatFloat(public.SuggestCPU, 'f', 2, 64),
		SuggestMemory: strconv.FormatFloat(public.SuggestMemory, 'f', 2, 64),
	}
	responseMetrics := make([]*bcsdatamanager.WorkloadMetrics, 0)
	for _, metric := range metricSlice {
		responseMetric := &bcsdatamanager.WorkloadMetrics{
			Time:               metric.Time.Time().String(),
			CPURequest:         strconv.FormatFloat(metric.CPURequest, 'f', 2, 64),
			CPULimit:           strconv.FormatFloat(metric.CPULimit, 'f', 2, 64),
			MemoryRequest:      strconv.FormatInt(metric.MemoryRequest, 10),
			MemoryLimit:        strconv.FormatInt(metric.MemoryLimit, 10),
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

// preAggregateMax xxx
// pre aggregate max value before update
// 预聚合最大值
func (m *ModelWorkload) preAggregateMax(data *types.WorkloadData, newMetric *types.WorkloadMetrics) {
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

// preAggregateMin xxx
// pre aggregate min value before update
// 预聚合最小值
func (m *ModelWorkload) preAggregateMin(data *types.WorkloadData, newMetric *types.WorkloadMetrics) {
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

// getMax 获取最大值
func getMax(old *bcsdatamanager.ExtremumRecord, new *bcsdatamanager.ExtremumRecord) *bcsdatamanager.ExtremumRecord {
	if old.Value > new.Value {
		return old
	}
	return new
}

// getMin 获取最小值
func getMin(old *bcsdatamanager.ExtremumRecord, new *bcsdatamanager.ExtremumRecord) *bcsdatamanager.ExtremumRecord {
	if old.Value < new.Value {
		return old
	}
	return new
}

// generateCond cond according to job options
// 构造查询条件，需要按顺序
func (m *ModelWorkload) generateCond(opts *types.JobCommonOpts, bucket string) []*operator.Condition {
	cond := make([]*operator.Condition, 0)
	if opts.ClusterID != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			ClusterIDKey: opts.ClusterID,
		}))
	}
	if opts.Dimension != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			DimensionKey: opts.Dimension,
		}))
	}
	if opts.Namespace != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			NamespaceKey: opts.Namespace,
		}))
	}
	if opts.WorkloadType != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			WorkloadTypeKey: map[string]string{"$regex": opts.WorkloadType, "$options": "$i"},
		}))
	}
	if opts.WorkloadName != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			WorkloadNameKey: opts.WorkloadName,
		}))
	}
	if bucket != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			BucketTimeKey: bucket,
		}))
	}
	return cond
}
