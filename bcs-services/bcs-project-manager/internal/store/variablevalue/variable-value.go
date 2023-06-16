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

package variablevalue

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/dbtable"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
)

const (
	// table name
	tableName          = "variable_value"
	FieldKeyClusterID  = "clusterID"
	FieldKeyNamespace  = "namespace"
	FieldKeyVariableID = "variableID"
	FieldKeyScope      = "scope"
)

// NOCC:deadcode/unused(设计如此:)
var (
	variableDefineIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: FieldKeyVariableID, Value: 1},
				bson.E{Key: FieldKeyClusterID, Value: 1},
				bson.E{Key: FieldKeyNamespace, Value: 1},
				bson.E{Key: FieldKeyScope, Value: 1},
			},
			Unique: false,
		},
	}
)

// VariableValue ...
type VariableValue struct {
	// ID          string `json:"id" bson:"_id"`
	VariableID string `json:"variableID" bson:"variableID"`
	Value      string `json:"value" bson:"value"`
	Scope      string `json:"scope" bson:"scope"`
	ClusterID  string `json:"clusterID" bson:"clusterID"`
	Namespace  string `json:"namespace" bson:"namespace"`
	CreateTime string `json:"createTime" bson:"createTime"`
	UpdateTime string `json:"updateTime" bson:"updateTime"`
	Creator    string `json:"creator" bson:"creator"`
	Updater    string `json:"updater" bson:"updater"`
}

// ModelVariableValue provide variable define db
type ModelVariableValue struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New return a new variable value model instance
func New(db drivers.DB) *ModelVariableValue {
	return &ModelVariableValue{
		tableName: dbtable.DataTableNamePrefix + tableName,
		// indexes:   variableDefineIndexes,
		db: db,
	}
}

// ensure table
func (m *ModelVariableValue) ensureTable(ctx context.Context) error {
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

// CreateVariableValue create variable value
func (m *ModelVariableValue) CreateVariableValue(ctx context.Context, vv *VariableValue) error {
	if vv == nil {
		return fmt.Errorf("variable value cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{vv}); err != nil {
		return err
	}
	return nil
}

// GetVariableValue get variable value
func (m *ModelVariableValue) GetVariableValue(ctx context.Context,
	variableID, clusterID, namespace, scope string) (*VariableValue, error) {
	condM := make(operator.M)
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	if variableID == "" {
		return nil, fmt.Errorf("variableID cannot be empty")
	}
	condM[FieldKeyVariableID] = variableID
	condM[FieldKeyScope] = scope
	if clusterID != "" {
		condM[FieldKeyClusterID] = clusterID
	}
	if namespace != "" {
		condM[FieldKeyNamespace] = namespace
	}
	cond := operator.NewLeafCondition(operator.Eq, condM)
	value := &VariableValue{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, value); err != nil {
		return nil, err
	}
	return value, nil
}

// ListClusterVariableValues implement for ListClusterVariableValues interface
func (m *ModelVariableValue) ListClusterVariableValues(ctx context.Context,
	variableID string) ([]VariableValue, error) {
	condM := make(operator.M)
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	if variableID == "" {
		return nil, fmt.Errorf("variableID cannot be empty")
	}
	condM[FieldKeyVariableID] = variableID
	condM[FieldKeyScope] = vdm.VariableScopeCluster
	cond := operator.NewLeafCondition(operator.Eq, condM)
	values := make([]VariableValue, 0)
	if err := m.db.Table(m.tableName).Find(cond).All(ctx, &values); err != nil {
		return nil, err
	}
	return values, nil
}

// ListNamespaceVariableValues implement for ListNamespaceVariableValues interface
func (m *ModelVariableValue) ListNamespaceVariableValues(ctx context.Context,
	variableID, clusterID string) ([]VariableValue, error) {
	condM := make(operator.M)
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	if variableID == "" {
		return nil, fmt.Errorf("variableID cannot be empty")
	}
	condM[FieldKeyVariableID] = variableID
	condM[FieldKeyScope] = vdm.VariableScopeNamespace
	condM[FieldKeyClusterID] = clusterID
	cond := operator.NewLeafCondition(operator.Eq, condM)
	values := make([]VariableValue, 0)
	if err := m.db.Table(m.tableName).Find(cond).All(ctx, &values); err != nil {
		return nil, err
	}
	return values, nil
}

// ListVariableValuesInCluster implement for ListVariableValuesInCluster interface
func (m *ModelVariableValue) ListVariableValuesInCluster(ctx context.Context,
	clusterID string) ([]VariableValue, error) {
	condM := make(operator.M)
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	if clusterID == "" {
		return nil, fmt.Errorf("clusterID cannot be empty")
	}
	condM[FieldKeyClusterID] = clusterID
	condM[FieldKeyScope] = vdm.VariableScopeCluster
	cond := operator.NewLeafCondition(operator.Eq, condM)
	values := make([]VariableValue, 0)
	if err := m.db.Table(m.tableName).Find(cond).All(ctx, &values); err != nil {
		return nil, err
	}
	return values, nil
}

// ListVariableValuesInNamespace implement for ListVariableValuesInNamespace interface
func (m *ModelVariableValue) ListVariableValuesInNamespace(ctx context.Context,
	clusterID, namespace string) ([]VariableValue, error) {
	condM := make(operator.M)
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	if clusterID == "" {
		return nil, fmt.Errorf("clusterID cannot be empty")
	}
	if namespace == "" {
		return nil, fmt.Errorf("namespace cannot be empty")
	}
	condM[FieldKeyClusterID] = clusterID
	condM[FieldKeyNamespace] = namespace
	condM[FieldKeyScope] = vdm.VariableScopeNamespace
	cond := operator.NewLeafCondition(operator.Eq, condM)
	values := make([]VariableValue, 0)
	if err := m.db.Table(m.tableName).Find(cond).All(ctx, &values); err != nil {
		return nil, err
	}
	return values, nil
}

// ListVariableValuesInAllNamespace implement for ListVariableValuesInAllNamespace interface
func (m *ModelVariableValue) ListVariableValuesInAllNamespace(ctx context.Context,
	clusterID string) ([]VariableValue, error) {
	condM := make(operator.M)
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	if clusterID == "" {
		return nil, fmt.Errorf("clusterID cannot be empty")
	}
	condM[FieldKeyClusterID] = clusterID
	condM[FieldKeyScope] = vdm.VariableScopeNamespace
	cond := operator.NewLeafCondition(operator.Eq, condM)
	values := make([]VariableValue, 0)
	if err := m.db.Table(m.tableName).Find(cond).All(ctx, &values); err != nil {
		return nil, err
	}
	return values, nil
}

// UpsertVariableValue upsert variable value
func (m *ModelVariableValue) UpsertVariableValue(ctx context.Context,
	value *VariableValue) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	condM := make(operator.M)
	condM[FieldKeyVariableID] = value.VariableID
	condM[FieldKeyScope] = value.Scope
	if value.ClusterID != "" {
		condM[FieldKeyClusterID] = value.ClusterID
	}
	if value.Namespace != "" {
		condM[FieldKeyNamespace] = value.Namespace
	}
	cond := operator.NewLeafCondition(operator.Eq, condM)
	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": value})
}

// DeleteVariableValuesByNamespace batch delete variable value records
func (m *ModelVariableValue) DeleteVariableValuesByNamespace(ctx context.Context,
	clusterID, namespace string) (int64, error) {
	if err := m.ensureTable(ctx); err != nil {
		return 0, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		FieldKeyClusterID: clusterID,
		FieldKeyNamespace: namespace,
	})
	return m.db.Table(m.tableName).Delete(ctx, cond)
}
