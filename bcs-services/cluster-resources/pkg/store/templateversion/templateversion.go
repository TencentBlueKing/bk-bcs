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

// Package templateversion template version
package templateversion

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/sync/errgroup"

	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/utils"
)

const (
	tableName = "templateversion"
)

var (
	tableIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: entity.FieldKeyProjectCode, Value: 1},
				bson.E{Key: entity.FieldKeyTemplateName, Value: 1},
				bson.E{Key: entity.FieldKeyVersion, Value: 1},
			},
			Unique: false,
		},
	}
)

// ModelTemplateVersion provides handling templateVersion version operations to database
type ModelTemplateVersion struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.Mutex
}

// New return a new ModelTemplateVersion instance
func New(db drivers.DB) *ModelTemplateVersion {
	return &ModelTemplateVersion{
		tableName: utils.DataTableNamePrefix + tableName,
		indexes:   tableIndexes,
		db:        db,
	}
}

// ensure object database table and table indexes
func (m *ModelTemplateVersion) ensureTable(ctx context.Context) error {
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

// GetTemplateVersion get a specific entity.TemplateVersion from database
func (m *ModelTemplateVersion) GetTemplateVersion(ctx context.Context, id string) (*entity.TemplateVersion, error) {
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

	templateVersion := &entity.TemplateVersion{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, templateVersion); err != nil {
		return nil, err
	}

	return templateVersion, nil
}

// ListTemplateVersion get a list of entity.TemplateVersion by condition from database
func (m *ModelTemplateVersion) ListTemplateVersion(
	ctx context.Context, cond *operator.Condition) ([]*entity.TemplateVersion, error) {

	t := make([]*entity.TemplateVersion, 0)
	err := m.db.Table(m.tableName).Find(cond).All(ctx, &t)

	if err != nil {
		return nil, err
	}

	return t, nil
}

// ListTemplateVersionFromTemplateIDs get a list of entity.TemplateVersion by condition from database
func (m *ModelTemplateVersion) ListTemplateVersionFromTemplateIDs(ctx context.Context, projectCode string,
	ids []entity.TemplateID) []*entity.TemplateVersion {
	result := make([]*entity.TemplateVersion, 0)
	eg := errgroup.Group{}
	eg.SetLimit(10)
	mux := sync.Mutex{}
	for _, v := range ids {
		id := v
		eg.Go(func() error {
			tv, err := m.GetTemplateVersionByNameVersion(ctx, projectCode, id.TemplateSpace, id.TemplateName,
				id.TemplateVersion)
			if err != nil {
				log.Error(ctx, "get template version failed, %s", err.Error())
				return nil
			}
			mux.Lock()
			defer mux.Unlock()
			result = append(result, tv)
			return nil
		})
	}
	_ = eg.Wait()

	return result
}

// GetTemplateVersionByNameVersion get a specific entity.TemplateVersion from database
func (m *ModelTemplateVersion) GetTemplateVersionByNameVersion(ctx context.Context, projectCode, templateSpace,
	templateName, version string) (*entity.TemplateVersion, error) {
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode:   projectCode,
		entity.FieldKeyTemplateSpace: templateSpace,
		entity.FieldKeyTemplateName:  templateName,
		entity.FieldKeyVersion:       version,
	})

	templateVersion := &entity.TemplateVersion{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, templateVersion); err != nil {
		return nil, err
	}

	return templateVersion, nil
}

// CreateTemplateVersion create a new entity.TemplateVersion into database
func (m *ModelTemplateVersion) CreateTemplateVersion(
	ctx context.Context, templateVersion *entity.TemplateVersion) (string, error) {
	if templateVersion == nil {
		return "", fmt.Errorf("can not create empty templateVersion")
	}

	if err := m.ensureTable(ctx); err != nil {
		return "", err
	}

	// 没有id的情况下生成
	now := time.Now()
	if templateVersion.ID.IsZero() {
		templateVersion.ID = primitive.NewObjectIDFromTimestamp(now)
	}

	if templateVersion.CreateAt == 0 {
		templateVersion.CreateAt = now.UTC().Unix()
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{templateVersion}); err != nil {
		return "", err
	}
	return templateVersion.ID.Hex(), nil
}

// UpdateTemplateVersion update an entity.TemplateVersion into database
func (m *ModelTemplateVersion) UpdateTemplateVersion(ctx context.Context, id string, templateVersion entity.M) error {
	if id == "" {
		return fmt.Errorf("can not update with empty id")
	}

	if templateVersion == nil {
		return fmt.Errorf("can not update empty templateVersion")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyObjectID: objectID,
	})
	old := &entity.TemplateVersion{}
	if err = m.db.Table(m.tableName).Find(cond).One(ctx, old); err != nil {
		return err
	}

	if templateVersion[entity.FieldKeyCreateAt] == nil {
		templateVersion.Update(entity.FieldKeyCreateAt, time.Now().UTC().Unix())
	}

	if err := m.db.Table(m.tableName).Update(ctx, cond, operator.M{"$set": templateVersion}); err != nil {
		return err
	}

	return nil
}

// UpdateTemplateVersionBySpecial update a Special entity.TemplateVersion into database
func (m *ModelTemplateVersion) UpdateTemplateVersionBySpecial(
	ctx context.Context, projectCode, templateName, templateSpace string, templateVersion entity.M) error {
	// 模板文件夹名称不能为空
	if templateSpace == "" || projectCode == "" {
		return fmt.Errorf("can not delete with empty templateSpace or projectCode")
	}

	if templateVersion == nil {
		return fmt.Errorf("can not update empty templateVersion")
	}

	operatorM := operator.M{
		entity.FieldKeyProjectCode:   projectCode,
		entity.FieldKeyTemplateSpace: templateSpace,
	}

	// 如果元数据名称没指定，则是根据文件夹来更新；如果指定了，则是更新某个文件夹的某个元数据的所有版本
	if templateName != "" {
		operatorM[entity.FieldKeyTemplateName] = templateName
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

	if templateVersion[entity.FieldKeyUpdateAt] == nil {
		templateVersion.Update(entity.FieldKeyUpdateAt, time.Now().UTC().Unix())
	}

	if _, err = m.db.Table(m.tableName).UpdateMany(ctx, cond, operator.M{"$set": templateVersion}); err != nil {
		return err
	}

	return nil
}

// DeleteTemplateVersion delete a specific entity.TemplateVersion from database
func (m *ModelTemplateVersion) DeleteTemplateVersion(ctx context.Context, id string) error {
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

// DeleteTemplateVersionBySpecial delete a specific entity.TemplateVersion from database
func (m *ModelTemplateVersion) DeleteTemplateVersionBySpecial(
	ctx context.Context, projectCode, templateName, templateSpace string) error {
	// 模板文件夹名称不能为空
	if templateSpace == "" || projectCode == "" {
		return fmt.Errorf("can not delete with empty templateName or projectCode")
	}

	operatorM := operator.M{
		entity.FieldKeyProjectCode:   projectCode,
		entity.FieldKeyTemplateSpace: templateSpace,
	}

	// 如果元数据名称没指定，则是根据文件夹来删除；如果指定了，则是删除某个文件夹的某个元数据的所有版本
	if templateName != "" {
		operatorM[entity.FieldKeyTemplateName] = templateName
	}

	cond := operator.NewLeafCondition(operator.Eq, operatorM)

	if _, err := m.db.Table(m.tableName).Delete(ctx, cond); err != nil {
		return err
	}

	return nil
}
