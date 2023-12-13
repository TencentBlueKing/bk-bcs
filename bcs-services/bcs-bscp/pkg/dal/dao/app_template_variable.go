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

package dao

import (
	"errors"

	"gorm.io/gorm"

	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
)

// AppTemplateVariable supplies all the app template variable related operations.
type AppTemplateVariable interface {
	// Upsert create or update one template variable instance.
	Upsert(kit *kit.Kit, appVar *table.AppTemplateVariable) error
	// UpsertWithTx create or update one template variable instance with transaction.
	UpsertWithTx(kit *kit.Kit, tx *gen.QueryTx, appVar *table.AppTemplateVariable) error
	// Get gets app template variables
	Get(kit *kit.Kit, bizID, appID uint32) (*table.AppTemplateVariable, error)
	// ListVariables lists all variables in app template variable
	ListVariables(kit *kit.Kit, bizID, appID uint32) ([]*table.TemplateVariableSpec, error)
}

var _ AppTemplateVariable = new(appTemplateVariableDao)

type appTemplateVariableDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// Upsert create or update one template variable instance.
func (dao *appTemplateVariableDao) Upsert(kit *kit.Kit, g *table.AppTemplateVariable) error {
	if err := g.ValidateUpsert(kit); err != nil {
		return err
	}

	m := dao.genQ.AppTemplateVariable
	q := dao.genQ.AppTemplateVariable.WithContext(kit.Ctx)
	old, findErr := q.Where(m.BizID.Eq(g.Attachment.BizID), m.AppID.Eq(g.Attachment.AppID)).Take()

	// 多个使用事务处理
	upsertTx := func(tx *gen.Query) error {
		var ad AuditDo
		// if old exists, update it.
		if findErr == nil {
			g.ID = old.ID
			if _, err := tx.AppTemplateVariable.WithContext(kit.Ctx).
				Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).
				Select(m.Variables, m.Reviser).
				Updates(g); err != nil {
				return err
			}
			ad = dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, old)
		} else if errors.Is(findErr, gorm.ErrRecordNotFound) {
			// if old not exists, create it.
			id, err := dao.idGen.One(kit, table.Name(g.TableName()))
			if err != nil {
				return err
			}
			g.ID = id
			if err := tx.AppTemplateVariable.WithContext(kit.Ctx).Create(g); err != nil {
				return err
			}
			ad = dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)
		}

		return ad.Do(tx)
	}
	if err := dao.genQ.Transaction(upsertTx); err != nil {
		return err
	}

	return nil
}

// UpsertWithTx create or update one template variable instance with transaction.
func (dao *appTemplateVariableDao) UpsertWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.AppTemplateVariable) error {
	if err := g.ValidateUpsert(kit); err != nil {
		return err
	}

	m := dao.genQ.AppTemplateVariable
	q := dao.genQ.AppTemplateVariable.WithContext(kit.Ctx)
	old, findErr := q.Where(m.BizID.Eq(g.Attachment.BizID), m.AppID.Eq(g.Attachment.AppID)).Take()

	var ad AuditDo
	// if old exists, update it.
	if findErr == nil {
		g.ID = old.ID
		if _, err := tx.AppTemplateVariable.WithContext(kit.Ctx).
			Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).
			Select(m.Variables, m.Reviser).
			Updates(g); err != nil {
			return err
		}
		ad = dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, old)
	} else if errors.Is(findErr, gorm.ErrRecordNotFound) {
		// if old not exists, create it.
		id, err := dao.idGen.One(kit, table.Name(g.TableName()))
		if err != nil {
			return err
		}
		g.ID = id
		if err := tx.AppTemplateVariable.WithContext(kit.Ctx).Create(g); err != nil {
			return err
		}
		ad = dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)
	}

	return ad.Do(tx.Query)
}

// Get gets app template variables.
func (dao *appTemplateVariableDao) Get(kit *kit.Kit, bizID, appID uint32) (*table.AppTemplateVariable, error) {
	m := dao.genQ.AppTemplateVariable
	q := dao.genQ.AppTemplateVariable.WithContext(kit.Ctx)
	appVars, err := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appID)).Find()
	if err != nil {
		return nil, err
	}
	if len(appVars) == 0 {
		return nil, nil
	}
	return appVars[0], nil
}

// ListVariables lists all variables in app template variable
func (dao *appTemplateVariableDao) ListVariables(kit *kit.Kit, bizID, appID uint32) (
	[]*table.TemplateVariableSpec, error) {
	m := dao.genQ.AppTemplateVariable
	q := dao.genQ.AppTemplateVariable.WithContext(kit.Ctx)
	appVars, err := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appID)).Find()
	if err != nil {
		return nil, err
	}
	if len(appVars) == 0 {
		return []*table.TemplateVariableSpec{}, nil
	}
	return appVars[0].Spec.Variables, nil
}
