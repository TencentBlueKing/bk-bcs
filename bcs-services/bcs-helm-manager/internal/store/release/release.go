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

// Package release xxx
package release

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/utils"
)

const (
	tableName = "release"
)

var (
	tableIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: entity.FieldKeyClusterID, Value: 1},
				bson.E{Key: entity.FieldKeyNamespace, Value: 1},
				bson.E{Key: entity.FieldKeyName, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelRelease provides handling release-related operations to database
type ModelRelease struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.Mutex
}

// New return a new ModelRelease instance
func New(db drivers.DB) *ModelRelease {
	return &ModelRelease{
		tableName: utils.DataTableNamePrefix + tableName,
		indexes:   tableIndexes,
		db:        db,
	}
}

func (m *ModelRelease) ensureTable(ctx context.Context) error {
	if m.isTableEnsured {
		return nil
	}

	m.isTableEnsuredMutex.Lock()
	defer m.isTableEnsuredMutex.Unlock()
	if m.isTableEnsured {
		return nil
	}

	if err := utils.EnsureTable(ctx, m.db, m.tableName, m.indexes); err != nil {
		return err
	}
	m.isTableEnsured = true
	return nil
}

// CreateRelease create a new entity.Release into database
func (m *ModelRelease) CreateRelease(ctx context.Context, release *entity.Release) error {
	if release == nil {
		return fmt.Errorf("can not create empty release")
	}

	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	timestamp := time.Now().UTC().Unix()
	if release.CreateTime == 0 {
		release.CreateTime = timestamp
	}
	if release.UpdateTime == 0 {
		release.UpdateTime = timestamp
	}
	if len(release.UpdateBy) == 0 {
		release.UpdateBy = release.CreateBy
	}
	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{release}); err != nil {
		return err
	}
	return nil
}

// UpdateRelease update an entity.Release into database
func (m *ModelRelease) UpdateRelease(ctx context.Context, clusterID, namespace, name string, release entity.M) error {
	if clusterID == "" || namespace == "" || name == "" {
		return fmt.Errorf("can not update with empty clusterID, namespace or name")
	}

	if release == nil {
		return fmt.Errorf("can not update empty release")
	}

	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyClusterID: clusterID,
		entity.FieldKeyNamespace: namespace,
		entity.FieldKeyName:      name,
	})
	old := &entity.Release{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, old); err != nil {
		return err
	}

	if release[entity.FieldKeyUpdateTime] == nil {
		release.Update(entity.FieldKeyUpdateTime, time.Now().UTC().Unix())
	}
	if err := m.db.Table(m.tableName).Update(ctx, cond, operator.M{"$set": release}); err != nil {
		return err
	}

	return nil
}

// GetRelease get a specific entity.Release from database
func (m *ModelRelease) GetRelease(ctx context.Context, clusterID, namespace, name string) (
	*entity.Release, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("can not get with empty clusterID")
	}
	if namespace == "" {
		return nil, fmt.Errorf("can not get with empty namespace")
	}
	if name == "" {
		return nil, fmt.Errorf("can not get with empty name")
	}

	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyClusterID: clusterID,
		entity.FieldKeyNamespace: namespace,
		entity.FieldKeyName:      name,
	})

	release := &entity.Release{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, release); err != nil {
		return nil, err
	}

	return release, nil
}

// ListRelease get a list of entity.Release by condition and option from database
func (m *ModelRelease) ListRelease(ctx context.Context, cond *operator.Condition, opt *utils.ListOption) (
	int64, []*entity.Release, error) {

	l := make([]*entity.Release, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(common.MapInt2MapIf(opt.Sort))
	}
	if opt.Page != 0 {
		finder = finder.WithStart(opt.Page * opt.Size)
	}
	if opt.Size != 0 {
		finder = finder.WithLimit(opt.Size)
	}

	if err := finder.All(ctx, &l); err != nil {
		return 0, nil, err
	}

	total, err := finder.Count(ctx)
	if err != nil {
		return 0, nil, err
	}

	return total, l, nil
}

// DeleteRelease delete a specific entity.Release from database
func (m *ModelRelease) DeleteRelease(ctx context.Context, clusterID, namespace, name string) error {
	if clusterID == "" {
		return fmt.Errorf("can not get with empty clusterID")
	}
	if namespace == "" {
		return fmt.Errorf("can not get with empty namespace")
	}
	if name == "" {
		return fmt.Errorf("can not get with empty name")
	}

	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyClusterID: clusterID,
		entity.FieldKeyNamespace: namespace,
		entity.FieldKeyName:      name,
	})

	if _, err := m.db.Table(m.tableName).Delete(ctx, cond); err != nil {
		return err
	}

	return nil
}
