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
	modelProjectIndexes = []drivers.Index{
		{
			Key: bson.D{
				bson.E{Key: CreateTimeKey, Value: 1},
			},
			Name: CreateTimeKey + "_1",
		},
		{
			Name: types.ProjectTableName + "_idx",
			Key: bson.D{
				bson.E{Key: ProjectIDKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: BucketTimeKey, Value: 1},
			},
			Unique: true,
		},
		{
			Name: types.ProjectTableName + "_get_idx",
			Key: bson.D{
				bson.E{Key: ProjectIDKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: MetricTimeKey, Value: 1},
			},
		},
		{
			Name: types.ProjectTableName + "_get_idx2",
			Key: bson.D{
				bson.E{Key: BusinessIDKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
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
			Name:       BusinessIDKey + "_1",
			Background: true,
		},
	}
)

// ModelProject project model
type ModelProject struct {
	Public
}

// NewModelProject new project model
func NewModelProject(db drivers.DB) *ModelProject {
	return &ModelProject{Public: Public{
		TableName: types.DataTableNamePrefix + types.ProjectTableName,
		Indexes:   modelProjectIndexes,
		DB:        db,
	}}
}

// GetProjectList get project list, if startTime or endTime is empty, return metrics with default time range
func (m *ModelProject) GetProjectList(ctx context.Context,
	req *bcsdatamanager.GetAllProjectListRequest) ([]*bcsdatamanager.Project, int64, error) {
	err := ensureTable(ctx, &m.Public)
	var total int64
	if err != nil {
		return nil, total, err
	}
	dimension := req.Dimension
	if dimension == "" {
		dimension = types.DimensionDay
	}
	cond := make([]*operator.Condition, 0)
	cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
		DimensionKey: dimension,
	}))
	startTime := getStartTime(dimension)
	if req.GetStartTime() != 0 {
		startTime = time.Unix(req.GetStartTime(), 0)
	}
	cond = append(cond, operator.NewLeafCondition(operator.Gte, operator.M{
		MetricTimeKey: primitive.NewDateTimeFromTime(startTime),
	}))
	if req.GetEndTime() != 0 {
		cond = append(cond, operator.NewLeafCondition(operator.Lte, operator.M{
			MetricTimeKey: primitive.NewDateTimeFromTime(time.Unix(req.GetEndTime(), 0)),
		}))
	}
	conds := operator.NewBranchCondition(operator.And, cond...)
	tempProjectList := make([]map[string]string, 0)
	err = m.DB.Table(m.TableName).Find(conds).WithProjection(map[string]int{ProjectIDKey: 1, "_id": 0}).
		WithSort(map[string]interface{}{ProjectIDKey: 1}).All(ctx, &tempProjectList)
	if err != nil {
		blog.Errorf("get project id list error")
		return nil, total, err
	}
	projectList := distinctSlice("project_id", &tempProjectList)
	if len(projectList) == 0 {
		return nil, total, nil
	}
	total = int64(len(projectList))
	page := int(req.Page)
	size := int(req.Size)
	if size == 0 {
		size = DefaultSize
	}
	endIndex := (page + 1) * size
	startIndex := page * size
	if startIndex >= len(projectList) {
		return nil, total, nil
	}
	if endIndex >= len(projectList) {
		endIndex = len(projectList)
	}
	chooseProject := projectList[startIndex:endIndex]
	response := make([]*bcsdatamanager.Project, 0)
	for _, project := range chooseProject {
		projectRequest := &bcsdatamanager.GetProjectInfoRequest{
			Project:   project,
			Dimension: dimension,
			StartTime: req.GetStartTime(),
			EndTime:   req.GetEndTime(),
		}
		projectInfo, err := m.GetProjectInfo(ctx, projectRequest)
		if err != nil {
			blog.Errorf("get project[%s] info err:%v", project, err)
		} else {
			response = append(response, projectInfo)
		}
	}
	return response, total, nil
}

// GetProjectInfo get project info data, if startTime or endTime is empty, return metrics with default time range
func (m *ModelProject) GetProjectInfo(ctx context.Context,
	request *bcsdatamanager.GetProjectInfoRequest) (*bcsdatamanager.Project, error) {
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	dimension := request.Dimension
	if dimension == "" {
		dimension = types.DimensionDay
	}
	metricStartTime := getStartTime(dimension)
	if request.GetStartTime() != 0 {
		metricStartTime = time.Unix(request.GetStartTime(), 0)
	}
	metricEndTime := time.Now()
	if request.GetEndTime() != 0 {
		metricEndTime = time.Unix(request.GetEndTime(), 0)
	}
	projectMetricsMap := make([]*types.ProjectData, 0)
	pipeline := make([]map[string]interface{}, 0)
	if request.Project != "" {
		pipeline = append(pipeline, map[string]interface{}{"$match": map[string]interface{}{
			ProjectIDKey: request.Project,
			DimensionKey: dimension,
			MetricTimeKey: map[string]interface{}{
				"$gte": primitive.NewDateTimeFromTime(metricStartTime),
				"$lte": primitive.NewDateTimeFromTime(metricEndTime),
			},
		}})
	} else if request.Business != "" {
		pipeline = append(pipeline, map[string]interface{}{"$match": map[string]interface{}{
			BusinessIDKey: request.Business,
			DimensionKey:  dimension,
			MetricTimeKey: map[string]interface{}{
				"$gte": primitive.NewDateTimeFromTime(metricStartTime),
				"$lte": primitive.NewDateTimeFromTime(metricEndTime),
			},
		}})
	}
	pipeline = append(pipeline, map[string]interface{}{"$unwind": "$metrics"},
		map[string]interface{}{"$match": map[string]interface{}{
			MetricTimeKey: map[string]interface{}{
				"$gte": primitive.NewDateTimeFromTime(metricStartTime),
			},
		}}, map[string]interface{}{"$project": map[string]interface{}{
			"_id":          0,
			"project_id":   1,
			"project_code": 1,
			"business_id":  1,
			"metrics":      1,
			"label":        1,
		}}, map[string]interface{}{"$group": map[string]interface{}{
			"_id":          nil,
			"project_id":   map[string]interface{}{"$first": "$project_id"},
			"project_code": map[string]interface{}{"$max": "$project_code"},
			"business_id":  map[string]interface{}{"$max": "$business_id"},
			"metrics":      map[string]interface{}{"$push": "$metrics"},
			"label":        map[string]interface{}{"$max": "$label"},
		}},
	)
	pipeline = append(pipeline) // nolint no-op append call
	err = m.DB.Table(m.TableName).Aggregation(ctx, pipeline, &projectMetricsMap)
	if err != nil {
		blog.Errorf("find project data fail, err:%v", err)
		return nil, err
	}
	if len(projectMetricsMap) == 0 {
		return &bcsdatamanager.Project{}, nil
	}
	projectMetrics := make([]*types.ProjectMetrics, 0)
	for _, metrics := range projectMetricsMap {
		projectMetrics = append(projectMetrics, metrics.Metrics...)
	}
	endTime := projectMetrics[len(projectMetrics)-1].Time.Time().String()
	startTime := projectMetrics[0].Time.Time().String()
	return m.generateProjectResponse(projectMetrics, projectMetricsMap[0], startTime, endTime), nil
}

// InsertProjectInfo insert project info data, pre aggregate before insert
func (m *ModelProject) InsertProjectInfo(ctx context.Context, metrics *types.ProjectMetrics,
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
		DimensionKey:  opts.Dimension,
		BucketTimeKey: bucketTime,
	})
	retProject := &types.ProjectData{}
	err = m.DB.Table(m.TableName).Find(cond).One(ctx, retProject)
	if err != nil {
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Infof("project info not found, create a new bucket")
			newMetrics := make([]*types.ProjectMetrics, 0)
			newMetrics = append(newMetrics, metrics)
			newProjectBucket := &types.ProjectData{
				CreateTime:  primitive.NewDateTimeFromTime(time.Now()),
				UpdateTime:  primitive.NewDateTimeFromTime(time.Now()),
				BucketTime:  bucketTime,
				Dimension:   opts.Dimension,
				ProjectID:   opts.ProjectID,
				ProjectCode: opts.ProjectCode,
				BusinessID:  opts.BusinessID,
				Metrics:     newMetrics,
				Label:       opts.Label,
			}
			m.preAggregate(newProjectBucket, metrics)
			_, err = m.DB.Table(m.TableName).Insert(ctx, []interface{}{newProjectBucket})
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	m.preAggregate(retProject, metrics)
	if retProject.BusinessID == "" {
		retProject.BusinessID = opts.BusinessID
	}
	retProject.Label = opts.Label
	retProject.ProjectCode = opts.ProjectCode
	retProject.UpdateTime = primitive.NewDateTimeFromTime(time.Now())
	retProject.Metrics = append(retProject.Metrics, metrics)
	return m.DB.Table(m.TableName).
		Update(ctx, cond, operator.M{"$set": retProject})
}

// GetRawProjectInfo get raw project info data，如果不指定bucket返回全部
func (m *ModelProject) GetRawProjectInfo(ctx context.Context, opts *types.JobCommonOpts,
	bucket string) ([]*types.ProjectData, error) {
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
	if bucket != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			BucketTimeKey: bucket,
		}))
	}
	conds := operator.NewBranchCondition(operator.And, cond...)
	retProject := make([]*types.ProjectData, 0)
	err = m.DB.Table(m.TableName).Find(conds).All(ctx, &retProject)
	if err != nil {
		return nil, err
	}
	return retProject, nil
}

// generateProjectResponse 构造response，将storage结构转化为proto结构
func (m *ModelProject) generateProjectResponse(metricSlice []*types.ProjectMetrics,
	data *types.ProjectData, startTime, endTime string) *bcsdatamanager.Project {
	response := &bcsdatamanager.Project{
		ProjectID:   data.ProjectID,
		ProjectCode: data.ProjectCode,
		BusinessID:  data.BusinessID,
		StartTime:   startTime,
		EndTime:     endTime,
		Metrics:     nil,
		Label:       data.Label,
	}
	responseMetrics := make([]*bcsdatamanager.ProjectMetrics, 0)
	for _, metric := range metricSlice {
		responseMetric := &bcsdatamanager.ProjectMetrics{
			Time:               metric.Time.Time().String(),
			ClustersCount:      strconv.FormatInt(metric.ClustersCount, 10),
			TotalCPU:           strconv.FormatFloat(metric.TotalCPU, 'f', 2, 64),
			TotalMemory:        strconv.FormatInt(metric.TotalMemory, 10),
			TotalLoadCPU:       strconv.FormatFloat(metric.TotalLoadCPU, 'f', 2, 64),
			TotalLoadMemory:    strconv.FormatInt(metric.TotalLoadMemory, 10),
			AvgLoadCPU:         strconv.FormatFloat(metric.AvgLoadCPU, 'f', 2, 64),
			AvgLoadMemory:      strconv.FormatInt(metric.AvgLoadMemory, 10),
			CPUUsage:           strconv.FormatFloat(metric.CPUUsage, 'f', 4, 64),
			MemoryUsage:        strconv.FormatFloat(metric.MemoryUsage, 'f', 4, 64),
			NodeCount:          strconv.FormatInt(metric.NodeCount, 10),
			AvailableNodeCount: strconv.FormatInt(metric.AvailableNodeCount, 10),
			// MinNodeCount:       int64(metric.MinNode.Value),
			// MinNodeTime:        metric.MinNode.Period,
			// MaxNodeCount:       int64(metric.MaxNode.Value),
			// MaxNodeTime:        metric.MaxNode.Period,
		}
		responseMetrics = append(responseMetrics, responseMetric)
	}
	response.Metrics = responseMetrics
	return response
}

// preAggregate 预聚合project统计类数值
func (m *ModelProject) preAggregate(data *types.ProjectData, newMetric *types.ProjectMetrics) {
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
