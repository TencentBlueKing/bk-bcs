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

package variabledefinition

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/dbtable"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	// table name
	tableName        = "variable_definition"
	idField          = "_id"
	projectCodeField = "projectCode"
	keyField         = "key"
)

var (
	variableDefinitionIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: projectCodeField, Value: 1},
				bson.E{Key: keyField, Value: 1},
			},
			Unique: true,
		},
	}
)

var (
	// VariableDefinitonCategorySys xxx
	VariableDefinitonCategorySys = "sys"
	// VariableDefinitionCategoryCustom xxx
	VariableDefinitionCategoryCustom = "custom"
	// VariableDefinitionScopeGlobal xxx
	VariableDefinitionScopeGlobal = "global"
	// VariableDefinitionScopeCluster xxx
	VariableDefinitionScopeCluster = "cluster"
	// VariableDefinitionScopeNamespace xxx
	VariableDefinitionScopeNamespace = "namespace"

	// VariableIdPrefix xxx
	VariableIdPrefix = "variable-"
)

// VariableDefinition xxx
type VariableDefinition struct {
	ID          string `json:"id" bson:"_id"`
	VariableID  string `json:"variableID" bson:"variableID"`
	Key         string `json:"key" bson:"key"`
	Default     string `json:"default" bson:"default"`
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
	ProjectCode string `json:"projectCode" bson:"projectCode"`
	Scope       string `json:"scope" bson:"scope"`       // global, cluster, namespace
	Category    string `json:"category" bson:"category"` // sys, custom
	CreateTime  string `json:"createTime" bson:"createTime"`
	UpdateTime  string `json:"updateTime" bson:"updateTime"`
	Creator     string `json:"creator" bson:"creator"`
	Updater     string `json:"updater" bson:"updater"`
	IsDeleted   bool   `json:"isDeleted" bson:"isDeleted"`
}

// ModelVariableDefinition provide variable definition db
type ModelVariableDefinition struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New return a new variable definition model instance
func New(db drivers.DB) *ModelVariableDefinition {
	return &ModelVariableDefinition{
		tableName: dbtable.DataTableNamePrefix + tableName,
		indexes:   variableDefinitionIndexes,
		db:        db,
	}
}

// ensureTable xxx
// ensure table
func (m *ModelVariableDefinition) ensureTable(ctx context.Context) error {
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

// Transfer2Proto transfer variable definition to proto
func (m *VariableDefinition) Transfer2Proto() *proto.VariableDefinition {
	return &proto.VariableDefinition{
		Id:   m.ID,
		Key:  m.Key,
		Name: m.Name,
		// TODO: Default
		DefaultValue: m.Default,
		Scope:        m.Scope,
		// TODO: ScopeName
		Category: m.Category,
		// TODO: CategoryName
		Desc:    m.Description,
		Created: m.CreateTime,
		Updated: m.UpdateTime,
		Creator: m.Creator,
		Updater: m.Updater,
	}
}

// CreateVariableDefinition create variable definition
func (m *ModelVariableDefinition) CreateVariableDefinition(ctx context.Context, vd *VariableDefinition) error {
	if vd == nil {
		return fmt.Errorf("variable definition cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{vd}); err != nil {
		return err
	}
	return nil
}

// GetVariableDefinition get variable definition info by key and project code
func (m *ModelVariableDefinition) GetVariableDefinition(ctx context.Context,
	variableID string) (*VariableDefinition, error) {
	// query variable definition info by the `and` operation
	condM := make(operator.M)
	condM[idField] = variableID
	cond := operator.NewLeafCondition(operator.Eq, condM)

	retVariableDefinition := &VariableDefinition{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, retVariableDefinition); err != nil {
		return nil, err
	}
	return retVariableDefinition, nil
}

// GetVariableDefinitionByKey get variable definition info by key and project code
func (m *ModelVariableDefinition) GetVariableDefinitionByKey(ctx context.Context,
	projectCode, key string) (*VariableDefinition, error) {
	// query variable definition info by the `and` operation
	condM := make(operator.M)
	condM[projectCodeField] = projectCode
	condM[keyField] = key
	cond := operator.NewLeafCondition(operator.Eq, condM)

	retVariableDefinition := &VariableDefinition{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, retVariableDefinition); err != nil {
		return nil, err
	}
	return retVariableDefinition, nil
}

// UpdateVariableDefinition update variable definition info
func (m *ModelVariableDefinition) UpdateVariableDefinition(ctx context.Context, vd *VariableDefinition) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	condM := make(operator.M)
	condM[idField] = vd.ID
	cond := operator.NewLeafCondition(operator.Eq, condM)
	// update variable definition info
	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": vd})
}

// DeleteVariableDefinition delete variable definition record
func (m *ModelVariableDefinition) DeleteVariableDefinition(ctx context.Context, key, projectCode string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	condKey := operator.NewLeafCondition(operator.Eq, operator.M{
		keyField: key,
	})
	condProjectCode := operator.NewLeafCondition(operator.Eq, operator.M{
		projectCodeField: projectCode,
	})
	cond := operator.NewBranchCondition(operator.And, condKey, condProjectCode)
	// delete variable definition info
	deleteCounter, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	if deleteCounter == 0 {
		logging.Warn("the variable key %s not found in project %s", key, projectCode)
	}
	return nil
}

// ListVariableDefinitions query variable definition list
func (m *ModelVariableDefinition) ListVariableDefinitions(ctx context.Context,
	cond *operator.Condition, pagination *page.Pagination) (
	[]VariableDefinition, int64, error) {
	vdList := make([]VariableDefinition, 0)
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
	if err := finder.All(ctx, &vdList); err != nil {
		return nil, 0, err
	}

	return vdList, total, nil
}
