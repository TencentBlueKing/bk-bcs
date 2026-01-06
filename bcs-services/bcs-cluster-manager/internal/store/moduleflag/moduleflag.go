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

// Package moduleflag xxx
package moduleflag

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
	cloudIDKeyName             = "cloudid"
	versionKeyName             = "version"
	moduleIDKeyName            = "moduleid"
	flagNameKeyName            = "flagname"
	cloudModuleFlagTableName   = "cloudmoduleflag"
	defaultNamespaceListLength = 1000
)

var (
	cloudVersionModuleFlagIndexes = []drivers.Index{
		{
			Name: cloudModuleFlagTableName + "_idx",
			Key: bson.D{
				bson.E{Key: cloudIDKeyName, Value: 1},
				bson.E{Key: versionKeyName, Value: 1},
				bson.E{Key: moduleIDKeyName, Value: 1},
				bson.E{Key: flagNameKeyName, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelCloudModuleFlag database operation for cloudmoduleflag
type ModelCloudModuleFlag struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New create namespace model
func New(db drivers.DB) *ModelCloudModuleFlag {
	return &ModelCloudModuleFlag{
		tableName: util.DataTableNamePrefix + cloudModuleFlagTableName,
		indexes:   cloudVersionModuleFlagIndexes,
		db:        db,
	}
}

// ensure table
func (m *ModelCloudModuleFlag) ensureTable(ctx context.Context) error {
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

// CreateCloudModuleFlag create cloud module flag
func (m *ModelCloudModuleFlag) CreateCloudModuleFlag(ctx context.Context, flag *types.CloudModuleFlag) error {
	if flag == nil {
		return fmt.Errorf("flag to be created cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{flag}); err != nil {
		return err
	}
	return nil
}

// UpdateCloudModuleFlag update cloud module flag
func (m *ModelCloudModuleFlag) UpdateCloudModuleFlag(ctx context.Context, flag *types.CloudModuleFlag) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		cloudIDKeyName:  flag.CloudID,
		versionKeyName:  flag.Version,
		moduleIDKeyName: flag.ModuleID,
		flagNameKeyName: flag.FlagName,
	})
	oldFlag := &types.CloudModuleFlag{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, oldFlag); err != nil {
		return err
	}
	return m.db.Table(m.tableName).Update(ctx, cond, operator.M{"$set": flag})
}

// DeleteCloudModuleFlag delete cloud module flag
func (m *ModelCloudModuleFlag) DeleteCloudModuleFlag(ctx context.Context, cloudID, version, module, flag string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	flagCond := operator.M{
		cloudIDKeyName:  cloudID,
		versionKeyName:  version,
		moduleIDKeyName: module,
	}
	if len(flag) > 0 {
		flagCond[flagNameKeyName] = flag
	}

	cond := operator.NewLeafCondition(operator.Eq, flagCond)
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// GetCloudModuleFlag get cloud moduleFlag
func (m *ModelCloudModuleFlag) GetCloudModuleFlag(ctx context.Context, cloudID, version, module, flag string) (
	*types.CloudModuleFlag, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		cloudIDKeyName:  cloudID,
		versionKeyName:  version,
		moduleIDKeyName: module,
		flagNameKeyName: flag,
	})

	retFlag := &types.CloudModuleFlag{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, retFlag); err != nil {
		return nil, err
	}
	retFlag.CreatTime = util.TransStrToUTCStr(time.RFC3339Nano, retFlag.CreatTime)
	retFlag.UpdateTime = util.TransStrToUTCStr(time.RFC3339Nano, retFlag.UpdateTime)
	return retFlag, nil
}

// ListCloudModuleFlag list cloud moduleFlag
func (m *ModelCloudModuleFlag) ListCloudModuleFlag(ctx context.Context, cond *operator.Condition,
	opt *options.ListOption) (
	[]*types.CloudModuleFlag, error) {
	retFlagList := make([]*types.CloudModuleFlag, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(util.MapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit == 0 {
		finder = finder.WithLimit(defaultNamespaceListLength)
	} else {
		finder = finder.WithLimit(opt.Limit)
	}
	if err := finder.All(ctx, &retFlagList); err != nil {
		return nil, err
	}
	return retFlagList, nil
}
