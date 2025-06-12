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

// Package mesh provides mesh-related storage operations for the mesh manager
package mesh

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/utils"
)

const (
	tableName = "mesh"
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

// ModelMesh provides database operations for mesh
type ModelMesh struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.Mutex
}

// New returns a new ModelMesh instance
func New(db drivers.DB) *ModelMesh {
	return &ModelMesh{
		tableName: utils.DataTableNamePrefix + tableName,
		indexes:   tableIndexes,
		db:        db,
	}
}

func (m *ModelMesh) ensureTable(ctx context.Context) error {
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

// ListMesh queries a list of meshes based on conditions and options
func (m *ModelMesh) ListMesh(ctx context.Context, cond *operator.Condition, opt *utils.ListOption) (
	int64, []*entity.Mesh, error) {
	if err := m.ensureTable(ctx); err != nil {
		return 0, nil, err
	}

	l := make([]*entity.Mesh, 0)
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

// UpdateMesh updates an existing mesh
func (m *ModelMesh) UpdateMesh(ctx context.Context, meshID string, mesh entity.M) error {
	if meshID == "" {
		return fmt.Errorf("meshID cannot be empty")
	}

	if mesh == nil {
		return fmt.Errorf("mesh cannot be empty")
	}

	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyMeshID: meshID,
	})

	// Check if mesh exists
	old := &entity.Mesh{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, old); err != nil {
		return err
	}

	// Set update time if not provided
	if mesh[entity.FieldKeyUpdateTime] == nil {
		mesh[entity.FieldKeyUpdateTime] = time.Now().Unix()
	}

	if err := m.db.Table(m.tableName).Update(ctx, cond, operator.M{"$set": mesh}); err != nil {
		return fmt.Errorf("update mesh %s failed: %v", meshID, err)
	}

	return nil
}

// DeleteMesh deletes a mesh by its ID
func (m *ModelMesh) DeleteMesh(ctx context.Context, meshID string) error {
	if meshID == "" {
		return fmt.Errorf("meshID cannot be empty")
	}

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

func (m *ModelMesh) generateMeshID() string {
	return fmt.Sprintf("mesh-%s", uuid.New().String())
}

// CreateMesh creates a new mesh
func (m *ModelMesh) CreateMesh(ctx context.Context, mesh *entity.Mesh) error {
	if mesh == nil {
		return fmt.Errorf("mesh cannot be empty")
	}

	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	// Generate mesh ID and set basic fields
	mesh.MeshID = m.generateMeshID()
	now := time.Now().Unix()
	mesh.CreateTime = now
	mesh.UpdateTime = now

	_, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{mesh})
	if err != nil {
		return fmt.Errorf("create mesh failed: %v", err)
	}

	return nil
}
