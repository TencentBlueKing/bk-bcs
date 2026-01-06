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

// Package variabledefinition xxx
package variabledefinition

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/dbtable"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	timeutil "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/time"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

const (
	// table name
	tableName = "variable_definition"
	// FieldKeyID id
	FieldKeyID = "_id"
	// FieldKeyKey key
	FieldKeyKey = "key"
	// FieldKeyDefault default
	FieldKeyDefault = "default"
	// FieldKeyName name
	FieldKeyName = "name"
	// FieldKeyDescription description
	FieldKeyDescription = "description"
	// FieldKeyProjectCode projectCode
	FieldKeyProjectCode = "projectCode"
	// FieldKeyScope scope
	FieldKeyScope = "scope"
	// FieldKeyCategory category
	FieldKeyCategory = "category"
	// FieldKeyCreateTime createTime
	FieldKeyCreateTime = "createTime"
	// FieldKeyUpdateTime updateTime
	FieldKeyUpdateTime = "updateTime"
	// FieldKeyCreator creator
	FieldKeyCreator = "creator"
	// FieldKeyUpdater updater
	FieldKeyUpdater = "updater"
	// FieldKeyIsDeleted isDeleted
	FieldKeyIsDeleted = "isDeleted"
	// FieldKeyDeleteTime deleteTime
	FieldKeyDeleteTime = "deleteTime"
)

var (
	variableDefinitionIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: FieldKeyKey, Value: 1},
				bson.E{Key: FieldKeyProjectCode, Value: 1},
				bson.E{Key: FieldKeyIsDeleted, Value: 1},
			},
			Unique: false,
		},
	}
)

var (
	// VariableCategorySys sys
	VariableCategorySys = "sys"
	// VariableCategoryCustom custom
	VariableCategoryCustom = "custom"
	// VariableScopeGlobal global scope
	VariableScopeGlobal = "global"
	// VariableScopeCluster cluster scope
	VariableScopeCluster = "cluster"
	// VariableScopeNamespace namespace scope
	VariableScopeNamespace = "namespace"
)

// GetScopeName get scope name by scope str
func GetScopeName(scope string) string {
	switch scope {
	case VariableScopeGlobal:
		return "全局变量"
	case VariableScopeCluster:
		return "集群变量"
	case VariableScopeNamespace:
		return "命名空间变量"
	default:
		return "非法作用范围"
	}

}

// GetCategoryName get category name by category str
func GetCategoryName(category string) string {
	switch category {
	case VariableCategorySys:
		return "系统内置"
	case VariableCategoryCustom:
		return "自定义"
	default:
		return "非法类型"
	}
}

// SystemVariables system buildin variables
var SystemVariables = map[string]*VariableDefinition{
	"SYS_NON_STANDARD_DATA_ID": {
		ID:       "variable-sys-non-standard-data-id",
		Key:      "SYS_NON_STANDARD_DATA_ID",
		Name:     "非标准日志采集DataId",
		Category: VariableCategorySys,
		Scope:    VariableScopeGlobal,
	},
	"SYS_STANDARD_DATA_ID": {
		ID:       "variable-sys-standard-data-id",
		Key:      "SYS_STANDARD_DATA_ID",
		Name:     "标准日志采集DataId",
		Category: VariableCategorySys,
		Scope:    VariableScopeGlobal,
	},
	"SYS_NAMESPACE": {
		ID:       "variable-sys-namespace",
		Key:      "SYS_NAMESPACE",
		Name:     "命名空间",
		Category: VariableCategorySys,
		Scope:    VariableScopeNamespace,
	},
	"SYS_JFROG_DOMAIN": {
		ID:       "variable-sys-jfrog-domain",
		Key:      "SYS_JFROG_DOMAIN",
		Name:     "仓库域名",
		Category: VariableCategorySys,
		Scope:    VariableScopeCluster,
	},
	"SYS_CLUSTER_ID": {
		ID:       "variable-sys-cluster-id",
		Key:      "SYS_CLUSTER_ID",
		Name:     "集群ID",
		Category: VariableCategorySys,
		Scope:    VariableScopeCluster,
	},
	"SYS_CC_APP_ID": {
		ID:       "variable-sys-cc-app-id",
		Key:      "SYS_CC_APP_ID",
		Name:     "业务ID",
		Category: VariableCategorySys,
		Scope:    VariableScopeGlobal,
	},
	"SYS_PROJECT_ID": {
		ID:       "variable-sys-project-id",
		Key:      "SYS_PROJECT_ID",
		Name:     "项目ID",
		Category: VariableCategorySys,
		Scope:    VariableScopeGlobal,
	},
}

// FilterSystemVariables filter system variables
func FilterSystemVariables(scope []string, searchKey string) []*VariableDefinition {
	variables := []*VariableDefinition{}
	for _, v := range SystemVariables {
		if !stringx.StringInSlice(v.Scope, scope) {
			continue
		}
		if searchKey != "" && !strings.Contains(strings.ToLower(v.Key), strings.ToLower(searchKey)) {
			continue
		}
		variables = append(variables, v)
	}
	return variables
}

// VariableDefinition ...
type VariableDefinition struct {
	ID          string `json:"id" bson:"_id"`
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
	DeleteTime  string `json:"deleteTime" bson:"deleteTime"`
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
		Id:           m.ID,
		Key:          m.Key,
		Name:         m.Name,
		Default:      m.Default,
		DefaultValue: m.Default,
		Scope:        m.Scope,
		ScopeName:    GetScopeName(m.Scope),
		Category:     m.Category,
		CategoryName: GetCategoryName(m.Category),
		Desc:         m.Description,
		Created:      m.CreateTime,
		Updated:      m.UpdateTime,
		Creator:      m.Creator,
		Updater:      m.Updater,
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
	if vd.CreateTime == "" {
		vd.CreateTime = time.Now().UTC().Format(time.RFC3339)
	}
	if vd.UpdateTime == "" {
		vd.UpdateTime = time.Now().UTC().Format(time.RFC3339)
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
	condM[FieldKeyID] = variableID
	condM[FieldKeyIsDeleted] = false
	cond := operator.NewLeafCondition(operator.Eq, condM)

	retVariableDefinition := &VariableDefinition{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, retVariableDefinition); err != nil {
		return nil, err
	}
	retVariableDefinition.CreateTime = timeutil.TransStrToUTCStr(time.RFC3339Nano, retVariableDefinition.CreateTime)
	retVariableDefinition.UpdateTime = timeutil.TransStrToUTCStr(time.RFC3339Nano, retVariableDefinition.UpdateTime)
	return retVariableDefinition, nil
}

// GetVariableDefinitionByKey get variable definition info by key and project code
func (m *ModelVariableDefinition) GetVariableDefinitionByKey(ctx context.Context,
	projectCode, key string) (*VariableDefinition, error) {
	// query variable definition info by the `and` operation
	condM := make(operator.M)
	condM[FieldKeyProjectCode] = projectCode
	condM[FieldKeyKey] = key
	condM[FieldKeyIsDeleted] = false
	cond := operator.NewLeafCondition(operator.Eq, condM)

	retVariableDefinition := &VariableDefinition{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, retVariableDefinition); err != nil {
		return nil, err
	}
	retVariableDefinition.CreateTime = timeutil.TransStrToUTCStr(time.RFC3339Nano, retVariableDefinition.CreateTime)
	retVariableDefinition.UpdateTime = timeutil.TransStrToUTCStr(time.RFC3339Nano, retVariableDefinition.UpdateTime)
	return retVariableDefinition, nil
}

// UpdateVariableDefinition update variable definition info
func (m *ModelVariableDefinition) UpdateVariableDefinition(
	ctx context.Context, vd entity.M) (*VariableDefinition, error) {
	if vd == nil {
		return nil, fmt.Errorf("can not update empty variable definition")
	}

	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		FieldKeyID:        vd.GetString(FieldKeyID),
		FieldKeyIsDeleted: false,
	})
	// update time
	vd[FieldKeyUpdateTime] = time.Now().UTC().Format(time.RFC3339)

	if err := m.db.Table(m.tableName).Update(ctx, cond, operator.M{"$set": vd}); err != nil {
		return nil, err
	}
	variableDefinition := &VariableDefinition{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, variableDefinition); err != nil {
		return nil, err
	}
	return variableDefinition, nil
}

// DeleteVariableDefinitions batch delete variable definition records
func (m *ModelVariableDefinition) DeleteVariableDefinitions(ctx context.Context, ids []string) (int64, error) {
	if err := m.ensureTable(ctx); err != nil {
		return 0, err
	}
	// delete variable definition info
	cond := operator.NewLeafCondition(operator.In, operator.M{
		FieldKeyID: ids,
	})
	cond = operator.NewBranchCondition(operator.And, cond,
		operator.NewLeafCondition(operator.Eq, operator.M{FieldKeyIsDeleted: false}))
	return m.db.Table(m.tableName).UpdateMany(ctx, cond, operator.M{"$set": operator.M{
		FieldKeyIsDeleted:  true,
		FieldKeyDeleteTime: time.Now().UTC().Format(time.RFC3339),
	}})
}

// ListVariableDefinitions query variable definition list
func (m *ModelVariableDefinition) ListVariableDefinitions(ctx context.Context,
	cond *operator.Condition, pagination *page.Pagination) (
	[]VariableDefinition, int64, error) {
	vdList := make([]VariableDefinition, 0)
	cond = operator.NewBranchCondition(operator.And, cond,
		operator.NewLeafCondition(operator.Eq, operator.M{FieldKeyIsDeleted: false}))
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
		finder = finder.WithStart(pagination.Offset)
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

// UpsertVariableDefinition upsert variable definition
func (m *ModelVariableDefinition) UpsertVariableDefinition(ctx context.Context,
	entity *VariableDefinition) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	condM := make(operator.M)
	condM[FieldKeyProjectCode] = entity.ProjectCode
	condM[FieldKeyKey] = entity.Key
	condM[FieldKeyIsDeleted] = false
	cond := operator.NewLeafCondition(operator.Eq, condM)
	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": entity})
}
