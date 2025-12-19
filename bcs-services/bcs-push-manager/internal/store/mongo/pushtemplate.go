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

package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/store/types"
)

var (
	modelPushTemplateIndexes = []drivers.Index{
		{
			Key: bson.D{
				bson.E{Key: pushDomainKey, Value: 1},
				bson.E{Key: pushTemplateUniqueKey, Value: 1},
			},
			Name:   pushTemplateTableName + "_1",
			Unique: true,
		},
		{
			Key: bson.D{
				bson.E{Key: pushTemplateUniqueKey, Value: 1},
			},
			Name:   pushTemplateUniqueKey + "_1",
			Unique: true,
		},
	}
)

// ModelPushTemplate is a MongoDB-based implementation of PushTemplateStore.
type ModelPushTemplate struct {
	Public
}

// NewModelPushTemplate creates a new PushTemplateStore instance.
func NewModelPushTemplate(db drivers.DB) *ModelPushTemplate {
	return &ModelPushTemplate{
		Public: Public{
			TableName: tableNamePrefix + pushTemplateTableName,
			Indexes:   modelPushTemplateIndexes,
			DB:        db,
		}}
}

// CreatePushTemplate inserts a new push template into the database.
func (m *ModelPushTemplate) CreatePushTemplate(ctx context.Context, template *types.PushTemplate) error {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return fmt.Errorf("ensure table failed: %v", err)
	}
	if template == nil {
		return fmt.Errorf("push template is nil")
	}

	template.ID = primitive.NewObjectID()
	template.CreatedAt = time.Now()

	if _, err := m.DB.Table(m.TableName).Insert(ctx, []interface{}{template}); err != nil {
		return fmt.Errorf("create push template failed: %v", err)
	}
	return nil
}

// DeletePushTemplate deletes a push template from the database by template_id.
func (m *ModelPushTemplate) DeletePushTemplate(ctx context.Context, templateID string) error {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return fmt.Errorf("ensure table failed: %v", err)
	}
	if templateID == "" {
		return fmt.Errorf("templateID is empty")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		pushTemplateUniqueKey: templateID,
	})

	if _, err := m.DB.Table(m.TableName).Delete(ctx, cond); err != nil {
		return fmt.Errorf("delete push template failed: %v", err)
	}
	return nil
}

// GetPushTemplate retrieves a single push template from the database by template_id.
func (m *ModelPushTemplate) GetPushTemplate(ctx context.Context, templateID string) (*types.PushTemplate, error) {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return nil, fmt.Errorf("ensure table failed: %v", err)
	}
	if templateID == "" {
		return nil, fmt.Errorf("templateID cannot be empty")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		pushTemplateUniqueKey: templateID,
	})

	var template types.PushTemplate
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, &template); err != nil {
		return nil, fmt.Errorf("get push template failed: %v", err)
	}
	return &template, nil
}

// ListPushTemplates retrieves a list of push templates from the database with filtering and pagination.
func (m *ModelPushTemplate) ListPushTemplates(ctx context.Context, filter operator.M, page, pageSize int64) ([]*types.PushTemplate, int64, error) {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return nil, 0, fmt.Errorf("ensure table failed: %v", err)
	}
	if page < 1 {
		return nil, 0, fmt.Errorf("invalid page: must be greater than or equal to 1")
	}
	if pageSize <= 0 {
		return nil, 0, fmt.Errorf("invalid pageSize: must be greater than 0")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{})
	if filter != nil {
		cond = operator.NewBranchCondition(operator.And, cond, operator.NewLeafCondition(operator.Eq, filter))
	}

	var templates []*types.PushTemplate
	finder := m.DB.Table(m.TableName).Find(cond)
	if page > 1 {
		finder = finder.WithStart((page - 1) * pageSize)
	}
	if pageSize > 0 {
		finder = finder.WithLimit(pageSize)
	}
	if err := finder.All(ctx, &templates); err != nil {
		return nil, 0, fmt.Errorf("list push templates failed: %v", err)
	}

	total, err := finder.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count push templates failed: %v", err)
	}

	return templates, total, nil
}

// UpdatePushTemplate updates a push template in the database.
func (m *ModelPushTemplate) UpdatePushTemplate(ctx context.Context, templateID string, update operator.M) error {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return fmt.Errorf("ensure table failed: %v", err)
	}
	if templateID == "" {
		return fmt.Errorf("templateID cannot be empty")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		pushTemplateUniqueKey: templateID,
	})
	if update == nil {
		return fmt.Errorf("update cannot be nil")
	}
	set, ok := update["$set"].(operator.M)
	if !ok {
		return fmt.Errorf("invalid update format: $set must be operator.M type")
	}
	if set == nil {
		set = operator.M{}
		update["$set"] = set
	}
	set["updated_at"] = time.Now()

	if err := m.DB.Table(m.TableName).Update(ctx, cond, update); err != nil {
		return fmt.Errorf("update push template failed: %v", err)
	}
	return nil
}
