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
	modelPodAutoscalerIndexes = []drivers.Index{
		{
			Key: bson.D{
				bson.E{Key: CreateTimeKey, Value: 1},
			},
			Name: CreateTimeKey + "_1",
		},
		{
			Name: types.PodAutoscalerTableName + "_idx",
			Key: bson.D{
				bson.E{Key: ProjectIDKey, Value: 1},
				bson.E{Key: ClusterIDKey, Value: 1},
				bson.E{Key: NamespaceKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: PodAutoscalerTypeKey, Value: 1},
				bson.E{Key: PodAutoscalerNameKey, Value: 1},
				bson.E{Key: BucketTimeKey, Value: 1},
			},
			Unique: true,
		},
		{
			Name: types.PodAutoscalerTableName + "_list_idx",
			Key: bson.D{
				bson.E{Key: ClusterIDKey, Value: 1},
				bson.E{Key: NamespaceKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: PodAutoscalerTypeKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
		},
		{
			Name: types.PodAutoscalerTableName + "_list_idx3",
			Key: bson.D{
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
			Background: true,
		},
		{
			Name: types.PodAutoscalerTableName + "_get_idx",
			Key: bson.D{
				bson.E{Key: ClusterIDKey, Value: 1},
				bson.E{Key: NamespaceKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: PodAutoscalerTypeKey, Value: 1},
				bson.E{Key: PodAutoscalerNameKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
		},
	}
)

// ModelPodAutoscaler podAutoscaler model
type ModelPodAutoscaler struct {
	Public
}

// NewModelPodAutoscaler new podAutoscaler model
func NewModelPodAutoscaler(db drivers.DB) *ModelPodAutoscaler {
	return &ModelPodAutoscaler{Public: Public{
		TableName: types.DataTableNamePrefix + types.PodAutoscalerTableName,
		Indexes:   modelPodAutoscalerIndexes,
		DB:        db,
	}}
}

// InsertPodAutoscalerInfo inserts the given pod autoscaler metrics into the database.
// It returns an error (if any).
func (m *ModelPodAutoscaler) InsertPodAutoscalerInfo(ctx context.Context, metrics *types.PodAutoscalerMetrics,
	opts *types.JobCommonOpts) error {
	// Ensure that the table exists in the database.
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return err
	}
	// Get the bucket time for the given current time and dimension.
	bucketTime, err := utils.GetBucketTime(opts.CurrentTime, opts.Dimension)
	if err != nil {
		return err
	}
	// Create a condition to find the pod autoscaler data for the given parameters.
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		ProjectIDKey:         opts.ProjectID,
		ClusterIDKey:         opts.ClusterID,
		NamespaceKey:         opts.Namespace,
		DimensionKey:         opts.Dimension,
		PodAutoscalerTypeKey: opts.PodAutoscalerType,
		PodAutoscalerNameKey: opts.PodAutoscalerName,
		BucketTimeKey:        bucketTime,
	})
	// Find the pod autoscaler data that matches the condition.
	retPodAutoscaler := &types.PodAutoscalerData{}
	err = m.DB.Table(m.TableName).Find(cond).One(ctx, retPodAutoscaler)
	if err != nil {
		// If the pod autoscaler data is not found, create a new bucket and insert the metrics.
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Infof(" podAutoscaler info not found, create a new bucket")
			newMetrics := make([]*types.PodAutoscalerMetrics, 0)
			newMetrics = append(newMetrics, metrics)
			newPodAutoscalerBucket := &types.PodAutoscalerData{
				CreateTime:        primitive.NewDateTimeFromTime(time.Now()),
				UpdateTime:        primitive.NewDateTimeFromTime(time.Now()),
				BucketTime:        bucketTime,
				Dimension:         opts.Dimension,
				ProjectID:         opts.ProjectID,
				BusinessID:        opts.BusinessID,
				ClusterID:         opts.ClusterID,
				ClusterType:       opts.ClusterType,
				Namespace:         opts.Namespace,
				WorkloadType:      opts.WorkloadType,
				WorkloadName:      opts.WorkloadName,
				PodAutoscalerType: opts.PodAutoscalerType,
				PodAutoscalerName: opts.PodAutoscalerName,
				Total:             metrics.TotalSuccessfulRescale,
				Metrics:           newMetrics,
			}
			_, err = m.DB.Table(m.TableName).Insert(ctx, []interface{}{newPodAutoscalerBucket})
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	// Update the existing pod autoscaler data with the new metrics.
	retPodAutoscaler.UpdateTime = primitive.NewDateTimeFromTime(time.Now())
	retPodAutoscaler.Metrics = append(retPodAutoscaler.Metrics, metrics)
	retPodAutoscaler.Total += metrics.TotalSuccessfulRescale
	retPodAutoscaler.Label = opts.Label
	retPodAutoscaler.ProjectCode = opts.ProjectCode
	return m.DB.Table(m.TableName).
		Update(ctx, cond, operator.M{"$set": retPodAutoscaler})
}

// GetPodAutoscalerList retrieves a list of pod autoscalers that match the given request parameters.
// It returns the list of pod autoscalers, the total number of pod autoscalers that match the request,
// and an error (if any).
func (m *ModelPodAutoscaler) GetPodAutoscalerList(ctx context.Context,
	request *bcsdatamanager.GetPodAutoscalerListRequest) ([]*bcsdatamanager.PodAutoscaler, int64, error) {
	var total int64
	// Ensure that the table exists in the database.
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, total, err
	}
	// Set the dimension to "minute" if it is not specified in the request.
	dimension := request.Dimension
	if dimension == "" {
		dimension = types.DimensionMinute
	}
	// Generate the list condition based on the request parameters.
	cond := genPodAutoscalerListCond(request)
	// Set the start time based on the dimension and the request parameters.
	startTime := getStartTime(dimension)
	if request.GetStartTime() != 0 {
		startTime = time.Unix(request.GetStartTime(), 0)
	}
	cond = append(cond, operator.NewLeafCondition(operator.Gte, operator.M{
		MetricTimeKey: primitive.NewDateTimeFromTime(startTime),
	}))
	// Set the end time based on the request parameters.
	if request.GetEndTime() != 0 {
		cond = append(cond, operator.NewLeafCondition(operator.Lte, operator.M{
			MetricTimeKey: primitive.NewDateTimeFromTime(time.Unix(request.GetEndTime(), 0)),
		}))
	}

	// Combine the conditions into a single branch condition.
	conds := operator.NewBranchCondition(operator.And, cond...)
	// Retrieve the list of pod autoscalers that match the conditions.
	tempList := make([]map[string]string, 0)
	err = m.DB.Table(m.TableName).Find(conds).WithProjection(
		map[string]int{ProjectIDKey: 1, ClusterIDKey: 1, NamespaceKey: 1,
			PodAutoscalerTypeKey: 1, PodAutoscalerNameKey: 1},
	).All(ctx, &tempList)
	if err != nil {
		blog.Errorf("get pod autoscaler list error")
		return nil, total, err
	}
	// Remove duplicates from the list of pod autoscalers.
	autoscalerList := distinctPodAutoscaler(&tempList)
	// If there are no pod autoscalers in the list, return an empty response.
	if len(autoscalerList) == 0 {
		return nil, total, nil
	}
	// Set the page and size parameters based on the request.
	page := int(request.Page)
	size := int(request.Size)
	if size == 0 {
		size = DefaultSize
	}
	// Calculate the start and end indices for the sublist of pod autoscalers to return.
	endIndex := (page + 1) * size
	startIndex := page * size
	if startIndex >= len(autoscalerList) {
		return nil, total, nil
	}
	if endIndex >= len(autoscalerList) {
		endIndex = len(autoscalerList)
	}
	// Retrieve the information for each pod autoscaler in the sublist.
	chooseAutoscaler := autoscalerList[startIndex:endIndex]
	response := make([]*bcsdatamanager.PodAutoscaler, 0)
	for _, autoscaler := range chooseAutoscaler {
		podAutoscalerRequest := &bcsdatamanager.GetPodAutoscalerRequest{
			ClusterID:         autoscaler[ClusterIDKey],
			Namespace:         autoscaler[NamespaceKey],
			Dimension:         dimension,
			PodAutoscalerType: autoscaler[PodAutoscalerTypeKey],
			PodAutoscalerName: autoscaler[PodAutoscalerNameKey],
			StartTime:         request.GetStartTime(),
			EndTime:           request.GetEndTime(),
		}
		autoscalerInfo, err := m.GetPodAutoscalerInfo(ctx, podAutoscalerRequest)
		if err != nil {
			blog.Errorf("get autoscaler[%s] info err:%v", autoscaler, err)
		} else {
			response = append(response, autoscalerInfo)
		}
	}
	// Return the list of pod autoscalers, the total number of pod autoscalers that match the request,
	// and any error that occurred.
	return response, total, nil
}

// GetPodAutoscalerInfo get podAutoscaler data with default time range by cluster id, namespace, workload type and name
func (m *ModelPodAutoscaler) GetPodAutoscalerInfo(ctx context.Context,
	request *bcsdatamanager.GetPodAutoscalerRequest) (*bcsdatamanager.PodAutoscaler, error) {
	// Ensure that the table exists in the database.
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	autoscalerMetricsMap := make([]*types.PodAutoscalerData, 0)
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
		ClusterIDKey:         request.ClusterID,
		DimensionKey:         dimension,
		NamespaceKey:         request.Namespace,
		PodAutoscalerTypeKey: request.PodAutoscalerType,
		PodAutoscalerNameKey: request.PodAutoscalerName,
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
			"_id":                 0,
			"metrics":             1,
			"business_id":         1,
			"project_id":          1,
			"project_code":        1,
			"namespace":           1,
			"cluster_id":          1,
			"workload_name":       1,
			"workload_type":       1,
			"pod_autoscaler_type": 1,
			"pod_autoscaler_name": 1,
			"label":               1,
		}}, map[string]interface{}{"$group": map[string]interface{}{
			"_id":                 nil,
			"cluster_id":          map[string]interface{}{"$first": "$cluster_id"},
			"namespace":           map[string]interface{}{"$first": "$namespace"},
			"project_id":          map[string]interface{}{"$first": "$project_id"},
			"workload_type":       map[string]interface{}{"$first": "$workload_type"},
			"workload_name":       map[string]interface{}{"$first": "$workload_name"},
			"pod_autoscaler_type": map[string]interface{}{"$first": "$pod_autoscaler_type"},
			"pod_autoscaler_name": map[string]interface{}{"$first": "$pod_autoscaler_name"},
			"business_id":         map[string]interface{}{"$max": "$business_id"},
			"metrics":             map[string]interface{}{"$push": "$metrics"},
			"label":               map[string]interface{}{"$max": "$label"},
			"project_code":        map[string]interface{}{"$max": "$project_code"},
		}},
	)

	err = m.DB.Table(m.TableName).Aggregation(ctx, pipeline, &autoscalerMetricsMap)
	if err != nil {
		blog.Errorf("find autoscaler data fail, err:%v", err)
		return nil, err
	}
	if len(autoscalerMetricsMap) == 0 {
		return &bcsdatamanager.PodAutoscaler{}, nil
	}
	autoscalerMetrics := make([]*types.PodAutoscalerMetrics, 0)
	for _, metrics := range autoscalerMetricsMap {
		autoscalerMetrics = append(autoscalerMetrics, metrics.Metrics...)
	}
	startTime := autoscalerMetrics[0].Time.Time().String()
	endTime := autoscalerMetrics[len(autoscalerMetrics)-1].Time.Time().String()
	return m.generateAutoscalerResponse(autoscalerMetrics, autoscalerMetricsMap[0], startTime, endTime), nil
}

// genPodAutoscalerListCond generate list condition
func genPodAutoscalerListCond(req *bcsdatamanager.GetPodAutoscalerListRequest) []*operator.Condition {
	cond := make([]*operator.Condition, 0)
	dimension := req.GetDimension()
	if req.GetDimension() == "" {
		dimension = types.DimensionMinute
	}
	if req.GetProject() != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			ProjectIDKey: req.GetProject(),
		}))
	}
	if req.GetBusiness() != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			BusinessIDKey: req.GetBusiness(),
		}))
	}
	if req.GetClusterID() != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			ClusterIDKey: req.GetClusterID(),
		}))
	}
	if req.GetNamespace() != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			NamespaceKey: req.GetNamespace(),
		}))
	}
	cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
		DimensionKey: dimension,
	}))
	if req.GetPodAutoscalerType() != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			PodAutoscalerTypeKey: req.GetPodAutoscalerType(),
		}))
	}
	return cond
}

// GetRawPodAutoscalerInfo is a function that retrieves raw pod autoscaler data without a time range.
func (m *ModelPodAutoscaler) GetRawPodAutoscalerInfo(
	ctx context.Context, opts *types.JobCommonOpts, bucket string) ([]*types.PodAutoscalerData, error) {
	// Ensure that the table exists in the database.
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	// Create a condition to filter the database query results.
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		ProjectIDKey:         opts.ProjectID,
		ClusterIDKey:         opts.ClusterID,
		NamespaceKey:         opts.Namespace,
		DimensionKey:         opts.Dimension,
		PodAutoscalerTypeKey: opts.PodAutoscalerType,
		PodAutoscalerNameKey: opts.PodAutoscalerName,
		BucketTimeKey:        bucket,
	})
	// Create an empty slice of PodAutoscalerData to store the results of the database query.
	retAutoscaler := make([]*types.PodAutoscalerData, 0)
	// Query the database with the condition and store the results in retAutoscaler.
	err = m.DB.Table(m.TableName).Find(cond).All(ctx, &retAutoscaler)
	if err != nil {
		return nil, err
	}
	// Return the results.
	return retAutoscaler, nil
}

// generateAutoscalerResponse generate response, transfer storage struct to proto struct
func (m *ModelPodAutoscaler) generateAutoscalerResponse(metricSlice []*types.PodAutoscalerMetrics,
	data *types.PodAutoscalerData, start, end string) *bcsdatamanager.PodAutoscaler {
	response := &bcsdatamanager.PodAutoscaler{
		ProjectID:         data.ProjectID,
		ProjectCode:       data.ProjectCode,
		BusinessID:        data.BusinessID,
		ClusterID:         data.ClusterID,
		Namespace:         data.Namespace,
		WorkloadType:      data.WorkloadType,
		WorkloadName:      data.WorkloadName,
		PodAutoscalerType: data.PodAutoscalerType,
		PodAutoscalerName: data.PodAutoscalerName,
		StartTime:         start,
		EndTime:           end,
		Metrics:           nil,
		Label:             data.Label,
	}
	responseMetrics := make([]*bcsdatamanager.PodAutoscalerMetrics, 0)
	for _, metric := range metricSlice {
		responseMetric := &bcsdatamanager.PodAutoscalerMetrics{
			Time:                   metric.Time.Time().String(),
			TotalSuccessfulRescale: strconv.FormatInt(metric.TotalSuccessfulRescale, 10),
		}
		responseMetrics = append(responseMetrics, responseMetric)
	}
	response.Metrics = responseMetrics
	return response
}
