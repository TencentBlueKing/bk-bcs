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

// Package store 提供mesh manager的store功能，DB相关操作
package store

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/istio"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/utils"
)

// MeshManagerModel defines the database operation interface for service mesh management
type MeshManagerModel interface {
	// Create creates a new mesh
	Create(ctx context.Context, mesh *entity.MeshIstio) error
	// Update updates an existing mesh
	Update(ctx context.Context, meshID string, mesh entity.M) error
	// Delete deletes a mesh by its ID
	Delete(ctx context.Context, meshID string) error
	// List queries a list of meshes based on conditions and options
	List(ctx context.Context, cond *operator.Condition, opt *utils.ListOption) (int64, []*entity.MeshIstio, error)
	// Get gets a mesh by its ID
	Get(ctx context.Context, cond *operator.Condition) (*entity.MeshIstio, error)
}

// modelSet implements MeshManagerModel by embedding ModelMesh
type modelSet struct {
	*istio.ModelMeshIstio
}

// New returns a new instance of MeshManagerModel
func New(db drivers.DB) MeshManagerModel {
	return &modelSet{
		ModelMeshIstio: istio.New(db),
	}
}
