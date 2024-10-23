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
	"fmt"
	"time"

	rawgen "gorm.io/gen"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// App supplies all the app related operations.
type App interface {
	// Create one app instance
	Create(kit *kit.Kit, app *table.App) (uint32, error)
	// Update one app's info
	Update(kit *kit.Kit, app *table.App) error
	// get app with id.
	Get(kit *kit.Kit, bizID, appID uint32) (*table.App, error)
	// get app only with id.
	GetByID(kit *kit.Kit, appID uint32) (*table.App, error)
	// get app by name.
	GetByName(kit *kit.Kit, bizID uint32, name string) (*table.App, error)
	// List apps with options.
	List(kit *kit.Kit, bizList []uint32, name, operator string, opt *types.BasePage) ([]*table.App, int64, error)
	// ListAppsByGroupID list apps by group id.
	ListAppsByGroupID(kit *kit.Kit, groupID, bizID uint32) ([]*table.App, error)
	// ListAppsByIDs list apps by app ids.
	ListAppsByIDs(kit *kit.Kit, ids []uint32) ([]*table.App, error)
	// DeleteWithTx delete one app instance with transaction.
	DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, app *table.App) error
	// ListAppMetaForCache list app's basic meta info.
	ListAppMetaForCache(kt *kit.Kit, bizID uint32, appID []uint32) (map[ /*appID*/ uint32]*types.AppCacheMeta, error)
	// GetByAlias 通过Alisa 查询
	GetByAlias(kit *kit.Kit, bizID uint32, alias string) (*table.App, error)
	// BatchUpdateLastConsumedTime 批量更新最后一次拉取时间
	BatchUpdateLastConsumedTime(kit *kit.Kit, bizID uint32, appIDs []uint32) error
}

var _ App = new(appDao)

type appDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
	event    Event
}

// BatchUpdateLastConsumedTime 批量更新最后一次拉取时间
func (dao *appDao) BatchUpdateLastConsumedTime(kit *kit.Kit, bizID uint32, appIDs []uint32) error {

	m := dao.genQ.App

	_, err := dao.genQ.App.WithContext(kit.Ctx).
		Where(m.BizID.Eq(bizID), m.ID.In(appIDs...)).
		Update(m.LastConsumedTime, time.Now().UTC())
	if err != nil {
		return err
	}

	return nil
}

// List app's detail info with the filter's expression.
func (dao *appDao) List(kit *kit.Kit, bizList []uint32, name, operator string, opt *types.BasePage) (
	[]*table.App, int64, error) {
	m := dao.genQ.App
	q := dao.genQ.App.WithContext(kit.Ctx)

	var conds []rawgen.Condition
	// 当len(bizList) > 1时，适用于导航查询场景
	conds = append(conds, m.BizID.In(bizList...))
	if operator != "" {
		conds = append(conds, m.Creator.Eq(operator))
	}
	if name != "" {
		// 按名称模糊搜索
		conds = append(conds, m.Name.Regexp("(?i)"+name)) // nolint: goconst
	}

	var (
		result []*table.App
		count  int64
		err    error
	)

	if opt.All {
		result, err = q.Where(conds...).Find()
		count = int64(len(result))
	} else {
		result, count, err = q.Where(conds...).FindByPage(opt.Offset(), opt.LimitInt())
	}
	if err != nil {
		return nil, 0, err
	}

	return result, count, nil
}

// ListAppsByGroupID list apps by group id.
func (dao *appDao) ListAppsByGroupID(kit *kit.Kit, groupID, bizID uint32) ([]*table.App, error) {
	if bizID == 0 {
		return nil, errors.New("biz id is 0")
	}
	if groupID == 0 {
		return nil, errors.New("group id is 0")
	}

	gm := dao.genQ.Group
	gq := dao.genQ.Group.WithContext(kit.Ctx)
	group, err := gq.Where(gm.BizID.Eq(bizID), gm.ID.Eq(groupID)).Take()
	if err != nil {
		return nil, fmt.Errorf("get group failed, err: %v", err)
	}

	bm := dao.genQ.GroupAppBind
	bq := dao.genQ.GroupAppBind.WithContext(kit.Ctx)
	am := dao.genQ.App
	aq := dao.genQ.App.WithContext(kit.Ctx)
	var conds []rawgen.Condition
	conds = append(conds, am.BizID.Eq(bizID))

	if !group.Spec.Public {
		conds = append(conds, aq.Columns(am.ID).In(bq.Select(bm.AppID).Where(bm.GroupID.Eq(groupID))))
	}

	result, err := aq.Where(conds...).Find()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListAppsByIDs list apps by app ids.
func (dao *appDao) ListAppsByIDs(kit *kit.Kit, ids []uint32) ([]*table.App, error) {
	m := dao.genQ.App
	q := dao.genQ.App.WithContext(kit.Ctx)
	result, err := q.Where(m.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Create one app instance
func (dao *appDao) Create(kit *kit.Kit, g *table.App) (uint32, error) {
	if g == nil {
		return 0, errf.Errorf(errf.InvalidArgument, i18n.T(kit, "app is nil"))
	}

	if err := g.ValidateCreate(kit); err != nil {
		return 0, err
	}

	// generate an app id and update to g.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	ad := dao.auditDao.DecoratorV2(kit, g.BizID).PrepareCreate(g)
	eDecorator := dao.event.Eventf(kit)

	// 多个使用事务处理
	createTx := func(tx *gen.Query) error {
		q := tx.App.WithContext(kit.Ctx)
		if err = q.Create(g); err != nil {
			return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "create data failed, err: %v", err))
		}

		if err = ad.Do(tx); err != nil {
			logs.Errorf("execution of transactions failed, err: %v", err)
			return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "create app failed, err: %v", err))
		}

		// fire the event with txn to ensure the if save the event failed then the business logic is failed anyway.
		one := types.Event{
			Spec: &table.EventSpec{
				Resource:   table.Application,
				ResourceID: g.ID,
				OpType:     table.InsertOp,
			},
			Attachment: &table.EventAttachment{BizID: g.BizID, AppID: g.ID},
			Revision:   &table.CreatedRevision{Creator: kit.User},
		}
		if err = eDecorator.Fire(one); err != nil {
			logs.Errorf("fire create app: %s event failed, err: %v, rid: %s", g.ID, err, kit.Rid)
			return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "create app failed, err: %v", err))
		}

		return nil
	}
	err = dao.genQ.Transaction(createTx)

	eDecorator.Finalizer(err)

	if err != nil {
		logs.Errorf("transaction processing failed %s", err)
		return 0, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "create app failed, err: %v", err))
	}

	return id, nil
}

// Update an app instance.
func (dao *appDao) Update(kit *kit.Kit, g *table.App) error {
	if g == nil {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "app is nil"))
	}

	oldOne, err := dao.Get(kit, g.BizID, g.ID)
	if err != nil {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "update app failed, err: %s", err))
	}

	if err = g.ValidateUpdate(kit, oldOne.Spec.ConfigType); err != nil {
		return err
	}

	// 更新操作, 获取当前记录做审计
	m := dao.genQ.App
	q := dao.genQ.App.WithContext(kit.Ctx)
	ad := dao.auditDao.DecoratorV2(kit, g.BizID).PrepareUpdate(g, oldOne)
	eDecorator := dao.event.Eventf(kit)

	// 多个使用事务处理
	updateTx := func(tx *gen.Query) error {
		q = tx.App.WithContext(kit.Ctx)
		if _, err = q.Where(m.BizID.Eq(g.BizID), m.ID.Eq(g.ID)).
			Select(m.Memo, m.Alias_, m.DataType, m.Reviser, m.UpdatedAt, m.IsApprove, m.ApproveType,
				m.Approver).Updates(g); err != nil {
			return err
		}

		if err = ad.Do(tx); err != nil {
			return err
		}

		// fire the event with txn to ensure the if save the event failed then the business logic is failed anyway.
		one := types.Event{
			Spec: &table.EventSpec{
				Resource:   table.Application,
				ResourceID: g.ID,
				OpType:     table.UpdateOp,
			},
			Attachment: &table.EventAttachment{BizID: g.BizID, AppID: g.ID},
			Revision:   &table.CreatedRevision{Creator: kit.User},
		}
		if err = eDecorator.Fire(one); err != nil {
			logs.Errorf("fire update app: %s event failed, err: %v, rid: %s", g.ID, err, kit.Rid)
			return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "update app failed, err: %s", err))
		}
		return nil
	}
	err = dao.genQ.Transaction(updateTx)

	eDecorator.Finalizer(err)

	if err != nil {
		return err
	}

	return nil
}

// DeleteWithTx delete one app instance with transaction.
func (dao *appDao) DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.App) error {
	if g == nil {
		return errors.New("app is nil")
	}

	if err := g.ValidateDelete(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := tx.App
	q := tx.App.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.BizID).PrepareDelete(oldOne)
	if err = ad.Do(tx.Query); err != nil {
		return err
	}

	if _, err = q.Where(m.BizID.Eq(g.BizID)).Delete(g); err != nil {
		return err
	}

	// archived this deleted app to archive table.
	if err = dao.archiveApp(kit, tx, oldOne); err != nil {
		return err
	}

	// fire the event with txn to ensure the if save the event failed then the business logic is failed anyway.
	one := types.Event{
		Spec: &table.EventSpec{
			Resource:   table.Application,
			ResourceID: g.ID,
			OpType:     table.DeleteOp,
		},
		Attachment: &table.EventAttachment{BizID: g.BizID, AppID: g.ID},
		Revision:   &table.CreatedRevision{Creator: kit.User},
	}
	eDecorator := dao.event.Eventf(kit)
	if err = eDecorator.FireWithTx(tx, one); err != nil {
		logs.Errorf("fire delete app: %s event failed, err: %v, rid: %s", g.ID, err, kit.Rid)
		return errors.New("fire event failed, " + err.Error()) // nolint: goconst
	}

	return nil
}

// Get 获取单个app详情
func (dao *appDao) Get(kit *kit.Kit, bizID uint32, appID uint32) (*table.App, error) {
	m := dao.genQ.App
	q := dao.genQ.App.WithContext(kit.Ctx)
	detail, err := q.Where(m.ID.Eq(appID), m.BizID.Eq(bizID)).Take()
	if err != nil {
		return nil, err
	}
	return detail, nil
}

// GetByID 通过 AppId 查询
func (dao *appDao) GetByID(kit *kit.Kit, appID uint32) (*table.App, error) {
	m := dao.genQ.App
	q := dao.genQ.App.WithContext(kit.Ctx)

	app, err := q.Where(m.ID.Eq(appID)).Take()
	if err != nil {
		return nil, err
	}

	return app, nil
}

// GetByName 通过 name 查询
func (dao *appDao) GetByName(kit *kit.Kit, bizID uint32, name string) (*table.App, error) {
	m := dao.genQ.App
	q := dao.genQ.App.WithContext(kit.Ctx)

	app, err := q.Where(m.BizID.Eq(bizID), m.Name.Eq(name)).Take()
	if err != nil {
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "get app failed, err: %v", err))
	}

	return app, nil
}

// GetByAlias 通过Alisa 查询
func (dao *appDao) GetByAlias(kit *kit.Kit, bizID uint32, alias string) (*table.App, error) {
	m := dao.genQ.App
	q := dao.genQ.App.WithContext(kit.Ctx)

	app, err := q.Where(m.BizID.Eq(bizID), m.Alias_.Eq(alias)).Take()
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (dao *appDao) archiveApp(kit *kit.Kit, tx *gen.QueryTx, g *table.App) error {
	id, err := dao.idGen.One(kit, table.ArchivedAppTable)
	if err != nil {
		return err
	}

	archivedApp := &table.ArchivedApp{
		ID:    id,
		AppID: g.ID,
		BizID: g.BizID,
	}

	q := tx.ArchivedApp.WithContext(kit.Ctx)
	if err = q.Create(archivedApp); err != nil {
		return fmt.Errorf("archived delete app failed, err: %v", err)
	}

	return nil
}

// ListAppMetaForCache list app's basic meta info.
func (dao *appDao) ListAppMetaForCache(kit *kit.Kit, bizID uint32, appIDs []uint32) (
	map[uint32]*types.AppCacheMeta, error) {
	if bizID <= 0 || len(appIDs) == 0 {
		return nil, errors.New("invalid biz id or app id list")
	}

	m := dao.genQ.App
	q := dao.genQ.App.WithContext(kit.Ctx)

	result, err := q.Select(m.ID, m.Name, m.ConfigType).
		Where(m.BizID.Eq(bizID), m.ID.In(appIDs...)).Find()
	if err != nil {
		return nil, err
	}

	meta := make(map[uint32]*types.AppCacheMeta)
	for _, one := range result {
		meta[one.ID] = &types.AppCacheMeta{
			Name:       one.Spec.Name,
			ConfigType: one.Spec.ConfigType,
		}
	}

	return meta, nil
}
