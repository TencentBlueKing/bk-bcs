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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/utils"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/search"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// ReleasedAppTemplate supplies all the released app template related operations.
type ReleasedAppTemplate interface {
	// BulkCreateWithTx bulk create released template config items.
	BulkCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, items []*table.ReleasedAppTemplate) error
	// Get released app template.
	Get(kit *kit.Kit, bizID, appID, releaseID, tmplRevisionID uint32) (*table.ReleasedAppTemplate, error)
	// List released app templates with options.
	List(kit *kit.Kit, bizID, appID, releaseID uint32, s search.Searcher, opt *types.BasePage, searchValue string) (
		[]*table.ReleasedAppTemplate, int64, error)
	// GetReleasedLately get released templates lately
	GetReleasedLately(kit *kit.Kit, bizID, appID uint32) ([]*table.ReleasedAppTemplate, error)
	// BatchDeleteByAppIDWithTx batch delete by app id with transaction.
	BatchDeleteByAppIDWithTx(kit *kit.Kit, tx *gen.QueryTx, appID, bizID uint32) error
	// BatchDeleteByReleaseIDWithTx batch delete by release id with transaction.
	BatchDeleteByReleaseIDWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, appID, releaseID uint32) error
	// ListAllCISigns lists all released template ci signatures of one biz, and only belongs to existing apps
	ListAllCISigns(kit *kit.Kit, bizID uint32) ([]string, error)
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

// Get released app template.
func (dao *releasedAppTemplateDao) Get(kit *kit.Kit, bizID, appID, releaseID,
	tmplRevisionID uint32) (*table.ReleasedAppTemplate, error) {
	m := dao.genQ.ReleasedAppTemplate
	return m.WithContext(kit.Ctx).
		Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.ReleaseID.Eq(releaseID), m.TemplateRevisionID.Eq(tmplRevisionID)).
		Take()
}

// List released app templates with options.
func (dao *releasedAppTemplateDao) List(kit *kit.Kit, bizID, appID, releaseID uint32, s search.Searcher,
	opt *types.BasePage, searchValue string) (
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
			do = do.Or(utils.RawCond(`CASE WHEN RIGHT(path, 1) = '/' THEN CONCAT(path,name)
			ELSE CONCAT_WS('/', path, name) END LIKE ?`, "%"+searchValue+"%"))
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
	subQuery := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appId)).
		Order(m.ReleaseID.Desc(), utils.NewCustomExpr("CASE WHEN RIGHT(path, 1) = '/' THEN CONCAT(path,`name`) ELSE "+
			"CONCAT_WS('/', path, `name`) END", nil)).
		Limit(1).
		Select(m.ReleaseID)
	return query.Where(q.Columns(m.ReleaseID).Eq(subQuery)).Find()
}

// BatchDeleteByAppIDWithTx batch delete by app id with transaction.
func (dao *releasedAppTemplateDao) BatchDeleteByAppIDWithTx(kit *kit.Kit, tx *gen.QueryTx, appID, bizID uint32) error {

	if bizID == 0 {
		return errf.New(errf.InvalidParameter, "bizID is 0")
	}
	if appID == 0 {
		return errf.New(errf.InvalidParameter, "appID is 0")
	}

	m := tx.ReleasedAppTemplate

	_, err := m.WithContext(kit.Ctx).Where(m.AppID.Eq(appID), m.BizID.Eq(bizID)).Delete()
	return err
}

// BatchDeleteByReleaseIDWithTx batch delete by release id with transaction.
func (dao *releasedAppTemplateDao) BatchDeleteByReleaseIDWithTx(kit *kit.Kit, tx *gen.QueryTx,
	bizID, appID, releaseID uint32) error {

	if bizID == 0 {
		return errf.New(errf.InvalidParameter, "bizID is 0")
	}
	if appID == 0 {
		return errf.New(errf.InvalidParameter, "appID is 0")
	}
	if releaseID == 0 {
		return errf.New(errf.InvalidParameter, "releaseID is 0")
	}

	m := tx.ReleasedAppTemplate

	_, err := m.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.ReleaseID.Eq(releaseID)).Delete()
	return err
}

// ListAllCISigns lists all released template ci signatures of one biz, and only belongs to existing apps
func (dao *releasedAppTemplateDao) ListAllCISigns(kit *kit.Kit, bizID uint32) ([]string, error) {
	am := dao.genQ.App
	aq := dao.genQ.App.WithContext(kit.Ctx)
	var appIDs []uint32
	if err := aq.Select(am.ID.Distinct()).
		Where(am.BizID.Eq(bizID)).
		Pluck(am.ID, &appIDs); err != nil {
		return nil, err
	}

	m := dao.genQ.ReleasedAppTemplate
	q := dao.genQ.ReleasedAppTemplate.WithContext(kit.Ctx)
	var signs []string
	if err := q.Select(m.Signature.Distinct()).
		Where(m.BizID.Eq(bizID), m.AppID.In(appIDs...)).
		Pluck(m.Signature, &signs); err != nil {
		return nil, err
	}

	return signs, nil
}
