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

package project

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store/dbtable"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project/proto/bcsproject"
)

const (
	// table name
	tableName        = "project"
	projectIDField   = "projectid"
	projectNameField = "name"
	englishNameField = "englishname"
)

var (
	projectIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: map[string]int32{
				projectIDField:   1,
				englishNameField: 1,
				projectNameField: 1,
			},
			Unique: true,
		},
	}
)

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
func (m *ModelProject) CreateProject(ctx context.Context, project *proto.Project) error {
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

// GetProject get project info by projectID
func (m *ModelProject) GetProject(ctx context.Context, projectIdOrCode string) (*proto.Project, error) {
	// query project info by the `or` operation
	projectIDCond := operator.NewLeafCondition(operator.Eq, operator.M{projectIDField: projectIdOrCode})
	englishNameCond := operator.NewLeafCondition(operator.Eq, operator.M{englishNameField: projectIdOrCode})
	cond := operator.NewBranchCondition(operator.Or, projectIDCond, englishNameCond)

	retProject := &proto.Project{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, retProject); err != nil {
		return nil, err
	}
	return retProject, nil
}

// ProjectField 项目属性, 包含项目ID、英文缩写、项目名称
type ProjectField struct {
	ProjectID   string
	EnglishName string
	Name        string
}

// GetProjectByField 通过项目的属性获取项目信息
func (m *ModelProject) GetProjectByField(ctx context.Context, pf *ProjectField) (*proto.Project, error) {
	if pf.ProjectID == "" && pf.Name == "" && pf.EnglishName == "" {
		return nil, fmt.Errorf("project field: [projectID, name, englishName] cannot be empty")
	}
	projectIDCond := operator.NewLeafCondition(operator.Eq, operator.M{projectIDField: pf.ProjectID})
	englishNameCond := operator.NewLeafCondition(operator.Eq, operator.M{englishNameField: pf.EnglishName})
	nameCond := operator.NewLeafCondition(operator.Eq, operator.M{projectNameField: pf.Name})
	cond := operator.NewBranchCondition(operator.Or, projectIDCond, englishNameCond, nameCond)

	retProject := &proto.Project{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, retProject); err != nil {
		return nil, err
	}
	return retProject, nil
}

// UpdateProject update project info
func (m *ModelProject) UpdateProject(ctx context.Context, project *proto.Project) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		projectIDField: project.ProjectID,
	})
	// update project info
	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": project})
}

func (m *ModelProject) DeleteProject(ctx context.Context, projectID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		projectIDField: projectID,
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

func (m *ModelProject) ListProjects(ctx context.Context, cond *operator.Condition, page *common.Pagination) (
	[]proto.Project, int64, error) {
	projectList := make([]proto.Project, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	// total 表示根据条件得到的总量
	total, err := finder.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	if len(page.Sort) != 0 {
		finder = finder.WithSort(dbtable.MapInt2MapIf(page.Sort))
	}
	if page.Offset != 0 {
		finder = finder.WithStart(page.Offset * page.Limit)
	}
	if page.Limit == 0 {
		finder = finder.WithLimit(common.DefaultProjectLimit)
	} else {
		finder = finder.WithLimit(page.Limit)
	}

	// 设置拉取全量数据
	if page.All {
		finder = finder.WithLimit(0).WithStart(0)
	}

	// 获取数据
	if err := finder.All(ctx, &projectList); err != nil {
		return nil, 0, err
	}
	return projectList, total, nil
}
