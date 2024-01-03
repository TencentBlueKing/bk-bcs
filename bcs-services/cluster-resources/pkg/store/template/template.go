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

// Package template template metadata
package template

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/utils"
)

const (
	tableName = "template"
)

var (
	tableIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: entity.FieldKeyProjectCode, Value: 1},
				bson.E{Key: entity.FieldKeyName, Value: 1},
				bson.E{Key: entity.FieldKeyTemplateSpace, Value: 1},
			},
			Unique: false,
		},
	}
)

// ModelTemplate provides handling template operations to database
type ModelTemplate struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.Mutex
}

// New return a new ModelTemplate instance
func New(db drivers.DB) *ModelTemplate {
	return &ModelTemplate{
		tableName: utils.DataTableNamePrefix + tableName,
		indexes:   tableIndexes,
		db:        db,
	}
}

// ensure object database table and table indexes
func (m *ModelTemplate) ensureTable(ctx context.Context) error {
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

// GetTemplate get a specific entity.Template from database
func (m *ModelTemplate) GetTemplate(ctx context.Context, id string) (*entity.Template, error) {
	if id == "" {
		return nil, fmt.Errorf("can not get with empty id")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyObjectID: objectID,
	})

	template := &entity.Template{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, template); err != nil {
		return nil, err
	}

	return template, nil
}

// ListTemplate get a list of entity.Template by condition from database
func (m *ModelTemplate) ListTemplate(
	ctx context.Context, cond *operator.Condition) ([]*entity.Template, error) {

	t := make([]*entity.Template, 0)
	err := m.db.Table(m.tableName).Find(cond).All(ctx, &t)

	if err != nil {
		return nil, err
	}

	return t, nil
}

// CreateTemplate create a new entity.Template into database
func (m *ModelTemplate) CreateTemplate(
	ctx context.Context, template *entity.Template) (string, error) {
	if template == nil {
		return "", fmt.Errorf("can not create empty template")
	}

	if err := m.ensureTable(ctx); err != nil {
		return "", err
	}

	// 没有id的情况下生成
	now := time.Now()
	if template.ID.IsZero() {
		template.ID = primitive.NewObjectIDFromTimestamp(now)
	}

	if template.CreateAt == 0 {
		template.CreateAt = now.UTC().Unix()
	}
	if template.UpdateAt == 0 {
		template.UpdateAt = now.UTC().Unix()
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{template}); err != nil {
		return "", err
	}
	return template.ID.Hex(), nil
}

// UpdateTemplate update an entity.Template into database
func (m *ModelTemplate) UpdateTemplate(ctx context.Context, id string, template entity.M) error {
	if id == "" {
		return fmt.Errorf("can not update with empty id")
	}

	if template == nil {
		return fmt.Errorf("can not update empty template")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyObjectID: objectID,
	})
	old := &entity.Template{}
	if err = m.db.Table(m.tableName).Find(cond).One(ctx, old); err != nil {
		return err
	}

	if template[entity.FieldKeyUpdateAt] == nil {
		template.Update(entity.FieldKeyUpdateAt, time.Now().UTC().Unix())
	}
	if err := m.db.Table(m.tableName).Update(ctx, cond, operator.M{"$set": template}); err != nil {
		return err
	}

	return nil
}

// UpdateTemplateBySpecial update a Special entity.Template into database
func (m *ModelTemplate) UpdateTemplateBySpecial(
	ctx context.Context, projectCode, templateSpace string, template entity.M) error {
	// 模板文件夹名称不能为空
	if templateSpace == "" || projectCode == "" {
		return fmt.Errorf("can not delete with empty templateSpace or projectCode")
	}

	if template == nil {
		return fmt.Errorf("can not update empty templateVersion")
	}

	operatorM := operator.M{
		entity.FieldKeyProjectCode:   projectCode,
		entity.FieldKeyTemplateSpace: templateSpace,
	}

	cond := operator.NewLeafCondition(operator.Eq, operatorM)

	total, err := m.db.Table(m.tableName).Find(cond).Count(ctx)
	if err != nil {
		return err
	}

	// 没有数据则不更新
	if total == 0 {
		return nil
	}

	if template[entity.FieldKeyUpdateAt] == nil {
		template.Update(entity.FieldKeyUpdateAt, time.Now().UTC().Unix())
	}

	if _, err = m.db.Table(m.tableName).UpdateMany(ctx, cond, operator.M{"$set": template}); err != nil {
		return err
	}

	return nil
}

// DeleteTemplate delete a specific entity.Template from database
func (m *ModelTemplate) DeleteTemplate(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("can not delete with empty id")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyObjectID: objectID,
	})

	if _, err := m.db.Table(m.tableName).Delete(ctx, cond); err != nil {
		return err
	}

	return nil
}

// DeleteTemplateBySpecial delete a specific entity.Template from database
func (m *ModelTemplate) DeleteTemplateBySpecial(
	ctx context.Context, projectCode, templateSpace string) error {
	// 模板文件夹名称不能为空
	if templateSpace == "" || projectCode == "" {
		return fmt.Errorf("can not delete with empty templateName or projectCode")
	}

	operatorM := operator.M{
		entity.FieldKeyProjectCode:   projectCode,
		entity.FieldKeyTemplateSpace: templateSpace,
	}

	cond := operator.NewLeafCondition(operator.Eq, operatorM)

	if _, err := m.db.Table(m.tableName).Delete(ctx, cond); err != nil {
		return err
	}

	return nil
}
