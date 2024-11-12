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

// Package config xxx
package config

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/dbtable"
)

const (
	// table name
	tableName = "config"
	// FieldKeyKey config key
	FieldKeyKey = "key"
	// FieldKeyValue config value
	FieldKeyValue = "value"

	// ConfigKeyCreateNamespaceItsmServiceID used to create an itsm ticket when creating a namespace in a shared cluster
	ConfigKeyCreateNamespaceItsmServiceID = "create_namespace_itsm_service_id"
	// ConfigKeyUpdateNamespaceItsmServiceID used to create an itsm ticket when updating a namespace in a shared cluster
	ConfigKeyUpdateNamespaceItsmServiceID = "update_namespace_itsm_service_id"
	// ConfigKeyDeleteNamespaceItsmServiceID used to create an itsm ticket when deleting a namespace in a shared cluster
	ConfigKeyDeleteNamespaceItsmServiceID = "delete_namespace_itsm_service_id"

	// QuotaManagerCommonItsmServiceID used to create an itsm ticket when quota manager
	QuotaManagerCommonItsmServiceID = "quota_manager_common_itsm_service_id"
)

// NOCC:deadcode/unused(设计如此:)
// nolint
var (
	configIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: FieldKeyKey, Value: 1},
			},
			Unique: true,
		},
	}
)

// Config ...
type Config struct {
	// ID          string `json:"id" bson:"_id"`
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}

// ModelConfig provide config define db
type ModelConfig struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New return a new variable value model instance
func New(db drivers.DB) *ModelConfig {
	return &ModelConfig{
		tableName: dbtable.DataTableNamePrefix + tableName,
		indexes:   configIndexes,
		db:        db,
	}
}

// ensure table
func (m *ModelConfig) ensureTable(ctx context.Context) error {
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

// GetConfig get config value
func (m *ModelConfig) GetConfig(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("config key cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return "", err
	}
	condM := make(operator.M)
	condM[FieldKeyKey] = key
	cond := operator.NewLeafCondition(operator.Eq, condM)
	conf := &Config{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, conf); err != nil {
		return "", err
	}
	return conf.Value, nil
}

// SetConfig set config value
func (m *ModelConfig) SetConfig(ctx context.Context, key, value string) error {
	if key == "" {
		return fmt.Errorf("config key cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	condM := make(operator.M)
	condM[FieldKeyKey] = key
	cond := operator.NewLeafCondition(operator.Eq, condM)

	config := &Config{
		Key:   key,
		Value: value,
	}
	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": config})
}
