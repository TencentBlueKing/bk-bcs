/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package project xxx
package project

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/dbtable"
)

const (
	// table name
	tableName           = "project"
	FieldKeyProjectID   = "projectID"
	FieldKeyName        = "name"
	FieldKeyProjectCode = "projectCode"
	FieldKeyKind        = "kind"
)

var (
	projectIndexes = []drivers.Index{
		{
			Name: tableName + "_projectID_idx",
			Key: bson.D{
				bson.E{Key: FieldKeyProjectID, Value: 1},
			},
			Unique: true,
		},
		{
			Name: tableName + "_projectCode_idx",
			Key: bson.D{
				bson.E{Key: FieldKeyProjectCode, Value: 1},
			},
			Unique: true,
		},
		{
			Name: tableName + "_name_idx",
			Key: bson.D{
				bson.E{Key: FieldKeyName, Value: 1},
			},
			Unique: false,
		},
		{
			Name: tableName + "_kind_idx",
			Key: bson.D{
				bson.E{Key: FieldKeyKind, Value: 1},
			},
			Unique: false,
		},
	}
)

// Project xxx
type Project struct {
	CreateTime  string `json:"createTime" bson:"createTime"`
	UpdateTime  string `json:"updateTime" bson:"updateTime"`
	Creator     string `json:"creator" bson:"creator"`
	Updater     string `json:"updater" bson:"updater"`
	Managers    string `json:"managers" bson:"managers"`
	ProjectID   string `json:"projectID" bson:"projectID"`
	Name        string `json:"name" bson:"name"`
	ProjectCode string `json:"projectCode" bson:"projectCode"`
	UseBKRes    bool   `json:"useBKRes" bson:"useBKRes"`
	Description string `json:"description" bson:"description"`
	IsOffline   bool   `json:"isOffline" bson:"isOffline"`
	Kind        string `json:"kind" bson:"kind"`
	BusinessID  string `json:"businessID" bson:"businessID"`
	IsSecret    bool   `json:"isSecret" bson:"isSecret"`
	ProjectType uint32 `json:"projectType" bson:"projectType"`
	DeployType  uint32 `json:"deployType" bson:"deployType"`
	BGID        string `json:"bgID" bson:"bgID"`
	BGName      string `json:"bgName" bson:"bgName"`
	DeptID      string `json:"deptID" bson:"deptID"`
	DeptName    string `json:"deptName" bson:"deptName"`
	CenterID    string `json:"centerID" bson:"centerID"`
	CenterName  string `json:"centerName" bson:"centerName"`
}

// ModelProject provide project db
type ModelProject struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New return a new project model instance
func New(db drivers.DB) *ModelProject {
	return &ModelProject{
		tableName: dbtable.DataTableNamePrefix + tableName,
		indexes:   projectIndexes,
		db:        db,
	}
}

// ensureTable xxx
// ensure table
func (m *ModelProject) ensureTable(ctx context.Context) error {
	m.isTableEnsuredMutex.RLock()
	if m.isTableEnsured {
		m.isTableEnsuredMutex.RUnlock()
		return nil
	}
	if err := dbtable.EnsureTable(ctx, m.db, m.tableName, m.indexes); err != nil {
		m.isTableEnsuredMutex.RUnlock()
		return err
	}
	m.isTableEnsuredMutex.RUnlock()

	m.isTableEnsuredMutex.Lock()
	m.isTableEnsured = true
	m.isTableEnsuredMutex.Unlock()
	return nil
}

// CreateProject create project
func (m *ModelProject) CreateProject(ctx context.Context, project *Project) error {
	if project == nil {
		return fmt.Errorf("project cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{project}); err != nil {
		return err
	}
	return nil
}

// GetProject get project info by projectID or projectCode
func (m *ModelProject) GetProject(ctx context.Context, projectIDOrCode string) (*Project, error) {
	// query project info by the `or` operation
	projectIDCond := operator.NewLeafCondition(operator.Eq, operator.M{FieldKeyProjectID: projectIDOrCode})
	projectCodeCond := operator.NewLeafCondition(operator.Eq, operator.M{FieldKeyProjectCode: projectIDOrCode})
	cond := operator.NewBranchCondition(operator.Or, projectIDCond, projectCodeCond)

	retProject := &Project{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, retProject); err != nil {
		return nil, err
	}
	return retProject, nil
}

// ProjectField 项目属性, 包含项目ID、英文缩写、项目名称
type ProjectField struct {
	ProjectID   string
	ProjectCode string
	Name        string
}

// GetProjectByField 通过项目的属性获取项目信息
func (m *ModelProject) GetProjectByField(ctx context.Context, pf *ProjectField) (*Project, error) {
	if pf.ProjectID == "" && pf.Name == "" && pf.ProjectCode == "" {
		return nil, fmt.Errorf("project field: [projectID, name, projectCode] cannot be empty")
	}
	projectIDCond := operator.NewLeafCondition(operator.Eq, operator.M{FieldKeyProjectID: pf.ProjectID})
	projectCodeCond := operator.NewLeafCondition(operator.Eq, operator.M{FieldKeyProjectCode: pf.ProjectCode})
	nameCond := operator.NewLeafCondition(operator.Eq, operator.M{FieldKeyName: pf.Name})
	cond := operator.NewBranchCondition(operator.Or, projectIDCond, projectCodeCond, nameCond)

	retProject := &Project{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, retProject); err != nil {
		return nil, err
	}
	return retProject, nil
}

// UpdateProject update project info
func (m *ModelProject) UpdateProject(ctx context.Context, project *Project) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		FieldKeyProjectID: project.ProjectID,
	})
	// update project info
	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": project})
}

// DeleteProject delete project record
func (m *ModelProject) DeleteProject(ctx context.Context, projectID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		FieldKeyProjectID: projectID,
	})
	deleteCounter, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	if deleteCounter == 0 {
		logging.Warn("the projectID %s of project not found", projectID)
	}
	return nil
}

// ListProjects query project list
func (m *ModelProject) ListProjects(ctx context.Context, cond *operator.Condition, pagination *page.Pagination) (
	[]Project, int64, error) {
	projectList := make([]Project, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	// total 表示根据条件得到的总量
	total, err := finder.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	if len(pagination.Sort) != 0 {
		finder = finder.WithSort(dbtable.MapInt2MapIf(pagination.Sort))
	}
	if pagination.Offset != 0 {
		finder = finder.WithStart(pagination.Offset * pagination.Limit)
	}
	if pagination.Limit == 0 {
		finder = finder.WithLimit(page.DefaultPageLimit)
	} else {
		finder = finder.WithLimit(pagination.Limit)
	}

	// 设置拉取全量数据
	if pagination.All {
		finder = finder.WithLimit(0).WithStart(0)
	}

	// 获取数据
	if err := finder.All(ctx, &projectList); err != nil {
		return nil, 0, err
	}

	return projectList, total, nil
}

// SearchProjects query project sort by ids
// NOCC:golint/fnsize(设计如此:该方法较长且不可拆分)
func (m *ModelProject) SearchProjects(ctx context.Context, ids []string, searchKey, kind string,
	pagination *page.Pagination) ([]Project, int64, error) {
	if pagination.Limit == 0 {
		pagination.Limit = page.DefaultPageLimit
	}
	projectList := make([]Project, 0)
	matchElement := bson.D{}
	if kind != "" {
		matchElement = append(matchElement, bson.E{Key: "kind", Value: kind})
	}
	if searchKey != "" {
		matchElement = append(matchElement, bson.E{Key: "$or", Value: bson.A{
			bson.D{{"name", bson.D{{"$regex", searchKey}, {"$options", "i"}}}},
			bson.D{{"projectCode", bson.D{{"$regex", searchKey}, {"$options", "i"}}}},
			bson.D{{"projectID", fmt.Sprintf("\"%s\"", searchKey)}},
			bson.D{{"businessID", fmt.Sprintf("\"%s\"", searchKey)}},
		}})
	}
	matchPipline := bson.D{{"$match", matchElement}}
	pipeline := mongo.Pipeline{}
	countPipeline := mongo.Pipeline{}
	if kind != "" || searchKey != "" {
		pipeline = append(pipeline, matchPipline)
		countPipeline = append(countPipeline, matchPipline)
	}
	pipeline = append(pipeline, bson.D{
		{"$addFields", bson.D{
			{"authorized", bson.D{
				{"$cond", bson.D{
					{"if", bson.D{{"$in", bson.A{"$projectID", ids}}}},
					{"then", 1},
					{"else", 0},
				}},
			}},
			// kind 为 k8s 则认为开启了容器服务，排在前面
			{"enabled", bson.D{
				{"$cond", bson.D{
					{"if", bson.D{{"$eq", bson.A{"$kind", "k8s"}}}},
					{"then", 1},
					{"else", 0},
				}},
			}},
		}},
	})
	pipeline = append(pipeline,
		bson.D{{"$project", bson.D{
			{"createTime", 1},
			{"updateTime", 1},
			{"creator", 1},
			{"updater", 1},
			{"managers", 1},
			{"projectID", 1},
			{"name", 1},
			{"projectCode", 1},
			{"useBKRes", 1},
			{"description", 1},
			{"isOffline", 1},
			{"kind", 1},
			{"businessID", 1},
			{"isSecret", 1},
			{"projectType", 1},
			{"deployType", 1},
			{"bgID", 1},
			{"bgName", 1},
			{"deptID", 1},
			{"deptName", 1},
			{"centerID", 1},
			{"centerName", 1},
			{"authorized", 1},
			{"enabled", 1}}}},
		bson.D{{"$sort", bson.D{
			{"authorized", -1},
			{"enabled", -1},
			{"name", 1},
		}}},
		bson.D{{"$skip", pagination.Offset}},
		bson.D{{"$limit", pagination.Limit}})
	if err := m.db.Table(m.tableName).Aggregation(ctx, pipeline, &projectList); err != nil {
		return nil, 0, err
	}
	var counts []struct {
		Count int64 `bson:"count"`
	}
	countPipeline = append(countPipeline, bson.D{
		{"$group", bson.D{
			{"_id", nil},
			{"count", bson.D{{"$sum", 1}}},
		}}})
	if err := m.db.Table(m.tableName).Aggregation(ctx, countPipeline, &counts); err != nil {
		return nil, 0, err
	}
	var count int64
	if len(counts) < 1 {
		count = 0
	} else {
		count = counts[0].Count
	}
	return projectList, count, nil
}

// ListProjectByIDs query project by project ids
func (m *ModelProject) ListProjectByIDs(ctx context.Context, kind string, ids []string, pagination *page.Pagination) (
	[]Project, int64, error) {
	projectList := make([]Project, 0)
	condKind := make(operator.M)
	condID := make(operator.M)
	if kind != "" {
		condKind["kind"] = kind
	}
	condID["projectID"] = ids
	cond := operator.NewBranchCondition(operator.And,
		operator.NewLeafCondition(operator.In, condID), operator.NewLeafCondition(operator.Eq, condKind))
	finder := m.db.Table(m.tableName).Find(cond)
	// 获取总量
	total, err := finder.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	// 拉取满足项目 ID 的全量数据
	if err := finder.All(ctx, &projectList); err != nil {
		return nil, 0, err
	}
	return projectList, total, nil
}
