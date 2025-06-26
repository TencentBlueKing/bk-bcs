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

// Package istio provides istio-related storage operations for the mesh manager
package istio

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/utils"
)

const (
	tableName = "istio"
)

var (
	tableIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: entity.FieldKeyMeshID, Value: 1},
			},
			Unique: true,
		},
		{
			Name: tableName + "_name_idx",
			Key: bson.D{
				bson.E{Key: entity.FieldKeyMeshName, Value: 1},
			},
			Unique: false,
		},
	}
)

// ModelMeshIstio provides database operations for mesh istio
type ModelMeshIstio struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.Mutex
}

// New returns a new ModelMeshIstio instance
func New(db drivers.DB) *ModelMeshIstio {
	return &ModelMeshIstio{
		tableName: utils.DataTableNamePrefix + tableName,
		indexes:   tableIndexes,
		db:        db,
	}
}

func (m *ModelMeshIstio) ensureTable(ctx context.Context) error {
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

// List queries a list of meshes based on conditions and options
func (m *ModelMeshIstio) List(ctx context.Context, cond *operator.Condition, opt *utils.ListOption) (
	int64, []*entity.MeshIstio, error) {
	if err := m.ensureTable(ctx); err != nil {
		return 0, nil, err
	}

	l := make([]*entity.MeshIstio, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(common.MapInt2MapIf(opt.Sort))
	}
	if opt.Page > 0 && opt.Size > 0 {
		finder = finder.WithStart((opt.Page - 1) * opt.Size)
	}
	if opt.Size > 0 {
		finder = finder.WithLimit(opt.Size)
	}

	if err := finder.All(ctx, &l); err != nil {
		return 0, nil, fmt.Errorf("find mesh list failed: %v", err)
	}

	total, err := finder.Count(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("count mesh list failed: %v", err)
	}

	return total, l, nil
}

// Update updates an existing mesh istio
func (m *ModelMeshIstio) Update(ctx context.Context, meshID string, entityM entity.M) error {
	if meshID == "" {
		return fmt.Errorf("meshID cannot be empty")
	}

	// If entityM is nil, do nothing
	if entityM == nil {
		return nil
	}

	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyMeshID: meshID,
	})

	// Check if mesh exists
	old := &entity.MeshIstio{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, old); err != nil {
		return err
	}

	// Set update time if not provided
	if entityM[entity.FieldKeyUpdateTime] == nil {
		entityM[entity.FieldKeyUpdateTime] = time.Now().Unix()
	}

	if err := m.db.Table(m.tableName).Update(ctx, cond, operator.M{"$set": entityM}); err != nil {
		return fmt.Errorf("update mesh %s failed: %v", meshID, err)
	}

	return nil
}

// Delete deletes a mesh istio by its ID
func (m *ModelMeshIstio) Delete(ctx context.Context, meshID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyMeshID: meshID,
	})

	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return fmt.Errorf("delete mesh failed: %v", err)
	}

	return nil
}

// Create creates a new mesh istio
func (m *ModelMeshIstio) Create(ctx context.Context, mesh *entity.MeshIstio) error {
	if mesh == nil {
		return fmt.Errorf("mesh cannot be empty")
	}

	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	now := time.Now().Unix()
	mesh.CreateTime = now
	mesh.UpdateTime = now

	_, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{mesh})
	if err != nil {
		return fmt.Errorf("create mesh failed: %v", err)
	}

	return nil
}

// Get gets a mesh by its ID
func (m *ModelMeshIstio) Get(ctx context.Context, cond *operator.Condition) (*entity.MeshIstio, error) {

	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}

	mesh := &entity.MeshIstio{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, mesh); err != nil {
		return nil, err
	}

	return mesh, nil
}
