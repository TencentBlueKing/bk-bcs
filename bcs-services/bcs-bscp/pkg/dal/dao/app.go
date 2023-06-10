/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package dao

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	rawgen "gorm.io/gen"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"

	"github.com/jmoiron/sqlx"
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
	// Delete one app instance.
	Delete(kit *kit.Kit, app *table.App) error
	// ListAppMetaForCache list app's basic meta info.
	ListAppMetaForCache(kt *kit.Kit, bizID uint32, appID []uint32) (map[ /*appID*/ uint32]*types.AppCacheMeta, error)
}

var _ App = new(appDao)

type appDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao

	orm   orm.Interface
	sd    *sharding.Sharding
	event Event
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
		conds = append(conds, m.Name.Regexp("(?i)"+name))
	}

	result, count, err := q.Where(conds...).FindByPage(opt.Offset(), opt.LimitInt())
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

	gM := dao.genQ.Group
	gQ := dao.genQ.Group.WithContext(kit.Ctx)
	group, err := gQ.Where(gM.BizID.Eq(bizID), gM.ID.Eq(groupID)).Take()
	if err != nil {
		return nil, fmt.Errorf("get group failed, err: %v", err)
	}

	bM := dao.genQ.GroupAppBind
	bQ := dao.genQ.GroupAppBind.WithContext(kit.Ctx)
	aM := dao.genQ.App
	aQ := dao.genQ.App.WithContext(kit.Ctx)
	var conds []rawgen.Condition
	conds = append(conds, aM.BizID.Eq(bizID))

	if !group.Spec.Public {
		conds = append(conds, aQ.Columns(aM.ID).In(bQ.Select(bM.AppID).Where(bM.GroupID.Eq(groupID))))
	}

	result, err := aQ.Where(conds...).Find()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Create one app instance
func (dao *appDao) Create(kit *kit.Kit, g *table.App) (uint32, error) {
	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	// generate an app id and update to app.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	ad := dao.auditDao.DecoratorV2(kit, g.BizID).PrepareCreate(g)

	// 多个使用事务处理
	createTx := func(tx *gen.Query) error {
		q := tx.App.WithContext(kit.Ctx)
		if err := q.Create(g); err != nil {
			return err
		}

		eQ := tx.Event.WithContext(kit.Ctx)
		if err := eQ.Create(g); err != nil {
			return err
		}

		if err := ad.Do(tx); err != nil {
			return err
		}

		return nil
	}
	if err := dao.genQ.Transaction(createTx); err != nil {
		return 0, nil
	}

	//return g.ID, nil

	if app == nil {
		return 0, errf.New(errf.InvalidParameter, "app is nil")
	}

	if err := app.ValidateCreate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	// generate an app id and update to app.
	id, err := ap.idGen.One(kit, table.AppTable)
	if err != nil {
		return 0, err
	}

	app.ID = id

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", table.AppTable.Name(),
		" (", table.AppColumns.ColumnExpr(), ") ", "VALUES(", table.AppColumns.ColonNameExpr(), ")")
	sql := filter.SqlJoint(sqlSentence)
	eDecorator := dao.event.Eventf(kit)
	err = dao.sd.ShardingOne(app.BizID).AutoTxn(kit, func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
		if err := dao.orm.Txn(txn).Insert(kit.Ctx, sql, app); err != nil {
			return err
		}

		// audit this to be create app details.
		au := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
		if err = dao.auditDao.Decorator(kit, app.BizID, enumor.App).AuditCreate(app, au); err != nil {
			return fmt.Errorf("audit create app failed, err: %v", err)
		}

		// fire the event with txn to ensure the if save the event failed then the business logic is failed anyway.
		one := types.Event{
			Spec: &table.EventSpec{
				Resource:   table.Application,
				ResourceID: app.ID,
				OpType:     table.InsertOp,
			},
			Attachment: &table.EventAttachment{BizID: app.BizID, AppID: app.ID},
			Revision:   &table.CreatedRevision{Creator: kit.User, CreatedAt: time.Now()},
		}
		if err = eDecorator.Fire(one); err != nil {
			logs.Errorf("fire create app: %s event failed, err: %v, rid: %s", app.ID, err, kit.Rid)
			return errf.New(errf.DBOpFailed, "fire event failed, "+err.Error())
		}
		return nil
	})

	eDecorator.Finalizer(err)

	if err != nil {
		logs.Errorf("create app, but do auto txn failed, err: %v, rid: %s", err, kit.Rid)
		return 0, fmt.Errorf("create app, but auto run txn failed, err: %v", err)
	}

	return id, nil
}

// Update an app instance.
func (dao *appDao) Update(kit *kit.Kit, app *table.App) error {

	if app == nil {
		return errf.New(errf.InvalidParameter, "app is nil")
	}

	updateApp, err := dao.Get(kit, app.BizID, app.ID)
	if err != nil {
		return fmt.Errorf("get update app failed, err: %v", err)
	}

	if err := app.ValidateUpdate(updateApp.Spec.ConfigType); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	opts := orm.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields("id", "biz_id")
	expr, toUpdate, err := orm.RearrangeSQLDataWithOption(app, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql expr failed, err: %v", err)
	}

	ab := ap.auditDao.Decorator(kit, app.BizID, enumor.App).PrepareUpdate(app)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "UPDATE ", table.AppTable.Name(), " SET ", expr, " WHERE id = ",
		strconv.Itoa(int(app.ID)), " and biz_id = ", strconv.Itoa(int(app.BizID)))
	sql := filter.SqlJoint(sqlSentence)

	eDecorator := ap.event.Eventf(kit)
	err = ap.sd.ShardingOne(app.BizID).AutoTxn(kit, func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
		effected, err := ap.orm.Txn(txn).Update(kit.Ctx, sql, toUpdate)
		if err != nil {
			logs.Errorf("update app: %d failed, err: %v, rid: %v", app.ID, err, kit.Rid)
			return err
		}

		if effected == 0 {
			logs.Errorf("update one app: %d, but record not found, rid: %v", app.ID, kit.Rid)
			return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
		}

		if effected > 1 {
			logs.Errorf("update one app: %d, but got updated app count: %d, rid: %v", app.ID, effected, kit.Rid)
			return fmt.Errorf("matched app count %d is not as excepted", effected)
		}

		// do audit
		if err := ab.Do(&AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}); err != nil {
			return fmt.Errorf("do app update audit failed, err: %v", err)
		}

		// fire the event with txn to ensure the if save the event failed then the business logic is failed anyway.
		one := types.Event{
			Spec: &table.EventSpec{
				Resource:   table.Application,
				ResourceID: app.ID,
				OpType:     table.UpdateOp,
			},
			Attachment: &table.EventAttachment{BizID: app.BizID, AppID: app.ID},
			Revision:   &table.CreatedRevision{Creator: kit.User, CreatedAt: time.Now()},
		}
		if err := eDecorator.Fire(one); err != nil {
			logs.Errorf("fire update app: %s event failed, err: %v, rid: %s", app.ID, err, kit.Rid)
			return errf.New(errf.DBOpFailed, "fire event failed, "+err.Error())
		}
		return nil
	})

	eDecorator.Finalizer(err)

	if err != nil {
		return err
	}

	return nil
}

// Delete an app instance.
func (dao *appDao) Delete(kit *kit.Kit, app *table.App) error {

	if app == nil {
		return errf.New(errf.InvalidParameter, "app is nil")
	}

	if err := app.ValidateDelete(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	ab := ap.auditDao.Decorator(kit, app.BizID, enumor.App).PrepareDelete(app.ID)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "DELETE FROM ", table.AppTable.Name(), " WHERE id = ",
		strconv.Itoa(int(app.ID)), " AND biz_id = ", strconv.Itoa(int(app.BizID)))
	sql := filter.SqlJoint(sqlSentence)

	eDecorator := ap.event.Eventf(kit)
	err := ap.sd.ShardingOne(app.BizID).AutoTxn(kit, func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
		oldApp, err := ap.Get(kit, app.BizID, app.ID)
		if err != nil {
			return fmt.Errorf("get pre app failed, err: %v", err)
		}

		// delete the app at first.
		err = ap.orm.Txn(txn).Delete(kit.Ctx, sql)
		if err != nil {
			return err
		}

		// archived this deleted app to archive table.
		if err := ap.archiveApp(kit, txn, oldApp); err != nil {
			return err
		}

		// audit this delete app details.
		auditOpt := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
		if err := ab.Do(auditOpt); err != nil {
			return fmt.Errorf("audit delete app failed, err: %v", err)
		}

		// fire the event with txn to ensure the if save the event failed then the business logic is failed anyway.
		one := types.Event{
			Spec: &table.EventSpec{
				Resource:   table.Application,
				ResourceID: app.ID,
				OpType:     table.DeleteOp,
			},
			Attachment: &table.EventAttachment{BizID: app.BizID, AppID: app.ID},
			Revision:   &table.CreatedRevision{Creator: kit.User, CreatedAt: time.Now()},
		}
		if err := eDecorator.Fire(one); err != nil {
			logs.Errorf("fire delete app: %s event failed, err: %v, rid: %s", app.ID, err, kit.Rid)
			return errf.New(errf.DBOpFailed, "fire event failed, "+err.Error())
		}

		return nil
	})

	eDecorator.Finalizer(err)

	if err != nil {
		logs.Errorf("delete app: %d failed, err: %v, rid: %v", app.ID, err, kit.Rid)
		return fmt.Errorf("delete app, but run txn failed, err: %v", err)
	}

	return nil
}

func (dao *appDao) Get(kit *kit.Kit, bizID uint32, appID uint32) (*table.App, error) {

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.AppColumns.NamedExpr(), " FROM ",
		table.AppTable.Name(), " WHERE id = ", strconv.Itoa(int(appID)), " AND biz_id = ", strconv.Itoa(int(bizID)))
	sql := filter.SqlJoint(sqlSentence)

	one := new(table.App)
	err := ap.orm.Do(ap.sd.MustSharding(bizID)).Get(kit.Ctx, one, sql)
	if err != nil {
		return nil, fmt.Errorf("get app details failed, err: %v", err)
	}

	return one, nil
}

// GetByID 通过 AppId 查询
func (dao *appDao) GetByID(kit *kit.Kit, appID uint32) (*table.App, error) {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.AppColumns.NamedExpr(), " FROM ", table.AppTable.Name(),
		" WHERE id = ", strconv.Itoa(int(appID)))
	expr := filter.SqlJoint(sqlSentence)
	one := new(table.App)
	err := ap.orm.Do(ap.sd.Admin().DB()).Get(kit.Ctx, one, expr)
	if err != nil {
		return nil, fmt.Errorf("get app details failed, err: %v", err)
	}

	return one, nil
}

// GetByName 通过 name 查询
func (dao *appDao) GetByName(kit *kit.Kit, bizID uint32, name string) (*table.App, error) {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.AppColumns.NamedExpr(), " FROM ", table.AppTable.Name(),
		" WHERE name = '", name, "' AND biz_id = ", strconv.Itoa(int(bizID)))
	expr := filter.SqlJoint(sqlSentence)
	one := new(table.App)
	err := ap.orm.Do(ap.sd.Admin().DB()).Get(kit.Ctx, one, expr)
	if err != nil {
		return nil, fmt.Errorf("get app details failed, err: %v", err)
	}

	return one, nil
}

func getAppMode(kit *kit.Kit, orm orm.Interface, sd *sharding.Sharding, bizID, appID uint32) (table.AppMode, error) {

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.AppColumns.NamedExpr(), " FROM ", table.AppTable.Name(),
		" WHERE id = ", strconv.Itoa(int(appID)), " AND biz_id = ", strconv.Itoa(int(bizID)))
	sql := filter.SqlJoint(sqlSentence)
	one := new(table.App)
	err := orm.Do(sd.MustSharding(bizID)).Get(kit.Ctx, one, sql)
	if err != nil {
		return "", errf.New(errf.DBOpFailed, fmt.Sprintf("get app mode failed, err: %v", err))
	}

	if err := one.Spec.Mode.Validate(); err != nil {
		return "", errf.New(errf.InvalidParameter, err.Error())
	}

	return one.Spec.Mode, nil
}

func (dao *appDao) archiveApp(kit *kit.Kit, txn *sqlx.Tx, app *table.App) error {

	id, err := ap.idGen.One(kit, table.ArchivedAppTable)
	if err != nil {
		return err
	}

	archivedApp := &table.ArchivedApp{
		ID:        id,
		AppID:     app.ID,
		BizID:     app.BizID,
		CreatedAt: time.Now(),
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", table.ArchivedAppTable.Name(),
		" (", table.ArchivedAppColumns.ColumnExpr(), ") ", "VALUES(", table.ArchivedAppColumns.ColonNameExpr(), ")")
	sql := filter.SqlJoint(sqlSentence)
	err = ap.orm.Txn(txn).Insert(kit.Ctx, sql, archivedApp)
	if err != nil {
		return fmt.Errorf("archived delete app failed, err: %v", err)
	}

	return nil
}

// ListAppMetaForCache list app's basic meta info.
func (dao *appDao) ListAppMetaForCache(kt *kit.Kit, bizID uint32, appIDs []uint32) (
	map[uint32]*types.AppCacheMeta, error) {

	if bizID <= 0 || len(appIDs) == 0 {
		return nil, errors.New("invalid biz id or app id list")
	}

	appIDList := tools.JoinUint32(appIDs, ",")
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT id, name AS 'spec.name', config_type AS 'spec.config_type', mode AS 'spec.mode', reload_type AS ",
		"'spec.reload.reload_type', reload_file_path AS 'spec.reload.file_reload_spec.reload_file_path' ",
		"FROM ", table.AppTable.Name(), " WHERE id IN (", appIDList, ") AND biz_id = ", strconv.Itoa(int(bizID)))
	sql := filter.SqlJoint(sqlSentence)
	appList := make([]*table.App, 0)
	if err := ap.orm.Do(ap.sd.MustSharding(bizID)).Select(kt.Ctx, &appList, sql); err != nil {
		return nil, fmt.Errorf("query db with app failed, err: %v", err)
	}

	meta := make(map[uint32]*types.AppCacheMeta)
	for _, one := range appList {
		meta[one.ID] = &types.AppCacheMeta{
			Name:       one.Spec.Name,
			ConfigType: one.Spec.ConfigType,
			Mode:       one.Spec.Mode,
			Reload:     one.Spec.Reload,
		}
	}

	return meta, nil
}
