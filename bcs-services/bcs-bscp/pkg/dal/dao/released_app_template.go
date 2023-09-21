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
	"fmt"

	rawgen "gorm.io/gen"

	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/search"
	"bscp.io/pkg/types"
)

// ReleasedAppTemplate supplies all the released app template related operations.
type ReleasedAppTemplate interface {
	// BulkCreateWithTx bulk create released template config items.
	BulkCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, items []*table.ReleasedAppTemplate) error
	// List released app templates with options.
	List(kit *kit.Kit, bizID, appID, releaseID uint32, s search.Searcher, opt *types.BasePage) (
		[]*table.ReleasedAppTemplate, int64, error)
	// GetReleasedLately get released templates lately
	GetReleasedLately(kit *kit.Kit, bizID, appID uint32) ([]*table.ReleasedAppTemplate, error)
}

var _ ReleasedAppTemplate = new(releasedAppTemplateDao)

type releasedAppTemplateDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// BulkCreateWithTx bulk create released template config items.
func (dao *releasedAppTemplateDao) BulkCreateWithTx(
	kit *kit.Kit, tx *gen.QueryTx, items []*table.ReleasedAppTemplate) error {
	if len(items) == 0 {
		return nil
	}

	// validate released config item field.
	for _, item := range items {
		if err := item.ValidateCreate(); err != nil {
			return err
		}
	}

	// generate released config items id.
	ids, err := dao.idGen.Batch(kit, table.Name(items[0].TableName()), len(items))
	if err != nil {
		return err
	}

	start := 0
	for _, item := range items {
		item.ID = ids[start]
		start++
	}
	batchSize := 100

	q := tx.ReleasedAppTemplate.WithContext(kit.Ctx)
	if err := q.CreateInBatches(items, batchSize); err != nil {
		return fmt.Errorf("create released template config items in batch failed, err: %v", err)
	}

	ad := dao.auditDao.DecoratorV2(kit, items[0].Attachment.BizID).PrepareCreate(table.RatiList(items))
	if err := ad.Do(tx.Query); err != nil {
		return err
	}

	return nil
}

// List released app templates with options.
func (dao *releasedAppTemplateDao) List(kit *kit.Kit, bizID, appID, releaseID uint32, s search.Searcher,
	opt *types.BasePage) (
	[]*table.ReleasedAppTemplate, int64, error) {
	m := dao.genQ.ReleasedAppTemplate
	q := dao.genQ.ReleasedAppTemplate.WithContext(kit.Ctx)

	var conds []rawgen.Condition
	// add search condition
	if s != nil {
		exprs := s.SearchExprs(dao.genQ)
		if len(exprs) > 0 {
			var do gen.IReleasedAppTemplateDo
			for i := range exprs {
				if i == 0 {
					do = q.Where(exprs[i])
				}
				do = do.Or(exprs[i])
			}
			conds = append(conds, do)
		}
	}

	d := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.ReleaseID.Eq(releaseID)).
		Where(conds...)
	if opt.All {
		result, err := d.Find()
		if err != nil {
			return nil, 0, err
		}
		return result, int64(len(result)), err
	}

	return d.FindByPage(opt.Offset(), opt.LimitInt())
}

// GetReleasedLately get released templates lately
func (dao *releasedAppTemplateDao) GetReleasedLately(kit *kit.Kit, bizID, appId uint32) (
	[]*table.ReleasedAppTemplate, error) {
	m := dao.genQ.ReleasedAppTemplate
	q := dao.genQ.ReleasedAppTemplate.WithContext(kit.Ctx)

	query := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appId))
	subQuery := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appId)).Order(m.ReleaseID.Desc()).Limit(1).Select(m.ReleaseID)
	return query.Where(q.Columns(m.ReleaseID).Eq(subQuery)).Find()
}
