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

// Package notifytemplate xxx
package notifytemplate

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

	tableName = "notifytemplate"
	// ProjectIDKey xxx
	ProjectIDKey                  = "projectid"
	templateIDKey                 = "notifytemplateid"
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

// ModelNotifyTemplate database operation for notifyTemplate
type ModelNotifyTemplate struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New create nodeTemplate model
func New(db drivers.DB) *ModelNotifyTemplate {
	return &ModelNotifyTemplate{
		tableName: util.DataTableNamePrefix + tableName,
		indexes:   cloudIndexes,
		db:        db,
	}
}

// ensure table
func (m *ModelNotifyTemplate) ensureTable(ctx context.Context) error {
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

// CreateNotifyTemplate insert notifyTemplate
func (m *ModelNotifyTemplate) CreateNotifyTemplate(ctx context.Context, template *types.NotifyTemplate) error {
	if template == nil {
		return fmt.Errorf("notifyTemplate to be created cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{template}); err != nil {
		return err
	}
	return nil
}

// UpdateNotifyTemplate update notifyTemplate with all fileds, if some fields are nil
// that field will be overwrite with empty
func (m *ModelNotifyTemplate) UpdateNotifyTemplate(ctx context.Context, template *types.NotifyTemplate) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		ProjectIDKey:  template.ProjectID,
		templateIDKey: template.NotifyTemplateID,
	})

	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": template})
}

// DeleteNotifyTemplate delete notifyTemplate
func (m *ModelNotifyTemplate) DeleteNotifyTemplate(ctx context.Context, projectID string, templateID string) error {
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

// GetNotifyTemplate get notifyTemplate
func (m *ModelNotifyTemplate) GetNotifyTemplate(ctx context.Context, projectID, templateID string) (
	*types.NotifyTemplate, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		ProjectIDKey:  projectID,
		templateIDKey: templateID,
	})

	nodeTemplate := &types.NotifyTemplate{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, nodeTemplate); err != nil {
		return nil, err
	}

	return nodeTemplate, nil
}

// GetNotifyTemplateByID get NotifyTemplate by ID
func (m *ModelNotifyTemplate) GetNotifyTemplateByID(ctx context.Context, templateID string) (
	*types.NotifyTemplate, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		templateIDKey: templateID,
	})

	nodeTemplate := &types.NotifyTemplate{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, nodeTemplate); err != nil {
		return nil, err
	}

	return nodeTemplate, nil
}

// ListNotifyTemplate list notifyTemplates
func (m *ModelNotifyTemplate) ListNotifyTemplate(ctx context.Context, cond *operator.Condition,
	opt *options.ListOption) (
	[]*types.NotifyTemplate, error) {
	templateList := make([]*types.NotifyTemplate, 0)

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
