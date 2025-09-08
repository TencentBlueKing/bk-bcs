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

// Package nodetemplate xxx
package nodetemplate

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/util"
)

const (

	//! we don't setting bson tag in proto file,
	//! all struct key in mongo is lowcase in default

	tableName = "nodetemplate"
	// ProjectIDKey xxx
	ProjectIDKey                  = "projectid"
	templateIDKey                 = "nodetemplateid"
	NameKey                       = "name"
	defaultCloudAccountListLength = 4000
)

var (
	cloudIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: ProjectIDKey, Value: 1},
				bson.E{Key: templateIDKey, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelNodeTemplate database operation for nodeTemplate
type ModelNodeTemplate struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New create nodeTemplate model
func New(db drivers.DB) *ModelNodeTemplate {
	return &ModelNodeTemplate{
		tableName: util.DataTableNamePrefix + tableName,
		indexes:   cloudIndexes,
		db:        db,
	}
}

// ensure table
func (m *ModelNodeTemplate) ensureTable(ctx context.Context) error {
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

// CreateNodeTemplate insert nodeTemplate
func (m *ModelNodeTemplate) CreateNodeTemplate(ctx context.Context, template *types.NodeTemplate) error {
	if template == nil {
		return fmt.Errorf("nodeTemplate to be created cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{template}); err != nil {
		return err
	}
	return nil
}

// UpdateNodeTemplate update nodeTemplate with all fileds, if some fields are nil
// that field will be overwrite with empty
func (m *ModelNodeTemplate) UpdateNodeTemplate(ctx context.Context, template *types.NodeTemplate) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		ProjectIDKey:  template.ProjectID,
		templateIDKey: template.NodeTemplateID,
	})

	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": template})
}

// DeleteNodeTemplate delete nodeTemplate
func (m *ModelNodeTemplate) DeleteNodeTemplate(ctx context.Context, projectID string, templateID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		ProjectIDKey:  projectID,
		templateIDKey: templateID,
	})
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// GetNodeTemplate get nodeTemplate
func (m *ModelNodeTemplate) GetNodeTemplate(ctx context.Context, projectID, templateID string) (
	*types.NodeTemplate, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		ProjectIDKey:  projectID,
		templateIDKey: templateID,
	})

	nodeTemplate := &types.NodeTemplate{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, nodeTemplate); err != nil {
		return nil, err
	}

	return nodeTemplate, nil
}

// GetNodeTemplateByID get NodeTemplate by ID
func (m *ModelNodeTemplate) GetNodeTemplateByID(ctx context.Context, templateID string) (*types.NodeTemplate, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		templateIDKey: templateID,
	})

	nodeTemplate := &types.NodeTemplate{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, nodeTemplate); err != nil {
		return nil, err
	}

	return nodeTemplate, nil
}

// ListNodeTemplate list nodeTemplates
func (m *ModelNodeTemplate) ListNodeTemplate(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
	[]*types.NodeTemplate, error) {
	templateList := make([]*types.NodeTemplate, 0)

	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(util.MapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit == 0 {
		finder = finder.WithLimit(defaultCloudAccountListLength)
	} else {
		finder = finder.WithLimit(opt.Limit)
	}

	if err := finder.All(ctx, &templateList); err != nil {
		return nil, err
	}

	return templateList, nil
}
