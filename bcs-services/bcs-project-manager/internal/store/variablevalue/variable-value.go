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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/dbtable"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	// table name
	tableName        = "variable_value"
	idField          = "_id"
	projectCodeField = "projectCode"
	clusterIDField   = "clusterID"
	nameSpaceField   = "namespace"
	keyField         = "key"
)

var (
	variableDefineIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: idField, Value: 1},
				bson.E{Key: projectCodeField, Value: 1},
				bson.E{Key: keyField, Value: 1},
				bson.E{Key: clusterIDField, Value: 1},
				bson.E{Key: nameSpaceField, Value: 1},
			},
			// TODO: 确认 unique 含义
			Unique: true,
		},
	}
)

// VariableValue ...
type VariableValue struct {
	ID               string `json:"id" bson:"_id"`
	VariableDefineID string `json:"variableDefineID" bson:"variableDefineID"`
	Key              string `json:"key" bson:"key"`
	Name             string `json:"name" bson:"name"`
	ProjectCode      string `json:"projectCode" bson:"projectCode"`
	Value            string `json:"value" bson:"value"`
	Scope            string `json:"scope" bson:"scope"`
	ClusterID        string `json:"clusterID" bson:"clusterID"`
	Namespace        string `json:"namespace" bson:"namespace"`
	CreateTime       string `json:"createTime" bson:"createTime"`
	UpdateTime       string `json:"updateTime" bson:"updateTime"`
	Creator          string `json:"creator" bson:"creator"`
	Updater          string `json:"updater" bson:"updater"`
	IsDeleted        bool   `json:"isDeleted" bson:"isDeleted"`
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
		indexes:   variableDefineIndexes,
		db:        db,
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
	projectCode, key, clusterID, namespace, scope string) (*VariableValue, error) {
	condM := make(operator.M)
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	if key == "" {
		return nil, fmt.Errorf("key cannot be empty")
	}
	condM["key"] = key
	condM["projectCode"] = projectCode
	condM["scope"] = scope
	if clusterID != "" {
		condM["clusterID"] = clusterID
	}
	if namespace != "" {
		condM["namespace"] = namespace
	}
	cond := operator.NewLeafCondition(operator.Eq, condM)
	value := &VariableValue{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, value); err != nil {
		return nil, err
	}
	return value, nil
}
