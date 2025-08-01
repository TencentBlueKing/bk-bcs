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

// Package storage xxx
package storage

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/storage/audit"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/storage/entity"
)

// Storage 提供了数据库操作的接口
type Storage interface {
	// Audit operation
	GetAudit(ctx context.Context, projectCode, clusterID string) (*entity.Audit, error)
	CreateAudit(ctx context.Context, audit *entity.Audit) (primitive.ObjectID, error)
	UpdateAudit(ctx context.Context, id string, audit entity.M) error
	DeleteAudit(ctx context.Context, id string) error
	FirstAuditOrCreate(ctx context.Context, audit *entity.Audit) (*entity.Audit, error)
}

type modelSet struct {
	*audit.ModelAudit
}

// New return a new ResourceManagerModel instance
func New(db drivers.DB) Storage {
	return &modelSet{
		ModelAudit: audit.New(db),
	}
}
