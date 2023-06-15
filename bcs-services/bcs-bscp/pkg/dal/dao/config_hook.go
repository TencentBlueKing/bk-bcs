/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package dao

import (
	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
)

// ConfigHook supplies all the ConfigHook related operations.
type ConfigHook interface {
	// Create one ConfigHook instance.
	Create(kit *kit.Kit, g *table.ConfigHook) (uint32, error)
	// Update one ConfigHook info.
	Update(kit *kit.Kit, g *table.ConfigHook) error
	// GetByAppID get configHook by name.
	GetByAppID(kit *kit.Kit, bizID, appID uint32) (*table.ConfigHook, error)
}

var _ ConfigHook = new(configHookDao)

type configHookDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// Create one ConfigHook instance.
func (dao *configHookDao) Create(kit *kit.Kit, g *table.ConfigHook) (uint32, error) {
	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	// generate a ConfigHook id and update to ConfigHook.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)

	// 多个使用事务处理
	createTx := func(tx *gen.Query) error {
		q := tx.ConfigHook.WithContext(kit.Ctx)
		if e := q.Create(g); e != nil {
			return e
		}

		if e := ad.Do(tx); e != nil {
			return e
		}

		return nil
	}
	if e := dao.genQ.Transaction(createTx); e != nil {
		return 0, e
	}

	return g.ID, nil
}

// Update one TemplateSpace's info.
func (dao *configHookDao) Update(kit *kit.Kit, g *table.ConfigHook) error {
	if err := g.ValidateUpdate(); err != nil {
		return err
	}

	m := dao.genQ.ConfigHook

	// 更新操作, 获取当前记录做审计
	q := dao.genQ.ConfigHook.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.AppID.Eq(g.Attachment.AppID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, oldOne)

	// 多个使用事务处理
	updateTx := func(tx *gen.Query) error {
		q = tx.ConfigHook.WithContext(kit.Ctx)
		if _, e := q.Where(m.AppID.Eq(g.Attachment.AppID), m.BizID.Eq(g.Attachment.BizID)).
			Select(m.PreHookID, m.PreHookReleaseID, m.PostHookID, m.PostHookReleaseID, m.Reviser).Updates(g); e != nil {
			return e
		}

		if e := ad.Do(tx); e != nil {
			return e
		}
		return nil
	}
	if e := dao.genQ.Transaction(updateTx); e != nil {
		return e
	}

	return nil
}

// GetByAppID get configHook by appID.
func (dao *configHookDao) GetByAppID(kit *kit.Kit, bizID, appID uint32) (*table.ConfigHook, error) {

	m := dao.genQ.ConfigHook
	q := dao.genQ.ConfigHook.WithContext(kit.Ctx)

	hook, err := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appID)).Take()
	if err != nil {
		return nil, err
	}
	return hook, nil

}
