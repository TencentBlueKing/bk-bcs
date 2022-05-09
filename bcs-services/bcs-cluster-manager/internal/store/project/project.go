/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/util"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	tableName = "project"
	//! we don't setting bson tag in proto file,
	//! all struct key in mongo is lowcase in default
	tableKey                 = "projectid"
	defaultProjectListLength = 1000
)

var (
	projectIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: tableKey, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelProject database operation for project
type ModelProject struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New create project model
func New(db drivers.DB) *ModelProject {
	return &ModelProject{
		tableName: util.DataTableNamePrefix + tableName,
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
	if err := util.EnsureTable(ctx, m.db, m.tableName, m.indexes); err != nil {
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
func (m *ModelProject) CreateProject(ctx context.Context, project *types.Project) error {
	if project == nil {
		return fmt.Errorf("project to be created cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	if err := util.EncryptProjectCred(project); err != nil {
		return err
	}
	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{project}); err != nil {
		return err
	}
	return nil
}

// UpdateProject update project with all fileds, if some fields are nil
// that field will be overwrite with empty
func (m *ModelProject) UpdateProject(ctx context.Context, project *types.Project) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		tableKey: project.ProjectID,
	})
	if err := util.EncryptProjectCred(project); err != nil {
		return err
	}
	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": project})
}

// DeleteProject delete project
func (m *ModelProject) DeleteProject(ctx context.Context, projectID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		tableKey: projectID,
	})
	deleteCounter, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	if deleteCounter == 0 {
		blog.Warnf("no project delete with projectID %s", projectID)
	}
	return nil
}

// GetProject get project
func (m *ModelProject) GetProject(ctx context.Context, projectID string) (*types.Project, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		tableKey: projectID,
	})
	pro := &types.Project{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, pro); err != nil {
		return nil, err
	}
	if err := util.DecryptProjectCred(pro); err != nil {
		return nil, err
	}
	return pro, nil
}

// ListProject list clusters
func (m *ModelProject) ListProject(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
	[]types.Project, error) {
	projectList := make([]types.Project, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(util.MapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit == 0 {
		finder = finder.WithLimit(defaultProjectListLength)
	} else {
		finder = finder.WithLimit(opt.Limit)
	}

	if opt.All {
		finder = finder.WithLimit(0)
	}

	if err := finder.All(ctx, &projectList); err != nil {
		return nil, err
	}
	for _, project := range projectList {
		if err := util.DecryptProjectCred(&project); err != nil {
			return nil, err
		}
	}
	return projectList, nil
}
