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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// ClientQuery supplies all the client query related operations.
type ClientQuery interface {
	// Create one client query instance
	Create(kit *kit.Kit, data *table.ClientQuery) (uint32, error)
	// Update one client query
	Update(kit *kit.Kit, data *table.ClientQuery) error
	// List client query with options.
	List(kit *kit.Kit, bizID, appID uint32, creator, search_type string, opt *types.BasePage) (
		[]*table.ClientQuery, int64, error)
	// Delete ..
	Delete(kit *kit.Kit, data *table.ClientQuery) error
}

var _ ClientQuery = new(clientQueryDao)

type clientQueryDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// Delete ..
func (dao *clientQueryDao) Delete(kit *kit.Kit, data *table.ClientQuery) error {
	if data == nil {
		return errors.New("client searche is nil")
	}

	if err := data.ValidateDelete(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := dao.genQ.ClientQuery
	q := dao.genQ.ClientQuery.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(data.ID), m.BizID.Eq(data.Attachment.BizID), m.AppID.Eq(data.Attachment.AppID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, data.Attachment.BizID).PrepareDelete(oldOne)

	// 多个使用事务处理
	deleteTx := func(tx *gen.Query) error {
		q = tx.ClientQuery.WithContext(kit.Ctx)
		if _, e := q.Where(m.BizID.Eq(data.Attachment.BizID), m.ID.Eq(data.ID)).Delete(data); e != nil {
			return e
		}

		if e := ad.Do(tx); e != nil {
			return e
		}
		return nil
	}
	if e := dao.genQ.Transaction(deleteTx); e != nil {
		return e
	}

	return nil
}

// Create one client query instance
func (dao *clientQueryDao) Create(kit *kit.Kit, data *table.ClientQuery) (uint32, error) {
	if data == nil {
		return 0, errors.New("client searche is nil")
	}

	if err := data.ValidateCreate(); err != nil {
		return 0, err
	}

	id, err := dao.idGen.One(kit, table.Name(data.TableName()))
	if err != nil {
		return 0, err
	}
	data.ID = id

	ad := dao.auditDao.DecoratorV2(kit, data.Attachment.BizID).PrepareCreate(data)

	createTx := func(tx *gen.Query) error {
		q := tx.ClientQuery.WithContext(kit.Ctx)
		if err = q.Create(data); err != nil {
			return err
		}
		if err = ad.Do(tx); err != nil {
			return err
		}

		return nil
	}
	if err = dao.genQ.Transaction(createTx); err != nil {
		return 0, err
	}

	return id, nil
}

// List client query with options.
func (dao *clientQueryDao) List(kit *kit.Kit, bizID uint32, appID uint32, creator, search_type string,
	opt *types.BasePage) ([]*table.ClientQuery, int64, error) {

	m := dao.genQ.ClientQuery
	q := dao.genQ.ClientQuery.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.Creator.Eq(creator))

	if len(search_type) != 0 {
		q = q.Where(m.SearchType.Eq(search_type))
		if search_type == string(table.Common) {
			q = q.Or(m.BizID.Eq(0), m.AppID.Eq(0), m.Creator.Eq("system"))
		}
	}

	d := q.Order(m.UpdatedAt.Desc())
	if opt.All {
		result, err := d.Find()
		if err != nil {
			return nil, 0, err
		}
		return result, int64(len(result)), err
	}
	return d.FindByPage(opt.Offset(), opt.LimitInt())
}

// Update one client query
func (dao *clientQueryDao) Update(kit *kit.Kit, data *table.ClientQuery) error {
	if data == nil {
		return errors.New("client searche is nil")
	}

	if err := data.ValidateUpdate(); err != nil {
		return err
	}

	m := dao.genQ.ClientQuery
	q := dao.genQ.ClientQuery.WithContext(kit.Ctx)

	if _, e := q.Where(m.BizID.Eq(data.Attachment.BizID), m.ID.Eq(data.ID)).Updates(data); e != nil {
		return e
	}

	return nil
}
