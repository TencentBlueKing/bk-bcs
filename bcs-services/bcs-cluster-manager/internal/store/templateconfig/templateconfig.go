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

// Package templateconfig xxx
package templateconfig

import (
	"context"
	"fmt"
	"sync"
	"time"

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

	tableName = "templateconfig"
	// ConfigIDKey xxx
	ConfigIDKey = "templateconfigid"
	// ProjectIDKey xxx
	ProjectIDKey = "projectid"
	// BusinessIDKey xxx
	BusinessIDKey = "businessid"
	// ClusterIDKey xxx
	ClusterIDKey = "clusterid"
	// ProviderKey xxx
	ProviderKey = "provider"
	// ConfigTypeKey xxx
	ConfigTypeKey = "configtype"
	// defaultCloudAccountListLength xxx
	defaultTemplateConfigListLength = 1000
)

var (
	cloudIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: BusinessIDKey, Value: 1},
				bson.E{Key: ProviderKey, Value: 1},
				bson.E{Key: ConfigTypeKey, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelTemplateConfig database operation for templateConfig
type ModelTemplateConfig struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New create nodeTemplate model
func New(db drivers.DB) *ModelTemplateConfig {
	return &ModelTemplateConfig{
		tableName: util.DataTableNamePrefix + tableName,
		indexes:   cloudIndexes,
		db:        db,
	}
}

// ensure table
func (m *ModelTemplateConfig) ensureTable(ctx context.Context) error {
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

// CreateTemplateConfig insert templateConfig
func (m *ModelTemplateConfig) CreateTemplateConfig(ctx context.Context, templateConfig *types.TemplateConfig) error {
	if templateConfig == nil {
		return fmt.Errorf("templateConfig to be created cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{templateConfig}); err != nil {
		return err
	}
	return nil
}

// UpdateTemplateConfig update templateConfig with all fileds, if some fields are nil
// that field will be overwrite with empty
func (m *ModelTemplateConfig) UpdateTemplateConfig(ctx context.Context, templateConfig *types.TemplateConfig) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		ConfigIDKey: templateConfig.TemplateConfigID,
	})

	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": templateConfig})
}

// DeleteTemplateConfig delete templateConfig
func (m *ModelTemplateConfig) DeleteTemplateConfig(ctx context.Context, templateConfigID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		ConfigIDKey: templateConfigID,
	})
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// GetTemplateConfig get templateConfig
func (m *ModelTemplateConfig) GetTemplateConfig(ctx context.Context, businessID, provider, configType string) (
	*types.TemplateConfig, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		BusinessIDKey: businessID,
		ProviderKey:   provider,
		ConfigTypeKey: configType,
	})

	nodeTemplate := &types.TemplateConfig{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, nodeTemplate); err != nil {
		return nil, err
	}

	// 兼容旧数据
	nodeTemplate.UpdateTime = util.TransStrToUTCStr(time.RFC3339Nano, nodeTemplate.UpdateTime)
	nodeTemplate.CreateTime = util.TransStrToUTCStr(time.RFC3339Nano, nodeTemplate.CreateTime)

	return nodeTemplate, nil
}

// GetTemplateConfigByID get TemplateConfig by ID
func (m *ModelTemplateConfig) GetTemplateConfigByID(ctx context.Context, templateConfigID string) (
	*types.TemplateConfig, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		ConfigIDKey: templateConfigID,
	})

	nodeTemplate := &types.TemplateConfig{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, nodeTemplate); err != nil {
		return nil, err
	}

	// 兼容旧数据
	nodeTemplate.UpdateTime = util.TransStrToUTCStr(time.RFC3339Nano, nodeTemplate.UpdateTime)
	nodeTemplate.CreateTime = util.TransStrToUTCStr(time.RFC3339Nano, nodeTemplate.CreateTime)

	return nodeTemplate, nil
}

// ListTemplateConfigs list templateConfigs
func (m *ModelTemplateConfig) ListTemplateConfigs(ctx context.Context, cond *operator.Condition,
	opt *options.ListOption) (
	[]*types.TemplateConfig, error) {
	templateConfigList := make([]*types.TemplateConfig, 0)

	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(util.MapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit == 0 {
		finder = finder.WithLimit(defaultTemplateConfigListLength)
	} else {
		finder = finder.WithLimit(opt.Limit)
	}

	if err := finder.All(ctx, &templateConfigList); err != nil {
		return nil, err
	}

	return templateConfigList, nil
}
