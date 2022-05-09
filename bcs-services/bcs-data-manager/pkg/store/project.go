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
	modelProjectIndexes = []drivers.Index{
		{
			Key: bson.D{
				bson.E{Key: CreateTimeKey, Value: 1},
			},
			Name: CreateTimeKey + "_1",
		},
		{
			Name: common.ProjectTableName + "_idx",
			Key: bson.D{
				bson.E{Key: ProjectIDKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
				bson.E{Key: BucketTimeKey, Value: 1},
			},
			Unique: true,
		},
		{
			Name: common.ProjectTableName + "_get_idx",
			Key: bson.D{
				bson.E{Key: ProjectIDKey, Value: 1},
				bson.E{Key: DimensionKey, Value: 1},
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
	}
)

// ModelProject project model
type ModelProject struct {
	Public
}

// NewModelProject new project model
func NewModelProject(db drivers.DB) *ModelProject {
	return &ModelProject{Public: Public{
		TableName: common.DataTableNamePrefix + common.ProjectTableName,
		Indexes:   modelProjectIndexes,
		DB:        db,
	}}
}

// GetProjectInfo get project info data
func (m *ModelProject) GetProjectInfo(ctx context.Context,
	request *bcsdatamanager.GetProjectInfoRequest) (*bcsdatamanager.Project, error) {
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	dimension := request.Dimension
	if dimension == "" {
		dimension = common.DimensionDay
	}
	projectMetricsMap := make([]map[string]*common.ProjectMetrics, 0)
	pipeline := make([]map[string]interface{}, 0)
	pipeline = append(pipeline, map[string]interface{}{"$unwind": "$metrics"})
	pipeline = append(pipeline, map[string]interface{}{"$match": map[string]interface{}{
		ProjectIDKey: request.ProjectID,
		DimensionKey: dimension,
		MetricTimeKey: map[string]interface{}{
			"$gte": primitive.NewDateTimeFromTime(getStartTime(dimension)),
		},
	}})
	pipeline = append(pipeline, map[string]interface{}{"$project": map[string]interface{}{
		"_id":     0,
		"metrics": 1,
	}})
	err = m.DB.Table(m.TableName).Aggregation(ctx, pipeline, &projectMetricsMap)
	if err != nil {
		blog.Errorf("find project data fail, err:%v", err)
		return nil, err
	}
	if len(projectMetricsMap) == 0 {
		return &bcsdatamanager.Project{}, nil
	}
	projectMetrics := make([]*common.ProjectMetrics, 0)
	for _, metrics := range projectMetricsMap {
		projectMetrics = append(projectMetrics, metrics["metrics"])
	}
	startTime := projectMetrics[len(projectMetrics)-1].Time.Time().String()
	endTime := projectMetrics[0].Time.Time().String()
	return m.generateProjectResponse(projectMetrics, request.ProjectID, dimension, startTime, endTime), nil
}

// InsertProjectInfo insert project info data
func (m *ModelProject) InsertProjectInfo(ctx context.Context, metrics *common.ProjectMetrics,
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
		DimensionKey:  opts.Dimension,
		BucketTimeKey: bucketTime,
	})
	retProject := &common.ProjectData{}
	err = m.DB.Table(m.TableName).Find(cond).One(ctx, retProject)
	if err != nil {
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Infof("project info not found, create a new bucket")
			newMetrics := make([]*common.ProjectMetrics, 0)
			newMetrics = append(newMetrics, metrics)
			newProjectBucket := &common.ProjectData{
				CreateTime: primitive.NewDateTimeFromTime(time.Now()),
				UpdateTime: primitive.NewDateTimeFromTime(time.Now()),
				BucketTime: bucketTime,
				Dimension:  opts.Dimension,
				ProjectID:  opts.ProjectID,
				Metrics:    newMetrics,
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
	retProject.UpdateTime = primitive.NewDateTimeFromTime(time.Now())
	retProject.Metrics = append(retProject.Metrics, metrics)
	return m.DB.Table(m.TableName).
		Update(ctx, cond, operator.M{"$set": bson.M{"update_time": time.Now()}, "$push": bson.M{"metrics": metrics}})
}

// GetRawProjectInfo get raw project info data
func (m *ModelProject) GetRawProjectInfo(ctx context.Context, opts *common.JobCommonOpts,
	bucket string) ([]*common.ProjectData, error) {
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
	retProject := make([]*common.ProjectData, 0)
	err = m.DB.Table(m.TableName).Find(conds).All(ctx, &retProject)
	if err != nil {
		return nil, err
	}
	return retProject, nil
}

func (m *ModelProject) generateProjectResponse(metricSlice []*common.ProjectMetrics, projectID, dimension,
	startTime, endTime string) *bcsdatamanager.Project {
	response := &bcsdatamanager.Project{
		ProjectID: projectID,
		Dimension: dimension,
		StartTime: startTime,
		EndTime:   endTime,
		Metrics:   nil,
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

func (m *ModelProject) preAggregate(data *common.ProjectData, newMetric *common.ProjectMetrics) {
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
